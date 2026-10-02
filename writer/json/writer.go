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

// Package json translates qif/reader data to JSON.
package json

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/maloquacious/qif/normalizer"
	"github.com/maloquacious/qif/reader"
)

// JSON is the data for a JSON file. Transactions are in input order and
// include both halves of each transfer. A section the file lacks is null.
type JSON struct {
	Accounts     []Account     `json:"accounts"`
	Categories   []Category    `json:"categories"`
	Transactions []Transaction `json:"transactions"`
	logger       *slog.Logger
}

// Account is an account from the account list. Type is bank, creditCard,
// cash, asset, liability, investment, brokerage or retirement.
type Account struct {
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	CreditLimit          string `json:"credit_limit,omitempty"`
	Description          string `json:"descr,omitempty"`
	StatementBalance     string `json:"balance,omitempty"`
	StatementBalanceDate string `json:"statement_date,omitempty"`
}

// Category is a category from the category list.
type Category struct {
	Name        string `json:"name"`
	Description string `json:"descr,omitempty"`
	Income      bool   `json:"income,omitempty"`
	TaxRelated  bool   `json:"tax_related,omitempty"`
	TaxSchedule string `json:"tax_schedule,omitempty"`
}

// Transaction is a transaction. Its splits are in the "lines" key.
type Transaction struct {
	Line          int     `json:"line,omitempty"`
	Type          string  `json:"type,omitempty"`
	Date          string  `json:"date,omitempty"`
	Account       string  `json:"account,omitempty"`
	ClearedStatus string  `json:"cleared_status,omitempty"`
	Memo          string  `json:"memo,omitempty"`
	Payee         string  `json:"payee,omitempty"`
	RefNo         string  `json:"ref_no,omitempty"`
	Split         []Split `json:"lines,omitempty"`
}

// Split is one line of a transaction.
type Split struct {
	Line     int    `json:"line,omitempty"`
	Account  string `json:"account,omitempty"`
	Amount   string `json:"amount,omitempty"`
	Category string `json:"category,omitempty"`
	Memo     string `json:"memo,omitempty"`
}

// Translate converts the reader's data. Write logs to logger; a nil logger
// discards the log. It returns an error for an account of unknown type.
func Translate(r *reader.Reader, logger *slog.Logger) (*JSON, error) {
	j := JSON{logger: logger}

	if r.Accounts != nil {
		for _, account := range r.Accounts.Records {
			var typ string
			switch account.Type {
			case "Bank":
				typ = "bank"
			case "CCard":
				typ = "creditCard"
			case "Cash":
				typ = "cash"
			case "Oth A":
				typ = "asset"
			case "Oth L":
				typ = "liability"
			case "Invst":
				typ = "investment"
			case "Port":
				typ = "brokerage"
			case "401(k)/403(b)":
				typ = "retirement"
			default:
				return nil, fmt.Errorf("%d: account %q: unknown account type %q", account.Line, account.Name, account.Type)
			}
			j.Accounts = append(j.Accounts, Account{
				Type:                 typ,
				Name:                 account.Name,
				CreditLimit:          account.CreditLimit,
				Description:          account.Description,
				StatementBalance:     account.StatementBalance,
				StatementBalanceDate: account.StatementBalanceDate,
			})
		}
	}

	if r.Categories != nil {
		for _, category := range r.Categories.Records {
			j.Categories = append(j.Categories, Category{
				Name:        category.Name,
				Description: category.Description,
				Income:      category.IsIncome,
				TaxRelated:  category.IsTaxRelated,
				TaxSchedule: category.TaxSchedule,
			})
		}
	}

	for _, transaction := range normalizer.Transactions(r.Transactions) {
		xact := Transaction{
			Line:          transaction.Line,
			Type:          transaction.Type,
			Account:       transaction.Account,
			ClearedStatus: transaction.ClearedStatus,
			Date:          transaction.Date,
			Memo:          transaction.Memo,
			Payee:         transaction.Payee,
			RefNo:         transaction.RefNo,
		}
		for _, line := range transaction.Split {
			split := Split{
				Line:     line.Line,
				Account:  line.Account,
				Amount:   line.Amount,
				Category: line.Category,
				Memo:     line.Memo,
			}
			xact.Split = append(xact.Split, split)
		}
		j.Transactions = append(j.Transactions, xact)
	}

	return &j, nil
}

// Write writes the data as indented JSON.
func (j *JSON) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(j); err != nil {
		return err
	}
	j.log().Info("json: write complete",
		"accounts", len(j.Accounts), "categories", len(j.Categories), "transactions", len(j.Transactions))
	return nil
}

// log returns the logger, or one that discards everything if there is none.
func (j *JSON) log() *slog.Logger {
	if j.logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return j.logger
}
