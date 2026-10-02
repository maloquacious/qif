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

package section_test

import (
	"strings"
	"testing"

	"github.com/maloquacious/qif/reader/internal/section"
	"github.com/maloquacious/qif/scanner"
)

// readLine is a record reader that accepts one "V" field per record.
func readLine(sc scanner.Scanner) (*string, scanner.Scanner, error) {
	v, bb := sc.Field("V")
	if v == nil {
		return nil, sc, nil
	}
	eor, bb := bb.EndOfRecord()
	if eor == nil {
		return nil, bb, errMissingTerminator
	}
	s := string(v)
	return &s, bb, nil
}

type sentinel string

func (e sentinel) Error() string { return string(e) }

const errMissingTerminator = sentinel("missing record terminator")

// TestRead is a regression test for issue #21: the generic section reader
// matches the header, reads records to the end of the section and reports
// errors with the section's line.
func TestRead(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		want        []string
		wantErr     string
		wantNil     bool
		wantRest    string
	}{
		{name: "no match", input: "!Type:Other\nVa\n^\n", wantNil: true, wantRest: "!Type:Other\nVa\n^\n"},
		{name: "records", input: "!Type:Test\nVa\n^\nVb\n^\n!Type:Next\n", want: []string{"a", "b"}, wantRest: "!Type:Next\n"},
		{name: "empty", input: "!Type:Test\n", want: nil, wantRest: ""},
		{name: "record error", input: "!Type:Test\nVa\nX\n", wantErr: "1: tests: missing record terminator"},
		{name: "unexpected input", input: "!Type:Test\nVa\n^\nX\n", wantErr: "1: tests: 4:1: unexpected input"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc, err := scanner.New([]byte(tc.input))
			if err != nil {
				t.Fatalf("scanner.New: %v", err)
			}
			s, rest, err := section.Read(sc, "!Type:Test", "tests", readLine)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("want error %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if s != nil {
					t.Errorf("want nil section, got %+v", s)
				}
			} else {
				if s == nil {
					t.Fatal("want a section, got nil")
				}
				var got []string
				for _, r := range s.Records {
					got = append(got, *r)
				}
				if strings.Join(got, ",") != strings.Join(tc.want, ",") {
					t.Errorf("records: want %v, got %v", tc.want, got)
				}
				if s.Line != 1 || s.Col != 1 {
					t.Errorf("position: want 1:1, got %d:%d", s.Line, s.Col)
				}
			}
			if string(rest.Buffer) != tc.wantRest {
				t.Errorf("rest: want %q, got %q", tc.wantRest, rest.Buffer)
			}
		})
	}
}
