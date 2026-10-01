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
	"github.com/maloquacious/qif/reader/transaction"
)

// TestTransactionsClass is a regression test for issue #14: the class must
// reach the split, including the split synthesized for a non-split transaction.
func TestTransactionsClass(t *testing.T) {
	records := []*transaction.Record{
		{Line: 1, Type: "Bank", AmountTCode: "-10.00", Category: "Food", Class: "Business"},
		{Line: 5, Type: "Bank", AmountTCode: "-10.00", Split: []*transaction.Split{
			{Line: 7, Account: "Savings", Amount: "-10.00", Class: "Biz"},
		}},
	}
	got := normalizer.Transactions(records)
	if len(got) != 2 {
		t.Fatalf("Transactions: want 2 transactions, got %d", len(got))
	}
	for i, want := range []string{"Business", "Biz"} {
		if len(got[i].Split) != 1 {
			t.Fatalf("transaction %d: want 1 split, got %d", i, len(got[i].Split))
		}
		if got[i].Split[0].Class != want {
			t.Errorf("transaction %d: Class: want %q, got %q", i, want, got[i].Split[0].Class)
		}
	}
}

// TestTransactionsIsLinked is a regression test for issue #16: in an Oth L
// transaction, only the splits that transfer to another account are linked,
// and the transaction is linked only when every split is.
func TestTransactionsIsLinked(t *testing.T) {
	for _, tc := range []struct {
		name       string
		record     *transaction.Record
		wantXact   bool
		wantSplits []bool
	}{
		{
			name: "Oth L opening balance",
			record: &transaction.Record{Line: 1, Type: "Oth L", Account: "Mortgage", Payee: "Opening Balance",
				AmountTCode: "-1,000.00", ToAccount: "Mortgage"},
			wantXact:   false,
			wantSplits: []bool{false},
		},
		{
			name: "Oth L single transfer",
			record: &transaction.Record{Line: 5, Type: "Oth L", Account: "Mortgage", Payee: "Payment",
				AmountTCode: "100.00", ToAccount: "Checking"},
			wantXact:   true,
			wantSplits: []bool{true},
		},
		{
			name: "Oth L transfer on split 0",
			record: &transaction.Record{Line: 9, Type: "Oth L", Account: "Mortgage", Payee: "Payment",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 12, Account: "Checking", Amount: "100.00"},
					{Line: 14, Category: "Loan Fee", Amount: "-10.00"},
				}},
			wantXact:   false,
			wantSplits: []bool{true, false},
		},
		{
			name: "Oth L transfer on split 1",
			record: &transaction.Record{Line: 20, Type: "Oth L", Account: "Mortgage", Payee: "Payment",
				AmountTCode: "90.00", Split: []*transaction.Split{
					{Line: 23, Category: "Loan Fee", Amount: "-10.00"},
					{Line: 25, Account: "Checking", Amount: "100.00"},
				}},
			wantXact:   false,
			wantSplits: []bool{false, true},
		},
		{
			name: "Oth L category only",
			record: &transaction.Record{Line: 30, Type: "Oth L", Account: "Mortgage", Payee: "Interest",
				AmountTCode: "-5.00", Category: "Interest"},
			wantXact:   false,
			wantSplits: []bool{false},
		},
		{
			name: "Bank transfer to Oth L",
			record: &transaction.Record{Line: 35, Type: "Bank", Account: "Checking", Payee: "Payment",
				AmountTCode: "-100.00", ToAccount: "Mortgage"},
			wantXact:   false,
			wantSplits: []bool{false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizer.Transactions([]*transaction.Record{tc.record})
			if len(got) != 1 {
				t.Fatalf("Transactions: want 1 transaction, got %d", len(got))
			}
			if got[0].IsLinked != tc.wantXact {
				t.Errorf("IsLinked: want %v, got %v", tc.wantXact, got[0].IsLinked)
			}
			if len(got[0].Split) != len(tc.wantSplits) {
				t.Fatalf("want %d splits, got %d", len(tc.wantSplits), len(got[0].Split))
			}
			for i, want := range tc.wantSplits {
				if got[0].Split[i].IsLinked != want {
					t.Errorf("split %d: IsLinked: want %v, got %v", i, want, got[0].Split[i].IsLinked)
				}
			}
		})
	}
}
