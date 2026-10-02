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
	"bytes"
	encsv "encoding/csv"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/account"
	"github.com/maloquacious/qif/reader/category"
	"github.com/maloquacious/qif/reader/transaction"
	"github.com/maloquacious/qif/writer/csv"
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
			if _, err := csv.Translate(tc.r, nil); err != nil {
				t.Fatalf("Translate returned error: %v", err)
			}
		})
	}
}

// TestTranslateUnknownAccountType verifies that an unknown account type is
// returned as an error instead of a panic (#12).
func TestTranslateUnknownAccountType(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 12, Name: "Foo", Type: "Mutual"}}},
	}
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Translate panicked: %v", p)
			}
		}()
		_, err = csv.Translate(r, nil)
	}()
	if err == nil {
		t.Fatal("Translate: expected error, got nil")
	}
	if want := `12: account "Foo": unknown account type "Mutual"`; err.Error() != want {
		t.Errorf("Translate: want error %q, got %q", want, err)
	}
}

// TestTranslateInvestmentAccount verifies that an Invst account is
// translated (#12).
func TestTranslateInvestmentAccount(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 3, Name: "Broker", Type: "Invst"}}},
	}
	var got *csv.CSV
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Translate panicked: %v", p)
			}
		}()
		got, err = csv.Translate(r, nil)
	}()
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if want := "INV"; len(got.Accounts) != 1 || got.Accounts[0].Type != want {
		t.Errorf("Translate: want one account of type %q, got %+v", want, got.Accounts)
	}
}

// TestWriteLinkedSplits is a regression test for issues #16 and #43: the
// writer skips the linked splits of a transaction, not the whole
// transaction, writes each transfer once, and writes Oth L opening balances
// with the liability sign flip.
func TestWriteLinkedSplits(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{
			{Line: 2, Name: "Checking", Type: "Bank"},
			{Line: 4, Name: "Mortgage", Type: "Oth L"},
			{Line: 6, Name: "Savings", Type: "Bank"},
		}},
		Transactions: []*transaction.Record{
			// opening balance: not linked, written, sign flipped
			{Line: 10, Type: "Oth L", Account: "Mortgage", Date: "2021/01/01", Payee: "Opening Balance",
				AmountTCode: "-1,000.00", ToAccount: "Mortgage"},
			// single transfer: Oth L half of line 60, skipped
			{Line: 15, Type: "Oth L", Account: "Mortgage", Date: "2021/01/02", Payee: "Single",
				AmountTCode: "100.00", ToAccount: "Checking"},
			// transfer on split 0 (half of line 62): only the category row is written
			{Line: 20, Type: "Oth L", Account: "Mortgage", Date: "2021/01/03", Payee: "Split0",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 22, Account: "Checking", Amount: "100.00"},
					{Line: 24, Category: "Loan Fee", Amount: "-10.00"},
				}},
			// transfer on split 1 (half of line 64): only the category row is written
			{Line: 30, Type: "Oth L", Account: "Mortgage", Date: "2021/01/04", Payee: "Split1",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 32, Category: "Loan Fee", Amount: "-20.00"},
					{Line: 34, Account: "Checking", Amount: "110.00"},
				}},
			// category only: written
			{Line: 40, Type: "Oth L", Account: "Mortgage", Date: "2021/01/05", Payee: "Interest",
				AmountTCode: "-5.00", Category: "Interest"},
			// Oth L transfer with no other half: written
			{Line: 45, Type: "Oth L", Account: "Mortgage", Date: "2021/01/07", Payee: "Unmatched",
				AmountTCode: "50.00", ToAccount: "Checking"},
			// Bank transfer with no other half: written
			{Line: 50, Type: "Bank", Account: "Checking", Date: "2021/01/06", Payee: "Payment",
				AmountTCode: "-100.00", ToAccount: "Mortgage"},
			// Bank halves of lines 15, 20 and 30: written
			{Line: 60, Type: "Bank", Account: "Checking", Date: "2021/01/02", Payee: "Single",
				AmountTCode: "-100.00", ToAccount: "Mortgage"},
			{Line: 62, Type: "Bank", Account: "Checking", Date: "2021/01/03", Payee: "Split0",
				AmountTCode: "-100.00", ToAccount: "Mortgage"},
			{Line: 64, Type: "Bank", Account: "Checking", Date: "2021/01/04", Payee: "Split1",
				AmountTCode: "-110.00", ToAccount: "Mortgage"},
			// Bank to Bank transfer (issue #43): the half found first is written
			{Line: 70, Type: "Bank", Account: "Checking", Date: "2021/01/08", Payee: "To Savings",
				AmountTCode: "-200.00", ToAccount: "Savings"},
			{Line: 80, Type: "Bank", Account: "Savings", Date: "2021/01/08", Payee: "From Checking",
				AmountTCode: "200.00", ToAccount: "Checking"},
		},
	}
	c, err := csv.Translate(r, nil)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	var buf bytes.Buffer
	if err := c.Write(&buf); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	rows, err := encsv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	// LINE, PAYEE, SLINE, TOACCT, CATEGORY, AMOUNT, FLIPPED
	var got []string
	for _, row := range rows[1:] {
		got = append(got, strings.Join([]string{row[0], row[5], row[10], row[11], row[12], row[14], row[15]}, "|"))
	}
	want := []string{
		"10|Opening Balance|10|Mortgage||1000.00|true",
		"60|Single|60|Mortgage||-100.00|false",
		"62|Split0|62|Mortgage||-100.00|false",
		"20|Split0|24||Loan Fee|-10.00|false",
		"64|Split1|64|Mortgage||-110.00|false",
		"30|Split1|32||Loan Fee|-20.00|false",
		"40|Interest|40||Interest|-5.00|false",
		"50|Payment|50|Mortgage||-100.00|false",
		"45|Unmatched|45|Checking||50.00|false",
		"70|To Savings|70|Savings||-200.00|false",
	}
	if len(got) != len(want) {
		t.Fatalf("Write: want %d rows, got %d:\n%s", len(want), len(got), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: want %q, got %q", i, want[i], got[i])
		}
	}
}

