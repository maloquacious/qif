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
