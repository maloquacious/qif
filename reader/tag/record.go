/*
 * qif - a package to convert QIF data
 *
 * Copyright (c) 2021 Michael D Henderson
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package tag

import (
	"fmt"
	"github.com/mdhender/qif/scanner"
)

type Record struct {
	Line        int    `json:"-"`
	Col         int    `json:"-"`
	Description string `json:"descr,omitempty"`
	Name        string `json:"name"`
}

func ReadRecord(sc scanner.Scanner) (*Record, scanner.Scanner, error) {
	saved, sname, record := sc, "tag", Record{Line: sc.Line, Col: sc.Col}

	var found bool
	var descr, name []byte
	for {
		if descr == nil {
			if descr, sc = sc.Field("D"); descr != nil {
				found, record.Description = true, string(descr)
				continue
			}
		}
		if name == nil {
			if name, sc = sc.Field("N"); name != nil {
				found, record.Name = true, string(name)
				continue
			}
		}

		break
	}

	if !found { // no fields found
		return nil, saved, nil
	}

	// check for required fields
	if name == nil {
		return nil, saved, fmt.Errorf("%d: %s: missing field %q", record.Line, sname, "name")
	}

	eor, bb := sc.EndOfRecord()
	if eor == nil {
		return nil, saved, fmt.Errorf("%d: %s: missing record terminator", sc.Line, sname)
	}
	sc = bb

	return &record, sc, nil
}
