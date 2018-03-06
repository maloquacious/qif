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

package reader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
)

// Import does
func Import(file string) (*File, error) {
	f := File{}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines, err := readLines(buf)
	if err != nil {
		return nil, err
	}
	log.Printf("%5d: %s\n", len(lines), file)

	autoSwitch, line := false, 0
	for line < len(lines) {
		switch {
		case bytes.Compare(lines[line], []byte("!Account")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				if autoSwitch {
					f.Accounts = details
				} else {
					f.Account = details
				}
			}
		case bytes.Compare(lines[line], []byte("!Clear:AutoSwitch")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			autoSwitch = false
		case bytes.Compare(lines[line], []byte("!Option:AutoSwitch")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			autoSwitch = true
		case bytes.Compare(lines[line], []byte("!Type:Bank")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.Bank = details
			}
		case bytes.Compare(lines[line], []byte("!Type:Cat")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.Category = details
			}
		case bytes.Compare(lines[line], []byte("!Type:CCard")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.CreditCard = details
			}
		case bytes.Compare(lines[line], []byte("!Type:Invst")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.Investment = details
			}
		case bytes.Compare(lines[line], []byte("!Type:Memorized")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.Memorized = details
			}
		case bytes.Compare(lines[line], []byte("!Type:Tag")) == 0:
			log.Printf("%05d: %s\n", line+1, string(lines[line]))
			line++
			linesConsumed, details := getDetails(line, lines[line:])
			line += linesConsumed
			if details != nil {
				f.Tag = details
			}
		default:
			return nil, fmt.Errorf("%d: invalid section %q", line, string(lines[line]))
		}
	}
	log.Printf("%05d: %-20s %8d\n", 0, "account", len(f.Account))
	log.Printf("%05d: %-20s %8d\n", 0, "accounts", len(f.Accounts))
	log.Printf("%05d: %-20s %8d\n", 0, "bank", len(f.Bank))
	log.Printf("%05d: %-20s %8d\n", 0, "category", len(f.Category))
	log.Printf("%05d: %-20s %8d\n", 0, "creditCard", len(f.CreditCard))
	log.Printf("%05d: %-20s %8d\n", 0, "investment", len(f.Investment))
	log.Printf("%05d: %-20s %8d\n", 0, "memorized", len(f.Memorized))
	log.Printf("%05d: %-20s %8d\n", 0, "tag", len(f.Tag))

	return &f, nil
}

func getDetails(lineNo int, lines [][]byte) (linesConsumed int, details []*Detail) {
	var detail *Detail
	for linesConsumed < len(lines) && lines[linesConsumed][0] != '!' {
		if lines[linesConsumed][0] == '^' {
			if detail != nil {
				details = append(details, detail)
			}
			detail = nil
		} else {
			if detail == nil {
				detail = &Detail{}
			}
			detail.Lines = append(detail.Lines, &Line{LineNo: linesConsumed + lineNo, Value: lines[linesConsumed]})
		}
		linesConsumed++
	}
	if detail != nil {
		details = append(details, detail)
	}
	return linesConsumed, details
}
