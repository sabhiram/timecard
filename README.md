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
Initialized new timecard for <gituser> in /current/path/.timecard
``` 

## Getting cute with git-hooks:

## Installation

