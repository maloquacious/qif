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
	"github.com/maloquacious/qif/reader/internal/section"
	"github.com/maloquacious/qif/scanner"
)

type Section struct {
	Line    int
	Col     int
	Records []*Record
}

// ReadSection reads a categories section. It returns a nil section and the
// unchanged scanner if the input doesn't start with the "!Type:Cat" header.
func ReadSection(sc scanner.Scanner) (*Section, scanner.Scanner, error) {
	s, sc, err := section.Read(sc, "!Type:Cat", "categories", ReadRecord)
	if s == nil {
		return nil, sc, err
	}
	return &Section{Line: s.Line, Col: s.Col, Records: s.Records}, sc, nil
}
