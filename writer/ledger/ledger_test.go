/*
 * qif - a package to convert QIF data
 *
 * Copyright (c) 2026 Michael D Henderson
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

package ledger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/transaction"
)

// TestWriteCountsSkipped verifies that entries with no amounts are counted
// as skipped and are not written.
func TestWriteCountsSkipped(t *testing.T) {
	l := &LEDGER{Entries: []*Entry{
		{Line: 1, Date: "2026/01/01", Payee: "Kept One", Account: "Checking",
			Lines: Lines{{Line: 2, Category: "Groceries", Amount: "10.00"}}},
		{Line: 3, Date: "2026/01/02", Payee: "Zero Entry", Account: "Checking", IsZero: true,
			Lines: Lines{{Line: 4, Category: "Groceries", Amount: "0.00", IsZero: true}}},
		{Line: 5, Date: "2026/01/03", Payee: "Kept Two", Account: "Checking",
			Lines: Lines{{Line: 6, Category: "Dining", Amount: "20.00"}}},
	}}

	var buf bytes.Buffer
	skipped, written, err := l.write(&buf)
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}
	if skipped != 1 {
		t.Errorf("skipped: got %d, want 1", skipped)
	}
	if written != 2 {
		t.Errorf("written: got %d, want 2", written)
	}

	out := buf.String()
	for _, want := range []string{"Kept One", "Kept Two"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Zero Entry") {
		t.Errorf("output contains skipped entry:\n%s", out)
	}
}

// openingBalance returns a reader holding a single-line opening balance
// transaction for an account of the given type.
func openingBalance(accountType string) *reader.Reader {
	return &reader.Reader{Transactions: []*transaction.Record{{
		Line: 7, Account: "Foo", Type: accountType, Date: "2016/01/02",
		Payee: "Opening Balance", AmountTCode: "100.00", ToAccount: "Foo",
	}}}
}

// TestTranslateUnknownAccountType verifies that an opening balance in an
// account of unknown type is returned as an error instead of a panic (#12).
func TestTranslateUnknownAccountType(t *testing.T) {
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Translate panicked: %v", p)
			}
		}()
		_, err = Translate(openingBalance("Mutual"))
	}()
	if err == nil {
		t.Fatal("Translate: expected error, got nil")
	}
	if want := `7: account "Foo": unknown account type "Mutual"`; err.Error() != want {
		t.Errorf("Translate: want error %q, got %q", want, err)
	}
}

// TestTranslateInvestmentOpeningBalance verifies that a single-line opening
// balance in an investment account is not flipped, like the asset types (#12).
func TestTranslateInvestmentOpeningBalance(t *testing.T) {
	for _, typ := range []string{"Invst", "Port", "401(k)/403(b)"} {
		t.Run(typ, func(t *testing.T) {
			var l *LEDGER
			var err error
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("Translate panicked: %v", p)
					}
				}()
				l, err = Translate(openingBalance(typ))
			}()
			if err != nil {
				t.Fatalf("Translate returned error: %v", err)
			}
			if len(l.Entries) != 1 || len(l.Entries[0].Lines) != 1 {
				t.Fatalf("Translate: want 1 entry with 1 line, got %+v", l.Entries)
			}
			if got := l.Entries[0].Lines[0].Amount; got != "100.00" {
				t.Errorf("Translate: amount: want %q, got %q", "100.00", got)
			}
		})
	}
}
