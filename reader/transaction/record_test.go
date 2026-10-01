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

package transaction_test

import (
	"testing"

	"github.com/maloquacious/qif/reader/transaction"
	"github.com/maloquacious/qif/scanner"
)

// TestReadRecordTransferWithClass is a regression test for issue #14.
func TestReadRecordTransferWithClass(t *testing.T) {
	sc, err := scanner.New([]byte("D1/ 2'24\nT-10.00\nL[Checking]/Business\n^\n"))
	if err != nil {
		t.Fatalf("scanner.New: unexpected error: %v", err)
	}
	record, _, err := transaction.ReadRecord(sc, "Chk", "Bank")
	if err != nil {
		t.Fatalf("ReadRecord: unexpected error: %v", err)
	}
	if record == nil {
		t.Fatalf("ReadRecord: want record, got nil")
	}
	if want := "Checking"; record.ToAccount != want {
		t.Errorf("ToAccount: want %q, got %q", want, record.ToAccount)
	}
	if want := ""; record.Category != want {
		t.Errorf("Category: want %q, got %q", want, record.Category)
	}
	if want := "Business"; record.Class != want {
		t.Errorf("Class: want %q, got %q", want, record.Class)
	}
}

// TestReadRecordSplitTransferWithClass is a regression test for issue #14.
func TestReadRecordSplitTransferWithClass(t *testing.T) {
	sc, err := scanner.New([]byte("D1/ 2'24\nT-10.00\nS[Savings]/Biz\n$-10.00\n^\n"))
	if err != nil {
		t.Fatalf("scanner.New: unexpected error: %v", err)
	}
	record, _, err := transaction.ReadRecord(sc, "Chk", "Bank")
	if err != nil {
		t.Fatalf("ReadRecord: unexpected error: %v", err)
	}
	if record == nil || len(record.Split) != 1 {
		t.Fatalf("ReadRecord: want one split, got %+v", record)
	}
	split := record.Split[0]
	if want := "Savings"; split.Account != want {
		t.Errorf("Account: want %q, got %q", want, split.Account)
	}
	if want := ""; split.Category != want {
		t.Errorf("Category: want %q, got %q", want, split.Category)
	}
	if want := "Biz"; split.Class != want {
		t.Errorf("Class: want %q, got %q", want, split.Class)
	}
	if want := "-10.00"; split.Amount != want {
		t.Errorf("Amount: want %q, got %q", want, split.Amount)
	}
}
