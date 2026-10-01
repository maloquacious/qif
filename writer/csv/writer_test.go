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
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/account"
	"github.com/maloquacious/qif/reader/category"
	"github.com/maloquacious/qif/reader/transaction"
	"github.com/maloquacious/qif/writer/csv"
	"strings"
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
		_, err = csv.Translate(r)
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
		got, err = csv.Translate(r)
	}()
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if want := "INV"; len(got.Accounts) != 1 || got.Accounts[0].Type != want {
		t.Errorf("Translate: want one account of type %q, got %+v", want, got.Accounts)
	}
}

// TestWriteLinkedSplits is a regression test for issue #16: the writer skips
// the linked splits of an Oth L transaction, not the whole transaction, and
// writes Oth L opening balances with the liability sign flip.
func TestWriteLinkedSplits(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{
			{Line: 2, Name: "Checking", Type: "Bank"},
			{Line: 4, Name: "Mortgage", Type: "Oth L"},
		}},
		Transactions: []*transaction.Record{
			// opening balance: not linked, written, sign flipped
			{Line: 10, Type: "Oth L", Account: "Mortgage", Date: "2021/01/01", Payee: "Opening Balance",
				AmountTCode: "-1,000.00", ToAccount: "Mortgage"},
			// single transfer: linked, skipped
			{Line: 15, Type: "Oth L", Account: "Mortgage", Date: "2021/01/02", Payee: "Single",
				AmountTCode: "100.00", ToAccount: "Checking"},
			// transfer on split 0: only the category row is written
			{Line: 20, Type: "Oth L", Account: "Mortgage", Date: "2021/01/03", Payee: "Split0",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 22, Account: "Checking", Amount: "100.00"},
					{Line: 24, Category: "Loan Fee", Amount: "-10.00"},
				}},
			// transfer on split 1: only the category row is written
			{Line: 30, Type: "Oth L", Account: "Mortgage", Date: "2021/01/04", Payee: "Split1",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 32, Category: "Loan Fee", Amount: "-20.00"},
					{Line: 34, Account: "Checking", Amount: "110.00"},
				}},
			// category only: written
			{Line: 40, Type: "Oth L", Account: "Mortgage", Date: "2021/01/05", Payee: "Interest",
				AmountTCode: "-5.00", Category: "Interest"},
			// Bank side of a transfer: written
			{Line: 50, Type: "Bank", Account: "Checking", Date: "2021/01/06", Payee: "Payment",
				AmountTCode: "-100.00", ToAccount: "Mortgage"},
		},
	}
	c, err := csv.Translate(r)
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
		"20|Split0|24||Loan Fee|-10.00|false",
		"30|Split1|32||Loan Fee|-20.00|false",
		"40|Interest|40||Interest|-5.00|false",
		"50|Payment|50|Mortgage||-100.00|false",
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
