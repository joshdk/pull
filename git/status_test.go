// Copyright 2018 Josh Komoroske. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE.txt file.

package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {

	tests := []struct {
		title    string
		commands [][]string
		err      string
		branch   string
		tags     []string
		files    []string
		message  string
	}{
		{
			title: "empty",
			err:   "reference not found",
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit"},
			},
			branch:  "master",
			message: "empty commit",
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit 1"},
				{"git", "commit", "--allow-empty", "-m", "empty commit 2"},
				{"git", "commit", "--allow-empty", "-m", "empty commit 3"},
			},
			branch:  "master",
			message: "empty commit 3",
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit"},
				{"git", "checkout", "-b", "develop"},
			},
			branch:  "develop",
			message: "empty commit",
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit"},
				{"git", "checkout", "--orphan", "develop"},
			},
			err: "reference not found",
		},
		{
			title: "empty",
			commands: [][]string{
				{"touch", "file.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit"},
			},
			branch:  "master",
			message: "test commit",
			files:   []string{"file.txt"},
		},
		{
			title: "empty",
			commands: [][]string{
				{"touch", "file-1.txt", "file-2.txt", "file-3.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit"},
			},
			branch:  "master",
			message: "test commit",
			files:   []string{"file-1.txt", "file-2.txt", "file-3.txt"},
		},
		{
			title: "empty",
			commands: [][]string{
				{"touch", "file-1.txt", "file-2.txt", "file-3.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit"},
				{"git", "tag", "0.0.0"},
			},
			branch:  "master",
			message: "test commit",
			files:   []string{"file-1.txt", "file-2.txt", "file-3.txt"},
			tags:    []string{"0.0.0"},
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit"},
				{"git", "tag", "0.0.0"},
				{"git", "tag", "1.0.0"},
				{"git", "tag", "2.0.0"},
			},
			branch:  "master",
			message: "empty commit",
			tags:    []string{"0.0.0", "1.0.0", "2.0.0"},
		},
		{
			title: "empty",
			commands: [][]string{
				{"git", "commit", "--allow-empty", "-m", "empty commit"},
				{"git", "tag", "0.0.0"},
				{"touch", "file-1.txt", "file-2.txt", "file-3.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit 1"},
				{"git", "tag", "1.0.0"},
				{"git", "checkout", "-b", "feature/test"},
				{"touch", "file-4.txt", "file-5.txt", "file-6.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit 2"},
				{"touch", "file-7.txt", "file-8.txt", "file-9.txt"},
				{"git", "tag", "2.0.0"},
			},
			branch:  "feature/test",
			message: "test commit 2",
			files:   []string{"file-4.txt", "file-5.txt", "file-6.txt"},
			tags:    []string{"2.0.0"},
		},
		{
			title: "empty",
			commands: [][]string{
				{"touch", "file-1.txt", "file-2.txt", "file-3.txt"},
				{"git", "add", "--all"},
				{"git", "commit", "-m", "test commit 1"},
				{"git", "tag", "0.0.0"},
				{"git", "commit", "--allow-empty", "-m", "test commit 2"},
				{"git", "commit", "--allow-empty", "-m", "test commit 3"},
				{"git", "checkout", "0.0.0"},
			},
			branch:  "", //TODO: Better handle detached head state
			message: "test commit 1",
			files:   []string{"file-1.txt", "file-2.txt", "file-3.txt"},
			tags:    []string{"0.0.0"},
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("#%d - %s", index, test.title)

		t.Run(name, func(t *testing.T) {

			// Create a temporary directory, in which to construct our git repo.
			tmp, err := ioutil.TempDir("", "")
			require.Nil(t, err)

			// Cleanup the temporary directory.
			defer func() {
				if os.RemoveAll(tmp) != nil {
					panic("failed to cleanup tmp directory")
				}
			}()

			err = script(tmp, test.commands)
			require.Nil(t, err)

			reporter, err := New(tmp)
			checkErrors(t, test.err, err)
			if err != nil {
				return
			}

			// Extract status information from the git repo.
			status := Report(reporter)
			assert.Equal(t, status.Branch, test.branch)
			assert.Equal(t, status.Files, test.files)
			assert.Equal(t, status.Tags, test.tags)
			assert.Equal(t, strings.TrimSpace(status.Message), strings.TrimSpace(test.message))
		})
	}

}

func script(directory string, commands [][]string) error {
	preCommands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test"},
		{"git", "config", "user.name", "test"},
	}

	allCommands := append(preCommands, commands...)

	// Run every setup command in order to prepare the working environment.
	for _, args := range allCommands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = directory

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func checkErrors(t *testing.T, expected string, actual error) {
	t.Helper()
	switch {
	case expected == "" && actual == nil:
		return
	case expected == "" && actual != nil:
		require.Equal(t, nil, actual.Error())
	case expected != "" && actual == nil:
		require.Equal(t, expected, nil)
	case expected != "" && actual != nil:
		require.Equal(t, expected, actual.Error())
	}
}
