// Copyright (c) 2025 Steve Taranto staranto@gmail.com.
// SPDX-License-Identifier: Apache-2.0

package csv

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/go-tfe"
)

func Finder(versions []*tfe.StateVersion, specs ...string) ([]*tfe.StateVersion, error) {
	var result = []*tfe.StateVersion{}

	// specs is going to be zero or more (almost certainly max=2) SV specs.  A
	// spec could be -
	//   empty  - the CSV.
	//   sv-id  - the SV with that ID.
	//   CSV~1  - the -1 SV.
	//   serial - the specific serial number.
	//   url    - the SV URL to download.
	//   file   - the SV file to read.

	if len(specs) == 0 {
		specs = []string{"CSV~0"}
	}

	var index int
	for _, s := range specs {

		if strings.HasPrefix(strings.ToUpper(s), "CSV~") {
			parts := strings.Split(s, "~")
			index, _ = strconv.Atoi(parts[1])
		} else if i, err := strconv.Atoi(s); err == nil {
			if i <= 0 {
				// <= 0 means it's a relative index into the version list
				index = -i
			} else {
				// Otherwise it's a state serial number that we have to go find.
				found := false
				for j, v := range versions {
					if v.Serial == int64(i) {
						index = j
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("failed to find state version with serial %d", i)
				}
			}
		} else if _, err := os.Stat(s); err == nil && !os.IsNotExist(err) {
			sv := tfe.StateVersion{
				ID:              s,
				Serial:          0,
				JSONDownloadURL: s,
			}
			result = append(result, &sv)
			continue

		} else {
			// It's an ID, go find it.  This is a starts with search.  If a full ID
			// has been specified, it will be the same as an equals.  If a partial ID,
			// then it will return the first (ie. newest) index found
			for j, v := range versions {
				if strings.HasPrefix(v.ID, s) {
					index = j
					break
				}
			}
		}

		if index > len(versions)-1 {
			return nil, fmt.Errorf("index %d out of range for versions of length %d", index, len(versions))
		}

		result = append(result, versions[index])
	}

	return result, nil
}
