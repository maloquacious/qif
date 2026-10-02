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

package normalizer_test

import (
	"testing"

	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/account"
	"github.com/maloquacious/qif/reader/transaction"
)

// TestByAccount is a test for issue #53: transactions are grouped under
// their account in account-list order, every account appears once, and
// transfers are linked across the whole input.
func TestByAccount(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{
			{Line: 2, Name: "Savings", Type: "Bank"},
			{Line: 4, Name: "Empty", Type: "Cash"},
			{Line: 6, Name: "Checking", Type: "Bank"},
		}},
		Transactions: []*transaction.Record{
			{Line: 10, Type: "Bank", Account: "Checking", Date: "2021/01/02", AmountTCode: "-100.00", ToAccount: "Savings"},
			{Line: 12, Type: "Bank", Account: "Checking", Date: "2021/01/01", AmountTCode: "-5.00", Category: "Fees"},
			{Line: 20, Type: "Bank", Account: "Savings", Date: "2021/01/02", AmountTCode: "100.00", ToAccount: "Checking"},
		},
	}
	got, err := normalizer.ByAccount(r)
	if err != nil {
		t.Fatalf("ByAccount returned error: %v", err)
	}
	want := []struct {
		name   string
		lines  []int
		linked []bool
	}{
		{"Savings", []int{20}, []bool{true}},
		{"Empty", nil, nil},
		{"Checking", []int{10, 12}, []bool{false, false}},
	}
	if len(got) != len(want) {
		t.Fatalf("ByAccount: want %d accounts, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].Record.Name != w.name {
			t.Errorf("account %d: want %q, got %q", i, w.name, got[i].Record.Name)
			continue
		}
		if len(got[i].Transactions) != len(w.lines) {
			t.Errorf("account %q: want %d transactions, got %d", w.name, len(w.lines), len(got[i].Transactions))
			continue
		}
		for j, xact := range got[i].Transactions {
			if xact.Line != w.lines[j] {
				t.Errorf("account %q: transaction %d: want line %d, got %d", w.name, j, w.lines[j], xact.Line)
			}
			if xact.IsLinked != w.linked[j] {
				t.Errorf("account %q: transaction %d: IsLinked: want %v, got %v", w.name, j, w.linked[j], xact.IsLinked)
			}
		}
	}
}

// TestByAccountErrors is a test for issue #53: a transaction in an account
// that isn't in the list, and a name listed twice, are errors.
func TestByAccountErrors(t *testing.T) {
	for _, tc := range []struct {
		name     string
		accounts *account.Section
		wantErr  string
	}{
		{
			name:     "no account list",
			accounts: nil,
			wantErr:  `40: transaction: account "Checking" is not in the account list`,
		},
		{
			name:     "unknown account",
			accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "Savings", Type: "Bank"}}},
			wantErr:  `40: transaction: account "Checking" is not in the account list`,
		},
		{
			name: "duplicate account name",
			accounts: &account.Section{Records: []*account.Record{
				{Line: 2, Name: "Checking", Type: "Bank"},
				{Line: 4, Name: "Checking", Type: "Bank"},
			}},
			wantErr: `4: account "Checking": duplicate account name (first at line 2)`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &reader.Reader{
				Accounts: tc.accounts,
				Transactions: []*transaction.Record{
					{Line: 40, Type: "Bank", Account: "Checking", Date: "2021/01/01", AmountTCode: "10.00"},
				},
			}
			got, err := normalizer.ByAccount(r)
			if err == nil {
				t.Fatalf("ByAccount: expected error, got %+v", got)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("ByAccount: want error %q, got %q", tc.wantErr, err)
			}
		})
	}
}

// TestByAccountEmpty checks that a reader with no accounts and no
// transactions yields no accounts and no error.
func TestByAccountEmpty(t *testing.T) {
	got, err := normalizer.ByAccount(&reader.Reader{})
	if err != nil || len(got) != 0 {
		t.Errorf("ByAccount: want no accounts and no error, got %+v, %v", got, err)
	}
}
