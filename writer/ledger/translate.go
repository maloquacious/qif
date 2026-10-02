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

// Package ledger translates QIF data to Ledger (ledger-cli.org) text.
package ledger

import (
	"fmt"
	"log/slog"

	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/stdlib"
)

// Translate converts the reader's transactions, sorted by date and then
// input line. Amounts change sign, except in a single-line "Opening
// Balance". Write logs to logger; a nil logger discards the log. It
// returns an error for a single-line opening balance in an account of
// unknown type.
func Translate(r *reader.Reader, logger *slog.Logger) (*Ledger, error) {
	l := &Ledger{logger: logger}

	for _, t := range normalizer.Transactions(r.Transactions) {
		// most transactions in ledger require the opposite of the QIF sign
		flipSign, err := doFlipSign(t.Type, t.Payee, len(t.Split))
		if err != nil {
			return nil, fmt.Errorf("%d: account %q: %w", t.Line, t.Account, err)
		}

		e := &Entry{
			Line:        t.Line,
			IsLinked:    t.IsLinked,
			IsZero:      true,
			Account:     t.Account,
			AccountType: t.Type,
			Cleared:     t.ClearedStatus,
			Date:        t.Date,
			Memo:        t.Memo,
			Payee:       t.Payee,
			RefNo:       t.RefNo,
		}

		for _, split := range t.Split {
			line := &Line{
				Line:     split.Line,
				IsLinked: split.IsLinked,
				IsZero:   split.IsZero,
			}

			amount := split.Amount
			if amount == "" {
				amount = "0.00"
			} else if flipSign {
				amount = stdlib.FlipSign(amount)
			}
			line.Amount = amount

			// the first non-empty name wins
			line.Category, line.Source = "Missing Category", "none"
			for _, c := range []struct{ name, source string }{
				{split.Account, "account"},
				{split.Category, "category"},
				{split.Ticker, "ticker"},
				{split.Memo, "memo"},
			} {
				if c.name != "" {
					line.Category, line.Source = c.name, c.source
					break
				}
			}

			// a linked line is written by the entry for the other half of the transfer
			if !line.IsZero && !line.IsLinked {
				e.IsZero = false
			}

			e.Lines = append(e.Lines, line)
		}

		// the normalizer moves the memo of a transaction with no splits to
		// the split it creates (on the transaction's own line); write it
		// unless it already names the category
		if e.Memo == "" && len(e.Lines) == 1 && e.Lines[0].Line == t.Line && e.Lines[0].Source != "memo" {
			e.Memo = t.Split[0].Memo
		}

		l.Entries = append(l.Entries, e)
	}

	l.Sort()

	return l, nil
}

// most transactions in ledger require the opposite of the QIF sign,
// but a couple don't. It returns an error for a single-line opening
// balance in an account of unknown type.
func doFlipSign(accountType, payee string, numberOfLines int) (bool, error) {
	if payee != "Opening Balance" || numberOfLines != 1 {
		return true, nil
	}
	switch accountType {
	case "Bank", "Cash", "CCard", "Oth A", "Oth L",
		// investment accounts are assets, so they follow the same rule as
		// the asset types (Bank, Cash, Oth A): no flip for a single-line
		// opening balance.
		"Invst", "Port", "401(k)/403(b)":
		return false, nil
	}
	return false, fmt.Errorf("unknown account type %q", accountType)
}
