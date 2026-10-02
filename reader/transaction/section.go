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

// Package transaction implements a simple parser for transaction data.
// It returns the first error found with the data.
package transaction

import (
	"github.com/maloquacious/qif/reader/account"
	"github.com/maloquacious/qif/reader/internal/section"
	"github.com/maloquacious/qif/scanner"
)

type Section struct {
	Line    int       `json:"-"`
	Col     int       `json:"-"`
	Records []*Record `json:"records,omitempty"`
}

// ReadSection reads the transaction section for an account. The section
// header is the one for the account's type (see account.TransactionType),
// so the transactions of a "Port" account are read from a "!Type:Invst"
// section. Each record keeps the account's own type. The accountType may
// also be "Memorized" or "Prices" for those sections.
func ReadSection(sc scanner.Scanner, accountName, accountType string) (*Section, scanner.Scanner, error) {
	var header string
	switch accountType {
	case "Memorized", "Prices":
		header = "!Type:" + accountType
	default:
		transactionType, ok := account.TransactionType(accountType)
		if !ok {
			// unknown (or empty) account type, so this can't be a transaction section
			return nil, sc, nil
		}
		header = "!Type:" + transactionType
	}

	readRecord := func(sc scanner.Scanner) (*Record, scanner.Scanner, error) {
		return ReadRecord(sc, accountName, accountType)
	}
	s, sc, err := section.Read(sc, header, "transactions", readRecord)
	if s == nil {
		return nil, sc, err
	}
	return &Section{Line: s.Line, Col: s.Col, Records: s.Records}, sc, nil
}
