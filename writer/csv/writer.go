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

// Package csv translates qif/reader data to CSV.
package csv

import (
	"cmp"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/stdlib"
)

type CSV struct {
	Accounts     []*Account
	Transactions []*Transaction
	logger       *slog.Logger
}

type Account struct {
	Line                 int
	Type                 string
	Name                 string
	CreditLimit          string
	Description          string
	StatementBalance     string
	StatementBalanceDate string
}

type Transaction struct {
	Line          int
	Type          string
	Date          string
	Account       *Account
	ClearedStatus string
	IsLinked      bool
	IsZero        bool
	Memo          string
	Payee         string
	RefNo         string
	Split         []Split
}

type Split struct {
	Line     int
	Account  string
	Amount   string
	Category string
	IsLinked bool
	IsZero   bool
	Memo     string
}

// Translate converts the reader's data. Write logs to logger; a nil logger
// discards the log.
func Translate(r *reader.Reader, logger *slog.Logger) (*CSV, error) {
	var c CSV
	c.logger = logger

	accounts, err := normalizer.ByAccount(r)
	if err != nil {
		return nil, err
	}

	for _, group := range accounts {
		account := group.Record
		var typ string
		switch account.Type {
		case "Bank":
			typ = "BNK"
		case "CCard":
			typ = "CCD"
		case "Cash":
			typ = "CSH"
		case "Oth A":
			typ = "ASS"
		case "Oth L":
			typ = "LBT"
		case "Invst":
			typ = "INV"
		case "Port":
			typ = "BRK"
		case "401(k)/403(b)":
			typ = "RET"
		default:
			return nil, fmt.Errorf("%d: account %q: unknown account type %q", account.Line, account.Name, account.Type)
		}
		acct := &Account{
			Line:                 account.Line,
			Type:                 typ,
			Name:                 account.Name,
			CreditLimit:          account.CreditLimit,
			Description:          account.Description,
			StatementBalance:     account.StatementBalance,
			StatementBalanceDate: account.StatementBalanceDate,
		}
		c.Accounts = append(c.Accounts, acct)

		for _, transaction := range group.Transactions {
			xact := &Transaction{
				Line:          transaction.Line,
				Account:       acct,
				ClearedStatus: transaction.ClearedStatus,
				Date:          transaction.Date,
				IsLinked:      transaction.IsLinked,
				IsZero:        transaction.IsZero,
				Memo:          transaction.Memo,
				Payee:         transaction.Payee,
				RefNo:         transaction.RefNo,
				Type:          transaction.Type,
			}
			for _, line := range transaction.Split {
				split := Split{
					Line:     line.Line,
					Account:  line.Account,
					Amount:   line.Amount,
					Category: line.Category,
					IsLinked: line.IsLinked,
					IsZero:   line.IsZero,
					Memo:     line.Memo,
				}
				xact.Split = append(xact.Split, split)
			}
			c.Transactions = append(c.Transactions, xact)
		}
	}

	slices.SortFunc(c.Transactions, func(a, b *Transaction) int {
		return cmp.Or(
			cmp.Compare(a.Date, b.Date),
			cmp.Compare(a.Account.Name, b.Account.Name),
			cmp.Compare(a.Line, b.Line))
	})

	return &c, nil
}

func (c *CSV) Write(w io.Writer) error {
	var skipped, written int

	cw := csv.NewWriter(w)

	record := []string{
		"LINE", "SEQ", "DATE", "STATUS", "REFNO", "PAYEE",
		"MEMO",
		"ALINE", "ATYPE", "ANAME",
		"SLINE", "TOACCT", "CATEGORY", "SMEMO", "AMOUNT", "FLIPPED",
	}
	if err := cw.Write(record); err != nil {
		return err
	}
	written++

	for _, t := range c.Transactions {
		// skip transactions that have no amount or are the receiving end of a linked transaction
		if t.IsZero || t.IsLinked {
			skipped++
			continue
		}

		var seq int
		for _, split := range t.Split {
			if split.IsZero { // skip splits that have zero amount
				continue
			}
			if split.IsLinked { // skip splits that are the receiving end of a linked transaction
				skipped++
				continue
			}

			seq++

			amount, flipped := strings.ReplaceAll(split.Amount, ",", ""), false
			if t.Payee == "Opening Balance" && len(t.Split) == 1 {
				if t.Account.Type == "ASS" || t.Account.Type == "LBT" {
					flipped = amount != "" && amount != "0.00"
					amount = stdlib.FlipSign(amount)
				}
			}

			// transaction
			record[0] = strconv.Itoa(t.Line)
			record[1] = strconv.Itoa(seq)
			record[2] = t.Date
			record[3] = t.ClearedStatus
			record[4] = t.RefNo
			record[5] = t.Payee
			record[6] = t.Memo

			// account
			record[7] = strconv.Itoa(t.Account.Line)
			record[8] = t.Account.Type
			record[9] = t.Account.Name

			// splits
			record[10] = strconv.Itoa(split.Line)
			record[11] = split.Account
			record[12] = split.Category
			record[13] = split.Memo
			record[14] = amount
			record[15] = strconv.FormatBool(flipped)

			if err := cw.Write(record); err != nil {
				return err
			}
			written++
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}

	c.log().Info("csv: write complete", "written", written, "skipped", skipped)

	return nil
}

// log returns the logger, or one that discards everything if there is none.
func (c *CSV) log() *slog.Logger {
	if c.logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return c.logger
}
