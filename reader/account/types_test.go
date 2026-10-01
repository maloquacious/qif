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

package account_test

import (
	"reflect"
	"testing"

	"github.com/maloquacious/qif/reader/account"
)

func TestTransactionType(t *testing.T) {
	for _, tc := range []struct {
		accountType string
		want        string
		ok          bool
	}{
		{"Bank", "Bank", true},
		{"Cash", "Cash", true},
		{"CCard", "CCard", true},
		{"Invst", "Invst", true},
		{"Oth A", "Oth A", true},
		{"Oth L", "Oth L", true},
		{"Port", "Invst", true},
		{"401(k)/403(b)", "Invst", true},
		{"Mutual", "", false},
		{"Memorized", "", false},
		{"", "", false},
	} {
		got, ok := account.TransactionType(tc.accountType)
		if got != tc.want || ok != tc.ok {
			t.Errorf("TransactionType(%q): want %q, %v, got %q, %v", tc.accountType, tc.want, tc.ok, got, ok)
		}
	}
}

func TestTransactionTypes(t *testing.T) {
	want := []string{"Bank", "Cash", "CCard", "Invst", "Oth A", "Oth L"}
	if got := account.TransactionTypes(); !reflect.DeepEqual(got, want) {
		t.Errorf("TransactionTypes: want %q, got %q", want, got)
	}
}
