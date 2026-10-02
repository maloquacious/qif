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

package normalizer

import (
	"fmt"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/account"
)

// Account is an account from the account list and the normalized
// transactions recorded in it.
type Account struct {
	Record       *account.Record
	Transactions []*Transaction
}

// ByAccount normalizes the reader's transactions (see Transactions) and
// groups them under their account. Accounts are in account-list order and
// each appears once, even with no transactions; each account's transactions
// stay in input order.
//
// It returns an error for a name that appears twice in the account list,
// and for a transaction whose account isn't in the list.
func ByAccount(r *reader.Reader) ([]*Account, error) {
	var accounts []*Account
	byName := make(map[string]*Account)
	if r.Accounts != nil {
		for _, record := range r.Accounts.Records {
			if first, ok := byName[record.Name]; ok {
				return nil, fmt.Errorf("%d: account %q: duplicate account name (first at line %d)", record.Line, record.Name, first.Record.Line)
			}
			a := &Account{Record: record}
			accounts = append(accounts, a)
			byName[record.Name] = a
		}
	}

	for _, xact := range Transactions(r.Transactions) {
		a, ok := byName[xact.Account]
		if !ok {
			return nil, fmt.Errorf("%d: transaction: account %q is not in the account list", xact.Line, xact.Account)
		}
		a.Transactions = append(a.Transactions, xact)
	}

	return accounts, nil
}
