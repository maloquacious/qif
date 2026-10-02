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
	"io"
	"log/slog"
	"math"
	"strconv"
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
		_, err = Translate(openingBalance("Mutual"), nil)
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
				l, err = Translate(openingBalance(typ), nil)
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

// TestEntryWriteAccountNameConsistent verifies that a posting line and the
// balancing posting render the same QIF account name as the same Ledger
// account, without Go-style quoting.
func TestEntryWriteAccountNameConsistent(t *testing.T) {
	e := &Entry{Line: 1, Date: "2026/01/01", Payee: "Transfer", Account: "My  Savings",
		Lines: Lines{{Line: 2, Category: "My  Savings", Amount: "10.00"}}}

	var buf bytes.Buffer
	if err := e.Write(&buf); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	out := buf.String()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), out)
	}
	posting := strings.Fields(lines[1])
	balance := strings.Fields(lines[2])
	if len(posting) == 0 || posting[0] != "My__Savings" {
		t.Errorf("posting account: got %q, want %q", lines[1], "My__Savings")
	}
	if len(balance) != 1 || balance[0] != "My__Savings" {
		t.Errorf("balancing account: got %q, want %q", lines[2], "My__Savings")
	}
	if strings.Contains(out, `"`) {
		t.Errorf("output contains a quote:\n%s", out)
	}
}

// TestLedgerName verifies the QIF to Ledger account name mapping.
func TestLedgerName(t *testing.T) {
	for _, tc := range []struct {
		name, want string
	}{
		{"", ""},
		{"Groceries", "Groceries"},
		{"Auto:Fuel", "Auto:Fuel"},
		{"My Savings", "My Savings"},
		{"My  Savings", "My__Savings"},
		{"My  Joint Savings", "My__Joint_Savings"},
		{"a   b", "a___b"},
		{"checking", "checking"},
		{"check acct", "check_acct"},
		{"Checking Acct", "Checking Acct"},
	} {
		if got := ledgerName(tc.name); got != tc.want {
			t.Errorf("ledgerName(%q): got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestWriteTransfersOnce is a regression test for issue #42: both halves of
// a transfer are in the export, but each transfer is written once, so the
// account balances are not inflated.
func TestWriteTransfersOnce(t *testing.T) {
	r := &reader.Reader{Transactions: []*transaction.Record{
		// Checking -> Savings, both halves
		{Line: 1, Type: "Bank", Account: "Checking", Date: "2021/01/02", Payee: "To Savings",
			AmountTCode: "-200.00", ToAccount: "Savings"},
		// Savings half of the split transfer at line 20
		{Line: 10, Type: "Bank", Account: "Savings", Date: "2021/01/03", Payee: "From Checking",
			AmountTCode: "50.00", ToAccount: "Checking"},
		{Line: 20, Type: "Bank", Account: "Checking", Date: "2021/01/03", Payee: "Split",
			AmountTCode: "-70.00", Split: []*transaction.Split{
				{Line: 22, Account: "Savings", Amount: "-50.00"},
				{Line: 24, Category: "Food", Amount: "-20.00"},
			}},
		// transfer to an account that is not in the export
		{Line: 30, Type: "Bank", Account: "Checking", Date: "2021/01/04", Payee: "To Brokerage",
			AmountTCode: "-30.00", ToAccount: "Brokerage"},
		// Savings half of line 1
		{Line: 40, Type: "Bank", Account: "Savings", Date: "2021/01/02", Payee: "From Checking",
			AmountTCode: "200.00", ToAccount: "Checking"},
	}}
	l, err := Translate(r, nil)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	var buf bytes.Buffer
	skipped, written, err := l.write(&buf)
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}
	if skipped != 1 || written != 4 {
		t.Errorf("write: want 1 skipped and 4 written, got %d and %d:\n%s", skipped, written, buf.String())
	}

	got := balances(t, buf.String())
	want := map[string]int64{"Checking": -30000, "Savings": 25000, "Food": 2000, "Brokerage": 3000}
	for account, cents := range want {
		if got[account] != cents {
			t.Errorf("balance of %s: want %d cents, got %d\n%s", account, cents, got[account], buf.String())
		}
	}
	if len(got) != len(want) {
		t.Errorf("balances: want %v, got %v", want, got)
	}
}

// balances returns the balance in cents of every account posted to in a
// ledger journal written by Entry.Write. A posting with no amount balances
// its entry.
func balances(t *testing.T, journal string) map[string]int64 {
	t.Helper()
	totals := make(map[string]int64)
	var entry int64
	for _, line := range strings.Split(journal, "\n") {
		if !strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "    ;") {
			continue
		}
		if i := strings.Index(line, ";;"); i != -1 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		last := fields[len(fields)-1]
		if !strings.HasPrefix(last, "$") {
			totals[strings.Join(fields, " ")] -= entry
			entry = 0
			continue
		}
		f, err := strconv.ParseFloat(strings.ReplaceAll(last[1:], ",", ""), 64)
		if err != nil {
			t.Fatalf("posting %q: %v", line, err)
		}
		cents := int64(math.Round(f * 100))
		totals[strings.Join(fields[:len(fields)-1], " ")] += cents
		entry += cents
	}
	return totals
}

// TestWriteLogs is a regression test for issue #22: Write logs its counts
// to the logger given to Translate instead of printing them.
func TestWriteLogs(t *testing.T) {
	r := &reader.Reader{Transactions: []*transaction.Record{
		{Line: 10, Type: "Bank", Account: "Checking", Date: "2021/01/01", AmountTCode: "-5.00", Category: "Food"},
		{Line: 20, Type: "Bank", Account: "Checking", Date: "2021/01/02", AmountTCode: "0.00", Category: "Food"},
	}}
	var log bytes.Buffer
	l, err := Translate(r, slog.New(slog.NewTextHandler(&log, nil)))
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if err := l.Write(io.Discard); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if want := `level=INFO msg="ledger: write complete" written=1 skipped=1`; !strings.Contains(log.String(), want) {
		t.Errorf("log: want %q, got %q", want, log.String())
	}
}
