package timecard

////////////////////////////////////////////////////////////////////////////////a

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/sabhiram/timecard/git"
)

////////////////////////////////////////////////////////////////////////////////

const (
	v1HeaderSize = 9 // bytes
)

////////////////////////////////////////////////////////////////////////////////

type Header struct {
	Size    byte   // Size of the timecard header
	Version uint32 // 32 bit hex version 8:Major 8:Minor 16:Patch
	Count   int32  // 32 bit count of number of timecard entries
}

func (h *Header) Unmarshal(data []byte) error {
	decoded, err := hex.DecodeString(string(data))
	if err != nil {
		return err
	}

	if len(decoded) < v1HeaderSize {
		return errors.New("insufficient data, cannot form header")
	}
	return binary.Read(bytes.NewBuffer(decoded), binary.LittleEndian, h)
}

func (h *Header) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1+4+4))
	binary.Write(buf, binary.LittleEndian, h.Size)
	binary.Write(buf, binary.LittleEndian, h.Version)
	binary.Write(buf, binary.LittleEndian, h.Count)
	return []byte(hex.EncodeToString(buf.Bytes())), nil
}

////////////////////////////////////////////////////////////////////////////////

const (
	cStateUnknown = iota // Invalid / init state
	cStatePending = iota // Waiting for timecard end
	cStatePartial = iota // Got timecard end, waiting for hash
	cStateHashed  = iota // Hash has been recorded
)

// Entry represents a single entry in a timecard.  NOTE: We are currently
// ignoring checkpoints in the entries.
type Entry struct {
	Start int64 // Seconds since epoch
	End   int64 // Seconds since epoch
	Hash  string
	State int
}

// Unmarshal takes a single line of timecard input and attempts to convert
// it into a valid timecard entry.
func (e *Entry) Unmarshal(data []byte) error {
	line := string(data)
	if len(line) == 0 {
		return errors.New("cannot make entry from empty line")
	}

	e.Start = 0
	e.End = 0
	e.Hash = ""
	e.State = cStateUnknown

	items := strings.Split(line, ",")
	switch len(items) {
	case 1:
		return errors.New("invalid timecard line detected")
	case 2:
		start, err := strconv.ParseInt(items[0], 10, 64)
		if err != nil {
			return errors.New("unable to parse start time")
		}
		e.Start = start
		e.State = cStatePending
		return nil
	case 3:
		start, err := strconv.ParseInt(items[0], 10, 64)
		if err != nil {
			return errors.New("unable to parse start time")
		}
		end, err := strconv.ParseInt(items[1], 10, 64)
		if err != nil {
			return errors.New("unable to parse end time")
		}
		e.Start = start
		e.End = end
		e.Hash = items[2]
		e.State = cStateHashed
		if len(e.Hash) == 0 {
			e.State = cStatePartial
		}
		return nil
	}
	return errors.New("invalid timecard line detected")
}

func (e *Entry) Marshal() ([]byte, error) {
	if e == nil || e.Start == 0 {
		return nil, errors.New("invalid timecard entry")
	}
	if e.End == 0 {
		return []byte(fmt.Sprintf("%d,", e.Start)), nil
	}
	return []byte(fmt.Sprintf("%d,%d,%s", e.Start, e.End, e.Hash)), nil
}

////////////////////////////////////////////////////////////////////////////////

// Timecard is the in-memory representation of the .timecard file.  The version
// specified in the structure can be used as a hint to migrate the header block
// should the below structure ever have to change.
type Timecard struct {
	Path    string   // path to file
	Header  *Header  // Timecard's header
	Entries []*Entry // Slice of timecard entries
	repo    *git.Git
}

func Init(r *git.Git, fp string) (*Timecard, error) {
	tc := &Timecard{
		Path: fp,
		Header: &Header{
			Size:    v1HeaderSize, // Fixed v1 header size
			Version: 0x00000001,   // 0.0.0001
			Count:   0,            // no entries as of now
		},
		Entries: []*Entry{},
		repo:    r,
	}
	return tc, tc.Flush()
}

