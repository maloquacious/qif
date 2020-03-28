// Copyright © 2018 MICHAEL D HENDERSON <mdhender@mdhender.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/qif/qif/reader"
	"github.com/pkg/errors"
	"io/ioutil"
)

func main() {
	testFiles := []string{
		"quicken/mac",
		"quicken/windows",
	}

	var errs []error
	for _, val := range testFiles {
		fmt.Printf("testing %q\n", val)

		q, err := reader.ImportFile(val + ".qif")
		if err != nil {
			errs = append(errs, errors.Wrap(err, "importing "+val+".qif"))
			continue
		}
		js, err := json.MarshalIndent(q, "", "  ")
		if err != nil {
			errs = append(errs, errors.Wrap(err, "marshaling "+val+".qif"))
			continue
		}
		if err = ioutil.WriteFile(val+".json", js, 0644); err != nil {
			errs = append(errs, errors.Wrap(err, "writing "+val+".json"))
			continue
		}
		fmt.Println("created " + val + ".json")
	}

	for _, err := range errs {
		fmt.Printf("error: %+v\n", err)
	}
}
