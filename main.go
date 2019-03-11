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
	"io/ioutil"
	"log"

	"github.com/mdhender/qif/qif/reader"
)

func main() {
	testFiles := []string{
		"quicken/mac",
		"quicken/windows",
	}

	for key, val := range testFiles {
		log.Printf("%05d: %s\n", key, val)

		q, err := reader.ImportFile(val + ".qif")
		if err != nil {
			log.Fatal(err)
		}
		js, err := json.MarshalIndent(q, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err = ioutil.WriteFile(val+".json", js, 0644); err != nil {
			log.Fatal(err)
		}
	}
}
