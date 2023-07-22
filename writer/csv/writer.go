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
	"encoding/csv"
	"fmt"
	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
	"io"
	"sort"
	"strings"
)

type CSV struct {
	Accounts     []*Account
	Transactions []*Transaction `json:"transactions"`
	Map          struct {
		Accounts map[string]*Account
	}
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
	ToAccount     string
	Amount        string
	Category      string
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
	IsZero   bool
	Memo     string
}

func Translate(r *reader.Reader) (*CSV, error) {
	var c CSV
	c.Map.Accounts = make(map[string]*Account)

	for _, account := range r.Accounts.Records {
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
		case "Port":
			typ = "BRK"
		case "401(k)/403(b)":
			typ = "RET"
		default:
			panic(fmt.Sprintf("assert(account.type != %q)", account.Type))
		}
		a := &Account{
			Line:                 account.Line,
			Type:                 typ,
			Name:                 account.Name,
			CreditLimit:          account.CreditLimit,
			Description:          account.Description,
			StatementBalance:     account.StatementBalance,
			StatementBalanceDate: account.StatementBalanceDate,
		}
		c.Accounts = append(c.Accounts, a)
		c.Map.Accounts[account.Name] = a
	}

	for _, transaction := range normalizer.Transactions(r.Transactions) {
		xact := &Transaction{
			Line:          transaction.Line,
			Account:       c.Map.Accounts[transaction.Account],
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
				IsZero:   line.IsZero,
				Memo:     line.Memo,
			}
			xact.Split = append(xact.Split, split)
		}
		c.Transactions = append(c.Transactions, xact)
	}

	sort.Sort(&c)

	return &c, nil
}

func (c *CSV) Write(w io.Writer) error {
	var skipped, written int

	cw := csv.NewWriter(w)

	record := []string{
		"LINE", "SEQ", "DATE", "STATUS", "REFNO", "PAYEE",
		"MEMO",
		"ALINE", "ATYPE", "ANAME",
		"SLINE", "TOACCT", "CATEGORY", "MEMO", "AMOUNT", "FLIPPED",
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

			seq++

			amount, flipped := strings.ReplaceAll(split.Amount, ",", ""), false
			if t.Payee == "Opening Balance" && len(t.Split) == 1 {
				if t.Account.Type == "ASS" || t.Account.Type == "LBT" {
					if amount == "" || amount == "0.00" {
						amount, flipped = "0.00", false
					} else if amount[0] == '-' {
						amount, flipped = amount[1:], true
					} else if amount[0] == '+' {
						amount, flipped = "-"+amount[1:], true
					} else {
						amount, flipped = "-"+amount, true
					}
				}
			}

			// transaction
			record[0] = fmt.Sprintf("%d", t.Line)
			record[1] = fmt.Sprintf("%d", seq)
			record[2] = t.Date
			record[3] = t.ClearedStatus
			record[4] = t.RefNo
			record[5] = t.Payee
			record[6] = t.Memo

			// account
			record[7] = fmt.Sprintf("%d", t.Account.Line)
			record[8] = t.Account.Type
			record[9] = t.Account.Name

			// splits
			record[10] = fmt.Sprintf("%d", split.Line)
			record[11] = split.Account
			record[12] = split.Category
			record[13] = split.Memo
			record[14] = amount
			record[15] = fmt.Sprintf("%v", flipped)

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

	fmt.Printf("csv: skipped   %8d records\n", skipped)
	fmt.Printf("csv: wrote     %8d records\n", written)

	return nil
}

func (c *CSV) Len() int {
	return len(c.Transactions)
}

func (c *CSV) Less(i, j int) bool {
	if c.Transactions[i].Date < c.Transactions[j].Date {
		return true
	}
	if c.Transactions[i].Date > c.Transactions[j].Date {
		return false
	}
	if c.Transactions[i].Account.Name < c.Transactions[j].Account.Name {
		return true
	}
	if c.Transactions[i].Account.Name > c.Transactions[j].Account.Name {
		return false
	}
	return c.Transactions[i].Line < c.Transactions[j].Line
}

func (c *CSV) Swap(i, j int) {
	c.Transactions[i], c.Transactions[j] = c.Transactions[j], c.Transactions[i]
}