// TestTranslateUnknownTransactionAccount is a regression test for issue #18:
// a transaction whose account is not in the account list is returned as an
// error instead of a panic.
func TestTranslateUnknownTransactionAccount(t *testing.T) {
	for _, tc := range []struct {
		name    string
		account string
		wantErr string
	}{
		{"unknown account", "Savings", `42: transaction: account "Savings" is not in the account list`},
		{"known account", "Checking", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &reader.Reader{
				Accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "Checking", Type: "Bank"}}},
				Transactions: []*transaction.Record{
					{Line: 40, Type: "Bank", Account: "Checking", Date: "2021/01/01", Payee: "Deposit", AmountTCode: "10.00"},
					{Line: 42, Type: "Bank", Account: tc.account, Date: "2021/01/01", Payee: "Fee", AmountTCode: "-1.00"},
				},
			}
			var got *csv.CSV
			var err error
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("Translate panicked: %v", p)
					}
				}()
				got, err = csv.Translate(r, nil)
			}()
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("Translate: expected error, got nil")
				}
				if err.Error() != tc.wantErr {
					t.Errorf("Translate: want error %q, got %q", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Translate returned error: %v", err)
			}
			if len(got.Transactions) != 2 {
				t.Fatalf("Translate: want 2 transactions, got %d", len(got.Transactions))
			}
			for _, x := range got.Transactions {
				if x.Account == nil || x.Account.Name != "Checking" {
					t.Errorf("transaction %d: want account %q, got %+v", x.Line, "Checking", x.Account)
				}
			}
		})
	}
}

// TestWriteLogs is a regression test for issue #22: Write logs its counts
// to the logger given to Translate instead of printing them.
func TestWriteLogs(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "Checking", Type: "Bank"}}},
		Transactions: []*transaction.Record{
			{Line: 10, Type: "Bank", Account: "Checking", Date: "2021/01/01", AmountTCode: "-5.00", Category: "Food"},
			{Line: 20, Type: "Bank", Account: "Checking", Date: "2021/01/02", AmountTCode: "0.00", Category: "Food"},
		},
	}
	var log bytes.Buffer
	c, err := csv.Translate(r, slog.New(slog.NewTextHandler(&log, nil)))
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if err := c.Write(io.Discard); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if want := `level=INFO msg="csv: write complete" written=2 skipped=1`; !strings.Contains(log.String(), want) {
		t.Errorf("log: want %q, got %q", want, log.String())
	}
}

// TestWriteHeader is a regression test for issue #21: the header names each
// column once.
func TestWriteHeader(t *testing.T) {
	c, err := csv.Translate(&reader.Reader{}, nil)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	var buf bytes.Buffer
	if err := c.Write(&buf); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	rows, err := encsv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	seen := map[string]bool{}
	for _, name := range rows[0] {
		if seen[name] {
			t.Errorf("header: column %q appears twice: %v", name, rows[0])
		}
		seen[name] = true
	}
	if !seen["SMEMO"] {
		t.Errorf("header: want an SMEMO column, got %v", rows[0])
	}
}

// TestWriteOpeningBalanceFlip is a regression test for issue #21: a
// single-line opening balance in an asset or liability account has its sign
// flipped by stdlib.FlipSign, and a zero amount is not reported as flipped.
func TestWriteOpeningBalanceFlip(t *testing.T) {
	for _, tc := range []struct{ amount, want, flipped string }{
		{"1,000.00", "-1000.00", "true"},
		{"-5.00", "5.00", "true"},
		{"+5.00", "-5.00", "true"},
	} {
		r := &reader.Reader{
			Accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "House", Type: "Oth A"}}},
			Transactions: []*transaction.Record{
				{Line: 10, Type: "Oth A", Account: "House", Date: "2021/01/01", Payee: "Opening Balance", AmountTCode: tc.amount, ToAccount: "House"},
			},
		}
		c, err := csv.Translate(r, nil)
		if err != nil {
			t.Fatalf("Translate returned error: %v", err)
		}
		var buf bytes.Buffer
		if err := c.Write(&buf); err != nil {
			t.Fatalf("Write returned error: %v", err)
		}
		rows, err := encsv.NewReader(&buf).ReadAll()
		if err != nil {
			t.Fatalf("reading output: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("%s: want 1 row, got %d", tc.amount, len(rows)-1)
		}
		if got := rows[1][14] + "|" + rows[1][15]; got != tc.want+"|"+tc.flipped {
			t.Errorf("%s: want %s|%s, got %s", tc.amount, tc.want, tc.flipped, got)
		}
	}
}
