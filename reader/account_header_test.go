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

package reader

import (
	"testing"

	"github.com/maloquacious/qif/scanner"
)

func read(t *testing.T, input string) *Reader {
	t.Helper()
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("scanner: %v", err)
	}
	r, err := Read(sc)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return r
}

// A file with no account list, just a single !Account header followed by
// its transactions, must make that account active (issue #4).
func TestReadSingleAccountHeader(t *testing.T) {
	r := read(t, "!Account\nNChk\nTBank\n^\n!Type:Bank\nD1/ 2'16\nT10.00\n^\n")
	if len(r.Transactions) != 1 {
		t.Fatalf("transactions: want 1, got %d", len(r.Transactions))
	}
	if xact := r.Transactions[0]; xact.Account != "Chk" || xact.Type != "Bank" {
		t.Errorf("transaction: want account %q type %q, got %q %q", "Chk", "Bank", xact.Account, xact.Type)
	}
	if r.Accounts == nil || len(r.Accounts.Records) != 1 {
		t.Fatalf("accounts: want 1 record, got %+v", r.Accounts)
	}
	if name := r.Accounts.Records[0].Name; name != "Chk" {
		t.Errorf("accounts: want %q, got %q", "Chk", name)
	}
}

// The normal Quicken layout: an account list, then a single-record !Account
// header before each account's transactions.
func TestReadAccountListThenHeaders(t *testing.T) {
	input := "!Option:AutoSwitch\n" +
		"!Account\nNChk\nTBank\n^\nNVisa\nTCCard\n^\n" +
		"!Clear:AutoSwitch\n" +
		"!Account\nNChk\nTBank\n^\n" +
		"!Type:Bank\nD1/ 2'16\nT10.00\n^\n" +
		"!Account\nNVisa\nTCCard\n^\n" +
		"!Type:CCard\nD1/ 3'16\nT-5.00\n^\n"
	r := read(t, input)
	if r.Accounts == nil || len(r.Accounts.Records) != 2 {
		t.Fatalf("accounts: want 2 records, got %+v", r.Accounts)
	}
	want := []struct{ account, typ string }{{"Chk", "Bank"}, {"Visa", "CCard"}}
	if len(r.Transactions) != len(want) {
		t.Fatalf("transactions: want %d, got %d", len(want), len(r.Transactions))
	}
	for i, w := range want {
		if xact := r.Transactions[i]; xact.Account != w.account || xact.Type != w.typ {
			t.Errorf("transaction %d: want account %q type %q, got %q %q", i, w.account, w.typ, xact.Account, xact.Type)
		}
	}
}