func Load(r *git.Git, fp string) (*Timecard, error) {
	bs, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	tc := &Timecard{
		Path:    fp,
		Header:  &Header{},
		Entries: []*Entry{},
		repo:    r,
	}
	return tc, tc.Unmarshal(bs)
}

// Unmarshal converts a file blob into a timecard instance `tc`.
func (tc *Timecard) Unmarshal(blob []byte) error {
	lines := strings.Split(string(blob), "\n")
	if len(lines) == 0 {
		return errors.New("empty .timecard file cannot be read")
	}

	hdr, lines := []byte(lines[0]), lines[1:]
	if err := tc.Header.Unmarshal(hdr); err != nil {
		return err
	}

	for _, line := range lines {
		e := &Entry{}
		if err := e.Unmarshal([]byte(line)); err == nil {
			tc.Entries = append(tc.Entries, e)
		}
	}
	return nil
}

// Marshal converts the timecard into a string.
func (tc *Timecard) Marshal() ([]byte, error) {
	hdr, err := tc.Header.Marshal()
	if err != nil {
		return nil, err
	}

	result := []string{string(hdr)}
	for _, entry := range tc.Entries {
		bs, err := entry.Marshal()
		if err != nil {
			log.Printf("Warning: Bad line: %s. Ignoring.\n", string(bs))
		} else {
			result = append(result, string(bs))
		}
	}
	return []byte(strings.Join(result, "\n")), nil
}

// Flush updates the timecard instance `tc` to it's specified path.
func (tc *Timecard) Flush() error {
	contents, err := tc.Marshal()
	if err != nil {
		return err
	}
	contents = append(contents, '\n')
	if err := ioutil.WriteFile(tc.Path, contents, 0755); err != nil {
		return err
	}

	return nil
}

// Start starts or re-starts the current entry. This includes figuring out the
// current commit hash so it can be attributed to the last commit.
func (tc *Timecard) Start() error {
	if int(tc.Header.Count) != len(tc.Entries) {
		panic("header count and entry length mismatch")
	}

	appendNewEntryFn := func(tc *Timecard, t int64) error {
		tc.Header.Count += 1
		tc.Entries = append(tc.Entries, &Entry{
			Start: t,
			State: cStatePending,
		})
		return tc.Flush()
	}

	// If we have no entries, we make a new one with just a start time.
	if tc.Header.Count == 0 {
		return appendNewEntryFn(tc, time.Now().Unix())
	}

	// Grab the last entry, if it is complete - make a new one.  If it is
	// pending - update the current one's start time.
	lastIdx := tc.Header.Count - 1
	switch tc.Entries[lastIdx].State {
	case cStatePending:
		// Pending entries should just be updated with a new start time.
		tc.Entries[lastIdx].Start = time.Now().Unix()
		return tc.Flush()
	case cStatePartial:
		// Latest entry is partial, figure out the right commit hash for it
		// and make a new entry.
		headHash, err := tc.repo.GetCurrentHash()
		if err != nil {
			return err
		}
		tc.Entries[lastIdx].Hash = headHash
		tc.Entries[lastIdx].State = cStateHashed
		return appendNewEntryFn(tc, time.Now().Unix())
	case cStateHashed:
		// Latest entry is recorded, make a new entry.
		return appendNewEntryFn(tc, time.Now().Unix())
	}
	return nil
}

// End closes a timecard entry.
func (tc *Timecard) End() error {
	if int(tc.Header.Count) != len(tc.Entries) {
		panic("header count and entry length mismatch")
	}

	// If we have no entries, throw an error.
	if tc.Header.Count == 0 {
		return errors.New("mismatched \"timecard end\" without \"timecard start\"")
	}

	lastIdx := tc.Header.Count - 1
	log.Printf("Last entry: %#v\n", tc.Entries[lastIdx])
	switch tc.Entries[lastIdx].State {
	case cStatePending:
		// Pending entries get promoted to partial
		tc.Entries[lastIdx].End = time.Now().Unix()
		tc.Entries[lastIdx].State = cStatePartial
		return tc.Flush()
	case cStatePartial:
		return errors.New("timecard entry already closed")
	}
	return errors.New("mismatched \"timecard end\" without \"timecard start\"")
}

////////////////////////////////////////////////////////////////////////////////
