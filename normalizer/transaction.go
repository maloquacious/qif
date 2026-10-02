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

// Package normalizer flattens QIF transactions for the writers: every
// transaction gets at least one split, and the duplicate half of each
// transfer is marked as linked.
package normalizer

import (
	"strings"

	"github.com/maloquacious/qif/reader/transaction"
)

type Transaction struct {
	Line          int
	Type          string
	Account       string
	Address       []string // Up to five lines (the sixth line is an optional message)
	Category      string
	ClearedStatus string
	Commission    string
	Date          string
	Interest      string
	IsLinked      bool
	IsZero        bool
	Memo          string
	MemorizedFlag string
	Quantity      string
	Payee         string
	Price         string
	RefNo         string
	Split         []*Split
	Ticker        string
}

type Split struct {
	Line     int
	Account  string
	Amount   string
	Category string
	Class    string
	IsLinked bool
	IsZero   bool
	Memo     string
	Ticker   string
}

func Transactions(transactions []*transaction.Record) []*Transaction {
	var normalized []*Transaction
	for _, t := range transactions {
		xact := Transaction{
			Line:          t.Line,
			Type:          t.Type,
			Date:          t.Date,
			Account:       t.Account,
			ClearedStatus: t.ClearedStatus,
			IsZero:        true, // assume the worst
			Memo:          t.Memo,
			Payee:         t.Payee,
			RefNo:         t.RefNo,
			Ticker:        t.Ticker,
		}
		if len(t.Split) == 0 {
			xact.Memo = ""
			split := Split{
				Line:     t.Line,
				Account:  t.ToAccount,
				Amount:   t.AmountTCode,
				Category: t.Category,
				Class:    t.Class,
				IsZero:   t.AmountTCode == "" || t.AmountTCode == "0.00",
				Memo:     t.Memo,
				Ticker:   t.Ticker,
			}
			xact.Split = append(xact.Split, &split)
			if !split.IsZero {
				xact.IsZero = false
			}
		} else {
			for i, line := range t.Split {
				split := Split{
					Line:     line.Line,
					Account:  line.Account,
					Amount:   line.Amount,
					IsZero:   line.Amount == "" || line.Amount == "0.00",
					Category: line.Category,
					Class:    line.Class,
					Memo:     line.Memo,
				}
				if i == 0 && split.Account == "" {
					split.Account = t.ToAccount
				}
				xact.Split = append(xact.Split, &split)
				if !split.IsZero {
					xact.IsZero = false
				}
			}
		}

		normalized = append(normalized, &xact)
	}
	linkTransfers(normalized)
	return normalized
}

// transferKey identifies one half of a transfer: the split recorded in
// account From that moves Amount to account To on Date.
type transferKey struct {
	Date, From, To, Amount string
}

// linkTransfers flags the duplicate half of each transfer between two
// accounts in the export. Quicken records a transfer twice, once in each
// account, with opposite amounts. Two transfer splits are halves of the
// same transfer when they have the same date, each names the other's
// account, and their amounts are negatives of each other. Halves are
// paired in input order.
//
// Of each pair, the half recorded in an Oth L account is the copy;
// otherwise it's the half that appears later in the input. A transfer to
// the account itself (an opening balance) and a transfer with no matching
// half (for example, to an account missing from the export) are never
// linked.
//
// A transaction is linked when every one of its splits is.
func linkTransfers(transactions []*Transaction) {
	type half struct {
		xact  *Transaction
		split *Split
	}
	unmatched := make(map[transferKey][]half)
	for _, xact := range transactions {
		for _, split := range xact.Split {
			if split.IsZero || split.Account == "" || split.Account == xact.Account {
				continue
			}
			amount := canonicalAmount(split.Amount)
			mirror := transferKey{Date: xact.Date, From: split.Account, To: xact.Account, Amount: negate(amount)}
			if halves := unmatched[mirror]; len(halves) != 0 {
				first := halves[0]
				unmatched[mirror] = halves[1:]
				if first.xact.Type == "Oth L" && xact.Type != "Oth L" {
					first.split.IsLinked = true
				} else {
					split.IsLinked = true
				}
				continue
			}
			key := transferKey{Date: xact.Date, From: xact.Account, To: split.Account, Amount: amount}
			unmatched[key] = append(unmatched[key], half{xact: xact, split: split})
		}
	}
	for _, xact := range transactions {
		xact.IsLinked = len(xact.Split) != 0
		for _, split := range xact.Split {
			if !split.IsLinked {
				xact.IsLinked = false
			}
		}
	}
}

// canonicalAmount removes thousands separators and a leading plus sign.
func canonicalAmount(amount string) string {
	return strings.TrimPrefix(strings.ReplaceAll(amount, ",", ""), "+")
}

// negate returns a canonical amount with the opposite sign.
func negate(amount string) string {
	if strings.HasPrefix(amount, "-") {
		return amount[1:]
	}
	return "-" + amount
}
