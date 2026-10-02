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

package reader

import (
	"strings"
	"testing"

	"github.com/maloquacious/qif/scanner"
)

// readNoPanic runs Read on the input and fails the test if Read panics.
func readNoPanic(t *testing.T, input string) (*Reader, error) {
	t.Helper()
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("scanner: %v", err)
	}
	var r *Reader
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Read panicked: %v", p)
			}
		}()
		r, err = Read(sc)
	}()
	return r, err
}

// A second section of accounts, categories or tags must return an error
// naming both sections instead of panicking (#17).
func TestReadDuplicateSections(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  string // prefix of the error
	}{
		{
			name: "accounts",
			input: "!Account\nNChk\nTBank\n^\nNVisa\nTCCard\n^\n" +
				"!Account\nNSav\nTBank\n^\nNCash\nTCash\n^\n",
			want: "8:1: duplicate account list (first at 1:1)",
		},
		{
			name: "categories",
			input: "!Type:Cat\nNFood\nE\n^\n" +
				"!Type:Cat\nNRent\nE\n^\n",
			want: "5:1: duplicate category list (first at 1:1)",
		},
		{
			name: "tags",
			input: "!Type:Tag\nNHome\n^\n" +
				"!Type:Tag\nNWork\n^\n",
			want: "4:1: duplicate tag list (first at 1:1)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := readNoPanic(t, tc.input)
			if err == nil {
				t.Fatalf("Read: expected error, got nil (reader %+v)", r)
			}
			if !strings.Contains(err.Error(), "duplicate") {
				t.Errorf("Read: error %q does not mention the duplicate", err)
			}
			if !strings.HasPrefix(err.Error(), tc.want) {
				t.Errorf("Read: want error starting with %q, got %q", tc.want, err)
			}
		})
	}
}

// A repeated section with no records is ignored, as before.
func TestReadEmptyDuplicateSections(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"accounts", "!Account\nNChk\nTBank\n^\nNVisa\nTCCard\n^\n!Account\n"},
		{"categories", "!Type:Cat\nNFood\nE\n^\n!Type:Cat\n"},
		{"tags", "!Type:Tag\nNHome\n^\n!Type:Tag\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := readNoPanic(t, tc.input); err != nil {
				t.Errorf("Read: unexpected error: %v", err)
			}
		})
	}
}
