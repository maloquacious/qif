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

import "github.com/maloquacious/qif/reader/transaction"

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
	IsZero   bool
	Memo     string
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
				IsZero:   t.AmountTCode == "" || t.AmountTCode == "0.00",
				Memo:     t.Memo,
			}
			if split.Category == "" && xact.Ticker != "" {
				split.Category = xact.Ticker
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

		// flag the receiving half of linked transactions
		if len(xact.Split) == 1 || xact.Payee != "Opening Balance" {
			switch xact.Type {
			case "Oth L":
				xact.IsLinked = xact.Split[0].Account != ""
			}
		}

		normalized = append(normalized, &xact)
	}
	return normalized
}
