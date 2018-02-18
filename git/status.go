// Copyright 2018 Josh Komoroske. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE.txt file.

package git

type Status struct {
	Branch  string
	Files   []string
	Message string
	Tags    []string
}

func Report(reporter Reporter) Status {
	return Status{
		reporter.Branch(),
		reporter.Files(),
		reporter.Message(),
		reporter.Tags(),
	}
}
