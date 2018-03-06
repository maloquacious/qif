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
	"fmt"
	"unicode"
	"unicode/utf8"
)

// readLines returns all the lines in a buffer as an
// array of slices. The lines are trimmed of trailing
// spaces (including new-lines and line-feeds).
func readLines(buf []byte) (lines [][]byte, err error) {
	pos, line := 0, 0
	for pos < len(buf) {
		line++

		start, end := pos, pos
		for pos < len(buf) {
			r, w := utf8.DecodeRune(buf[pos:])
			pos += w

			if r == utf8.RuneError {
				return nil, fmt.Errorf("%d: %s", line, "invalid utf-8 character")
			}

			if r == '\n' {
				break
			}

			if !unicode.IsSpace(r) {
				// this avoids copying trailing spaces
				end = pos
			}
		}

		lines = append(lines, buf[start:end])
	}

	return lines, err
}
