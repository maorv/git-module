// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"container/list"
	"errors"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Repository represents a Git repository.
type Repository struct {
	Path string

	commitCache *objectCache
	tagCache    *objectCache
}

const _PRETTY_LOG_FORMAT = `--pretty=format:%H`

func (repo *Repository) parsePrettyFormatLogToList(logs []byte) (*list.List, error) {
	l := list.New()
	if len(logs) == 0 {
		return l, nil
	}

	parts := bytes.Split(logs, []byte{'\n'})

	for _, commitId := range parts {
		commit, err := repo.GetCommit(string(commitId))
		if err != nil {
			return nil, err
		}
		l.PushBack(commit)
	}

	return l, nil
}

// InitRepository initializes a new Git repository.
func InitRepository(repoPath string, bare bool) error {
	os.MkdirAll(repoPath, os.ModePerm)

	cmd := NewCommand("init")
	if bare {
		cmd.AddArguments("--bare")
	}
	_, err := cmd.RunInDir(repoPath)
	return err
}

// OpenRepository opens the repository at the given path.
func OpenRepository(repoPath string) (*Repository, error) {
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	} else if !isDir(repoPath) {
		return nil, errors.New("no such file or directory")
	}

	return &Repository{
		Path:        repoPath,
		commitCache: newObjectCache(),
		tagCache:    newObjectCache(),
	}, nil
}

type CloneRepoOptions struct {
	Mirror  bool
	Bare    bool
	Quiet   bool
	Timeout time.Duration
}

// Clone clones original repository to target path.
func Clone(from, to string, opts CloneRepoOptions) (err error) {
	toDir := path.Dir(to)
	if err = os.MkdirAll(toDir, os.ModePerm); err != nil {
		return err
	}

	cmd := NewCommand("clone")
	if opts.Mirror {
		cmd.AddArguments("--mirror")
	}
	if opts.Bare {
		cmd.AddArguments("--bare")
	}
	if opts.Quiet {
		cmd.AddArguments("--quiet")
	}
	cmd.AddArguments(from, to)

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}

	_, err = cmd.RunTimeout(opts.Timeout)
	return err
}

type PullRemoteOptions struct {
	All     bool
	Timeout time.Duration
}

// Pull pulls changes from remotes.
func Pull(repoPath string, opts PullRemoteOptions) error {
	cmd := NewCommand("pull")
	if opts.All {
		cmd.AddArguments("--all")
	}

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}

	_, err := cmd.RunInDirTimeout(opts.Timeout, repoPath)
	return err
}

type FetchRemoteOptions struct {
	Prune   bool
	Timeout time.Duration
}

// Fetch fetch changes from remotes.
func Fetch(repoPath string, opts FetchRemoteOptions) error {
	cmd := NewCommand("fetch")
	if opts.Prune {
		cmd.AddArguments("--prune")
	}

	if opts.Timeout <= 0 {
		opts.Timeout = -1
	}

	_, err := cmd.RunInDirTimeout(opts.Timeout, repoPath)
	return err
}

type RebaseOptions struct {
	Branch string
}

// Rebase rebase local commits on top of remote branch.
func Rebase(repoPath string, opts RebaseOptions) error {
	cmd := NewCommand("rebase")

	if opts.Branch != "" {
		cmd.AddArguments(opts.Branch)
	}

	_, err := cmd.RunInDir(repoPath)
	return err
}

type PushOptions struct {
	Force bool
}

// Push pushs local commits to given remote branch.
func Push(repoPath, remote, branch string, opts PushOptions) error {
	cmd := NewCommand("push")

	if opts.Force {
		cmd.AddArguments("--force")
	}

	cmd.AddArguments(remote)
	cmd.AddArguments(branch)

	_, err := cmd.RunInDir(repoPath)
	return err
}

// ResetHEAD resets HEAD to given revision or head of branch.
func ResetHEAD(repoPath string, hard bool, revision string) error {
	cmd := NewCommand("reset")
	if hard {
		cmd.AddArguments("--hard")
	}
	_, err := cmd.AddArguments(revision).RunInDir(repoPath)
	return err
}

func Checkout(repoPath, version string) error {
	_, err := NewCommand("checkout", version).RunInDir(repoPath)
	return err
}
