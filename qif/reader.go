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

package qif

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// readline reads the first line from the buffer.
// It returns either a slice or an error, never both.
// On end-of-input, it returns io.EOF.
// The line is trimmed of trailing spaces (including
// line-feeds and carriage-returns).
func readline(buf []byte) (bytesConsumed int, line []byte, err error) {
	if len(buf) == 0 {
		return 0, nil, io.EOF
	}

	pos, col, end := 0, 0, 0
	for pos < len(buf) {
		col++
		r, w := utf8.DecodeRune(buf[pos:])
		pos += w
		if r == utf8.RuneError {
			return w, nil, fmt.Errorf("%d: invalid utf-8 character", col)
		} else if r == '\n' {
			break
		}

		if !unicode.IsSpace(r) {
			// this avoids copying trailing spaces
			end = pos
		}
	}

	return pos, buf[:end], nil
}
