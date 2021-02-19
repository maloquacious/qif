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
	"github.com/mdhender/qif/normalizer"
	"github.com/mdhender/qif/reader"
	"io"
)

type JSON struct {
	Accounts     []Account     `json:"accounts"`
	Categories   []Category    `json:"categories"`
	Transactions []Transaction `json:"transactions"`
}

type Account struct {
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	CreditLimit          string `json:"credit_limit,omitempty"`
	Description          string `json:"descr,omitempty"`
	StatementBalance     string `json:"balance,omitempty"`
	StatementBalanceDate string `json:"statement_date,omitempty"`
}

type Category struct {
	Name        string `json:"name"`
	Description string `json:"descr,omitempty"`
	Income      bool   `json:"income,omitempty"`
	TaxRelated  bool   `json:"tax_related,omitempty"`
	TaxSchedule string `json:"tax_schedule,omitempty"`
}

type Transaction struct {
	Line          int     `json:"line,omitempty"`
	Type          string  `json:"type,omitempty"`
	Date          string  `json:"date,omitempty"`
	Account       string  `json:"account,omitempty"`
	ToAccount     string  `json:"to_account,omitempty"`
	Amount        string  `json:"amount,omitempty"`
	Category      string  `json:"category,omitempty"`
	ClearedStatus string  `json:"cleared_status,omitempty"`
	Memo          string  `json:"memo,omitempty"`
	Payee         string  `json:"payee,omitempty"`
	RefNo         string  `json:"ref_no,omitempty"`
	Split         []Split `json:"lines,omitempty"`
}

type Split struct {
	Line     int    `json:"line,omitempty"`
	Account  string `json:"account,omitempty"`
	Amount   string `json:"amount,omitempty"`
	Category string `json:"category,omitempty"`
	Memo     string `json:"memo,omitempty"`
}

func Translate(r *reader.Reader) (*JSON, error) {
	var j JSON

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
		case "Port":
			typ = "brokerage"
		case "401(k)/403(b)":
			typ = "retirement"
		default:
			panic(fmt.Sprintf("assert(account.type != %q)", account.Type))
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

	for _, category := range r.Categories.Records {
		j.Categories = append(j.Categories, Category{
			Name:        category.Name,
			Description: category.Description,
			Income:      category.IsIncome,
			TaxRelated:  category.IsTaxRelated,
			TaxSchedule: category.TaxSchedule,
		})
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

func (j *JSON) Write(w io.Writer) error {
	buf, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("short write")
	}
	fmt.Printf("json: wrote %8d accounts\n", len(j.Accounts))
	fmt.Printf("json: wrote %8d categories\n", len(j.Categories))
	fmt.Printf("json: wrote %8d transactions\n", len(j.Transactions))
	return nil
}
