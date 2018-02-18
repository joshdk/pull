// Copyright 2018 Josh Komoroske. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE.txt file.

package git

import (
	"io"
	"sort"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Reporter interface {
	Branch() string
	Files() []string
	Message() string
	Tags() []string
}

func New(path string) (*Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	return &Repository{r, ref}, nil
}

type Repository struct {
	repo *git.Repository
	head *plumbing.Reference
}

func (repo *Repository) Branch() string {
	if repo.head.Name().IsBranch() {
		return repo.head.Name().Short()
	}
	return ""
}

func (repo *Repository) Files() []string {
	// Get current commit state
	headObject, err := repo.repo.CommitObject(repo.head.Hash())
	if err != nil {
		panic("failed to get HEAD commit")
	}
	headTree, err := headObject.Tree()
	if err != nil {
		panic("failed to get HEAD tree")
	}

	// Get previous (parent) commit state
	parentObject, err := headObject.Parents().Next()

	if err != nil {
		if err != io.EOF {
			panic("unknown error")
		}
		return filesTouched(nil, headTree)
	}

	parentTree, err := parentObject.Tree()
	if err != nil {
		panic("failed to get HEAD parent tree")
	}
	return filesTouched(parentTree, headTree)
}

func (repo *Repository) Message() string {
	commit, err := repo.repo.CommitObject(repo.head.Hash())
	if err != nil {
		panic("failed to get commit")
	}

	return commit.Message
}

func (repo *Repository) Tags() []string {
	var tags []string

	// Iterator to all tag references
	iter, err := repo.repo.Tags()
	if err != nil {
		panic("failed to get tags")
	}

	// Iterate over all tag references
	err = iter.ForEach(func(reference *plumbing.Reference) error {
		// Check to see if the given tag reference also points to HEAD
		if reference.Hash() == repo.head.Hash() {
			// Save reference for later
			tags = append(tags, reference.Name().Short())
		}
		return nil
	})
	if err != nil {
		panic("failed to iterate over references")
	}

	sort.Strings(tags)

	return tags
}

func filesTouched(parent *object.Tree, child *object.Tree) []string {
	// There is no parent commit (possibly first commit on a branch?) so diff against nothing.
	var (
		changes object.Changes
		err     error
		files   []string
	)

	if parent == nil {
		changes, err = child.Diff(nil)
	} else {
		changes, err = parent.Diff(child)
	}

	if err != nil {
		panic("could not diff")
	}

	fileset := make(map[string]struct{})

	for _, change := range changes {
		from, to, err := change.Files()
		if err != nil {
			panic("could not diff changes")
		}

		if from != nil {
			fileset[from.Name] = struct{}{}
		}

		if to != nil {
			fileset[to.Name] = struct{}{}
		}
	}

	for file := range fileset {
		files = append(files, file)
	}

	sort.Strings(files)

	return files
}
