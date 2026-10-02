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

package reader_test

import (
	"strings"
	"testing"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/scanner"
)

// Issue #3: a transaction section with no active account must return an
// error instead of panicking.
func TestReadTransactionsWithoutAccount(t *testing.T) {
	input := "!Type:Bank\nD1/ 2'16\nT10.00\n^\n"
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("scanner: unexpected error: %v", err)
	}

	var r *reader.Reader
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Read panicked: %v", p)
			}
		}()
		r, err = reader.Read(sc)
	}()
	if err == nil {
		t.Fatalf("Read: expected error, got nil (reader %+v)", r)
	}
	if !strings.Contains(err.Error(), `"!Type:Bank"`) || !strings.Contains(err.Error(), "before any !Account header") {
		t.Errorf("Read: error %q does not mention the transaction header", err)
	}
	// Columns are 1-based on line 1 too (issue #47).
	if !strings.HasPrefix(err.Error(), "1:1:") {
		t.Errorf("Read: error %q does not start with the position 1:1:", err)
	}
}

// A memorized transaction section does not need an active account.
func TestReadMemorizedWithoutAccount(t *testing.T) {
	input := "!Type:Memorized\nKC\nT-10.00\nPGrocer\n^\n"
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("scanner: unexpected error: %v", err)
	}
	r, err := reader.Read(sc)
	if err != nil {
		t.Fatalf("Read: unexpected error: %v", err)
	}
	if len(r.Memorized) != 1 {
		t.Fatalf("Read: expected 1 memorized transaction, got %d", len(r.Memorized))
	}
	if got := r.Memorized[0].Payee; got != "Grocer" {
		t.Errorf("Read: memorized payee: expected %q, got %q", "Grocer", got)
	}
}
