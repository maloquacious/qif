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
	"log"

	"github.com/mdhender/qif/config"
	"github.com/mdhender/qif/qif"
)

// go : generate stringer --type=LexemeKind qif
// go : generate stringer --type=lexerState qif

func main() {
	cfg := config.New()
	if err := cfg.MergeFile("quicken/config.json"); err != nil {
		log.Fatal(err)
	}

	for key, val := range cfg.InputFiles {
		log.Printf("%05d: %s\n", key, val)

		_, err := qif.ImportFile(val)
		if err != nil {
			log.Fatal(err)
		}
	}
}
