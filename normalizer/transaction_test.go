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

// TestTransactionsIsLinked is a regression test for issues #16, #42 and #43.
// The two halves of a transfer are matched by date, accounts and opposite
// amounts; only the duplicate half is linked, and a transaction is linked
// only when every split is.
func TestTransactionsIsLinked(t *testing.T) {
	// xfer returns a transaction without splits
	xfer := func(line int, typ, account, date, amount, to string) *transaction.Record {
		return &transaction.Record{Line: line, Type: typ, Account: account, Date: date, AmountTCode: amount, ToAccount: to}
	}
	for _, tc := range []struct {
		name    string
		records []*transaction.Record
		// want[i][j] is IsLinked of split j of transaction i
		want     [][]bool
		wantXact []bool
	}{
		{
			name:     "opening balance is a self-transfer",
			records:  []*transaction.Record{xfer(1, "Oth L", "Mortgage", "2021/01/01", "-1,000.00", "Mortgage")},
			want:     [][]bool{{false}},
			wantXact: []bool{false},
		},
		{
			name: "bank to bank: later half is linked",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "-200.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/02", "200.00", "Checking"),
			},
			want:     [][]bool{{false}, {true}},
			wantXact: []bool{false, true},
		},
		{
			name: "Oth L half is linked even when it comes first",
			records: []*transaction.Record{
				xfer(1, "Oth L", "Mortgage", "2021/01/02", "100.00", "Checking"),
				xfer(5, "Bank", "Checking", "2021/01/02", "-100.00", "Mortgage"),
			},
			want:     [][]bool{{true}, {false}},
			wantXact: []bool{true, false},
		},
		{
			name:     "transfer without its other half is kept",
			records:  []*transaction.Record{xfer(1, "Oth L", "Mortgage", "2021/01/02", "100.00", "Checking")},
			want:     [][]bool{{false}},
			wantXact: []bool{false},
		},
		{
			name: "different dates do not match",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "-200.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/03", "200.00", "Checking"),
			},
			want:     [][]bool{{false}, {false}},
			wantXact: []bool{false, false},
		},
		{
			name: "same-sign amounts do not match",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "-200.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/02", "-200.00", "Checking"),
			},
			want:     [][]bool{{false}, {false}},
			wantXact: []bool{false, false},
		},
		{
			name: "amounts match ignoring commas and plus sign",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "-1,200.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/02", "+1200.00", "Checking"),
			},
			want:     [][]bool{{false}, {true}},
			wantXact: []bool{false, true},
		},
		{
			name: "two identical transfers on one day pair up separately",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "-50.00", "Savings"),
				xfer(3, "Bank", "Checking", "2021/01/02", "-50.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/02", "50.00", "Checking"),
				xfer(7, "Bank", "Savings", "2021/01/02", "50.00", "Checking"),
			},
			want:     [][]bool{{false}, {false}, {true}, {true}},
			wantXact: []bool{false, false, true, true},
		},
		{
			name: "zero-amount transfer is not linked",
			records: []*transaction.Record{
				xfer(1, "Bank", "Checking", "2021/01/02", "0.00", "Savings"),
				xfer(5, "Bank", "Savings", "2021/01/02", "0.00", "Checking"),
			},
			want:     [][]bool{{false}, {false}},
			wantXact: []bool{false, false},
		},
		{
			name: "split transfer first: the other account's transaction is linked",
			records: []*transaction.Record{
				{Line: 1, Type: "Bank", Account: "Checking", Date: "2021/01/02", AmountTCode: "-70.00", Split: []*transaction.Split{
					{Line: 3, Account: "Savings", Amount: "-50.00"},
					{Line: 5, Category: "Food", Amount: "-20.00"},
				}},
				xfer(9, "Bank", "Savings", "2021/01/02", "50.00", "Checking"),
			},
			want:     [][]bool{{false, false}, {true}},
			wantXact: []bool{false, true},
		},
		{
			name: "split transfer second: only the transfer split is linked",
			records: []*transaction.Record{
				xfer(1, "Bank", "Savings", "2021/01/02", "50.00", "Checking"),
				{Line: 5, Type: "Bank", Account: "Checking", Date: "2021/01/02", AmountTCode: "-70.00", Split: []*transaction.Split{
					{Line: 7, Category: "Food", Amount: "-20.00"},
					{Line: 9, Account: "Savings", Amount: "-50.00"},
				}},
			},
			want:     [][]bool{{false}, {false, true}},
			wantXact: []bool{false, false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizer.Transactions(tc.records)
			if len(got) != len(tc.want) {
				t.Fatalf("Transactions: want %d transactions, got %d", len(tc.want), len(got))
			}
			for i, xact := range got {
				if xact.IsLinked != tc.wantXact[i] {
					t.Errorf("transaction %d: IsLinked: want %v, got %v", i, tc.wantXact[i], xact.IsLinked)
				}
				if len(xact.Split) != len(tc.want[i]) {
					t.Fatalf("transaction %d: want %d splits, got %d", i, len(tc.want[i]), len(xact.Split))
				}
				for j, want := range tc.want[i] {
					if xact.Split[j].IsLinked != want {
						t.Errorf("transaction %d split %d: IsLinked: want %v, got %v", i, j, want, xact.Split[j].IsLinked)
					}
				}
			}
		})
	}
}
