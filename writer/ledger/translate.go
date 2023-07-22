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

// Package ledger translates QIF data to ledger.
package ledger

import (
	"fmt"
	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/stdlib"
)

func Translate(r *reader.Reader) (*LEDGER, error) {
	l := &LEDGER{}

	for _, t := range normalizer.Transactions(r.Transactions) {
		// most transactions in ledger require the opposite of the QIF sign
		if t.Payee == "Opening Balance" {
			fmt.Printf("%06d: %5q %q %d\n", t.Line, t.Type, t.Payee, len(t.Split))
		}
		flipSign := doFlipSign(t.Type, t.Payee, len(t.Split))

		e := &Entry{
			Line:        t.Line,
			IsZero:      true,
			Account:     t.Account,
			AccountType: t.Type,
			Cleared:     t.ClearedStatus,
			Date:        t.Date,
			Payee:       t.Payee,
			RefNo:       t.RefNo,
		}

		for _, split := range t.Split {
			line := &Line{
				Line:   split.Line,
				IsZero: split.IsZero,
			}

			amount := split.Amount
			if t.Payee == "Opening Balance" {
				fmt.Printf("%06d: %5q %q %d %v %q\n", t.Line, t.Type, t.Payee, len(t.Split), flipSign, amount)
			}
			if amount == "" {
				amount = "0.00"
			} else if flipSign {
				amount = stdlib.FlipSign(amount)
				if t.Payee == "Opening Balance" {
					fmt.Printf("%06d: %5q %q %d %v %q\n", t.Line, t.Type, t.Payee, len(t.Split), flipSign, amount)
				}
			}
			line.Amount = amount

			if line.Category == "" {
				line.Category, line.Source = split.Account, "account"
			}
			if line.Category == "" {
				line.Category, line.Source = split.Category, "category"
			}
			if line.Category == "" {
				line.Category, line.Source = split.Memo, "memo"
			}
			if line.Category == "" {
				line.Category, line.Source = "Missing Category", "none"
			}

			if !line.IsZero {
				e.IsZero = false
			}

			e.Lines = append(e.Lines, line)
		}

		l.Entries = append(l.Entries, e)
	}

	l.Sort()

	return l, nil
}

// most transactions in ledger require the opposite of the QIF sign,
// but a couple don't.
func doFlipSign(accountType, payee string, numberOfLines int) bool {
	if payee != "Opening Balance" {
		return true
	}
	if numberOfLines != 1 {
		return true
	}
	switch accountType {
	case "Bank":
		return false
	case "Cash":
		return false
	case "CCard":
		return false
	case "Oth A":
		return false
	case "Oth L":
		return false
	}
	panic(fmt.Sprintf("assert(account.type != %q)", accountType))
}
