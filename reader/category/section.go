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

// Package category implements a simple parser for category data.
// It returns the first error found with the data.
package category

import (
	"fmt"
	"github.com/maloquacious/qif/scanner"
)

type Section struct {
	Line    int
	Col     int
	Records []*Record
}

func ReadSection(sc scanner.Scanner) (*Section, scanner.Scanner, error) {
	saved, sname, section := sc, "categories", Section{Line: sc.Line, Col: sc.Col}

	lit, bb := sc.Literal("!Type:Cat")
	if lit == nil {
		return nil, saved, nil
	}
	sc = bb

	// read the section detail
	var err error
	for {
		var record *Record
		record, sc, err = ReadRecord(sc)
		if err != nil {
			return nil, sc, fmt.Errorf("%d: %s: %w", section.Line, sname, err)
		} else if record == nil {
			break
		}
		section.Records = append(section.Records, record)
	}

	// read the end of section marker
	eos, bb := sc.EndOfSection()
	if eos == nil {
		return nil, saved, fmt.Errorf("%d: %s: %d:%d: unexpected input", section.Line, sname, sc.Line, sc.Col)
	}
	sc = bb

	return &section, sc, nil
}
