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

package csv_test

import (
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/category"
	"github.com/maloquacious/qif/writer/csv"
	"testing"
)

// TestTranslateNilSections verifies that Translate does not dereference
// sections that the reader left nil because the file did not contain them.
func TestTranslateNilSections(t *testing.T) {
	for _, tc := range []struct {
		name string
		r    *reader.Reader
	}{
		{"empty", &reader.Reader{}},
		{"categories only", &reader.Reader{
			Categories: &category.Section{Records: []*category.Record{{Name: "Groceries"}}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("Translate panicked: %v", p)
				}
			}()
			if _, err := csv.Translate(tc.r); err != nil {
				t.Fatalf("Translate returned error: %v", err)
			}
		})
	}
}
