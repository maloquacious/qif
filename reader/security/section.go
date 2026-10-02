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

// Package security implements a simple parser for security data.
// It returns the first error found with the data.
package security

import (
	"github.com/maloquacious/qif/reader/internal/section"
	"github.com/maloquacious/qif/scanner"
)

// Section holds the records of one section. Line and Col are where its
// header appears.
type Section struct {
	Line    int       `json:"-"`
	Col     int       `json:"-"`
	Records []*Record `json:"records,omitempty"`
}

// ReadSection reads a securities section. It returns a nil section and the
// unchanged scanner if the input doesn't start with the "!Type:Security" header.
func ReadSection(sc scanner.Scanner) (*Section, scanner.Scanner, error) {
	s, sc, err := section.Read(sc, "!Type:Security", "securities", ReadRecord)
	if s == nil {
		return nil, sc, err
	}
	return &Section{Line: s.Line, Col: s.Col, Records: s.Records}, sc, nil
}
