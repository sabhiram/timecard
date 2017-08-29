# timecard

Keep track of how long your commits take

## What is it?

`timecard` is a dead simple way to keep track of how long you spend working on various commits within git projects. It intends to live alongside the `.git` directory and figures out how long you spent working on code by analyzing changes in the git tree.

Since the commit hash is a SHA hash of the file tree, each commit can only be added to the `.timecard` file on a subsequent commit. This is a not a big deal, since the current commit matches the `CURRENT` tag in the `.timecard` file.

## Install

```
go get github.com/sabhiram/timecard
```

## Usage

Setting up:

```
$ timecard init
Error: Could not find a valid git repository at /current/path. Did you "git init"?
$ git init
Initialized empty Git repository in /current/path/.git/
$ timecard init
Initialized new timecard for <gituser> in /current/path/.timecard.
``` 

## Getting cute with git-hooks:



## The `.timecard` file

The very first line of the file is special, and is used to store a binary blob that is specific to the timecard application. Each subsequent line in this file represents a single commit hash with a start time, end time and the times at which various checkpoints might have been taken.

A `.timecard` file with three commits that have been made would look like:
```
CD25691738CEF2B8EB3E5391C44D2472
start0,end0,commithash0
start1,end1,commithash1
start2,end2,CURRENT_COMMIT
```

And once `timecard start` has executed:
```
67391A6724D6E764D1021E36F55FECEB
start0,end0,commithash0
start1,end1,commithash1
start2,end2,commithash2
start3,PENDING
```

And once `git commit` has run:
```
89482E12EA363AA0EA40B3364ACD1699
start0,end0,commithash0
start1,end1,commithash1
start2,end2,commithash2
start3,end3,CURRENT_COMMIT
```


