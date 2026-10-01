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

package category_test

import (
	"github.com/maloquacious/qif/reader/category"
	"github.com/maloquacious/qif/scanner"
	"testing"
)

func TestReadRecordTaxRelated(t *testing.T) {
	for _, tc := range []struct {
		input    string
		expected bool
	}{
		{"NFood\nT\n^\n", true},
		{"NFood\n^\n", false},
	} {
		sc, err := scanner.New([]byte(tc.input))
		if err != nil {
			t.Fatalf("input of %q: scanner: unexpected error %v\n", tc.input, err)
		}
		record, _, err := category.ReadRecord(sc)
		if err != nil {
			t.Fatalf("input of %q: unexpected error %v\n", tc.input, err)
		} else if record == nil {
			t.Fatalf("input of %q: expected record, got nil\n", tc.input)
		}
		if record.IsTaxRelated != tc.expected {
			t.Errorf("input of %q yields IsTaxRelated %v: expected value is %v\n", tc.input, record.IsTaxRelated, tc.expected)
		}
	}
}
