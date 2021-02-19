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

// Package reader implements a simple QIF parser. It reads the QIF data and
// converts it to structs with no attempt at cleaning up the data. If there
// are errors that prevent parsing (mostly missing fields), it will return
// only the first error found. The error should include the line and column
// in the original data file to help with troubleshooting.
//
// There are some errors that just cause a panic. I'm not sure why.
package reader

import (
	"fmt"
	"github.com/mdhender/qif/reader/account"
	"github.com/mdhender/qif/reader/category"
	"github.com/mdhender/qif/reader/security"
	"github.com/mdhender/qif/reader/tag"
	"github.com/mdhender/qif/reader/transaction"
	"github.com/mdhender/qif/scanner"
)

type Reader struct {
	active struct {
		account     string
		accountType string
	}
	Accounts     *account.Section      `json:"accounts,omitempty"`
	Categories   *category.Section     `json:"categories,omitempty"`
	Securities   *security.Section     `json:"securities,omitempty"`
	Tags         *tag.Section          `json:"tags,omitempty"`
	Transactions []*transaction.Record `json:"transactions,omitempty"`
	Memorized    []*transaction.Record `json:"-"`
	Prices       []*transaction.Record `json:"-"`
}

func Read(sc scanner.Scanner) (*Reader, error) {
	var r Reader
	for len(sc.Buffer) != 0 {
		if literal, bb := sc.Literal("!Clear:AutoSwitch"); literal != nil {
			// ignore
			sc = bb
			continue
		}
		if literal, bb := sc.Literal("!Option:AutoSwitch"); literal != nil {
			// ignore
			sc = bb
			continue
		}
		if section, bb, err := account.ReadSection(sc); err != nil {
			return nil, err
		} else if section != nil {
			if len(section.Records) != 0 {
				if r.Accounts == nil {
					r.Accounts = section
				} else if len(section.Records) == 1 {
					r.active.account = section.Records[0].Name
					r.active.accountType = section.Records[0].Type
				} else {
					panic("!")
				}
			}
			sc = bb
			continue
		}
		if section, bb, err := category.ReadSection(sc); err != nil {
			return nil, err
		} else if section != nil {
			if len(section.Records) != 0 {
				if r.Categories == nil {
					r.Categories = section
				} else {
					panic("!")
				}
			}
			sc = bb
			continue
		}
		if section, bb, err := security.ReadSection(sc); err != nil {
			return nil, err
		} else if section != nil {
			if len(section.Records) != 0 {
				if r.Securities == nil {
					r.Securities = &security.Section{
						Line: sc.Line,
						Col:  sc.Col,
					}
				}
				r.Securities.Records = append(r.Securities.Records, section.Records...)
			}
			sc = bb
			continue
		}
		if section, bb, err := tag.ReadSection(sc); err != nil {
			return nil, err
		} else if section != nil {
			if len(section.Records) != 0 {
				if r.Tags == nil {
					r.Tags = section
				} else {
					panic("!")
				}
			}
			sc = bb
			continue
		}
		if section, bb, err := transaction.ReadSection(sc, r.active.account, r.active.accountType); err != nil {
			return nil, err
		} else if section != nil {
			for _, xact := range section.Records {
				r.Transactions = append(r.Transactions, xact)
			}
			sc = bb
			continue
		}
		if section, bb, err := transaction.ReadSection(sc, "", "Memorized"); err != nil {
			return nil, err
		} else if section != nil {
			for _, xact := range section.Records {
				r.Memorized = append(r.Memorized, xact)
			}
			sc = bb
			continue
		}
		if section, bb, err := transaction.ReadSection(sc, "", "Prices"); err != nil {
			return nil, err
		} else if section != nil {
			for _, xact := range section.Records {
				r.Prices = append(r.Prices, xact)
			}
			sc = bb
			continue
		}
		return nil, fmt.Errorf("%d:%d: unexpected input", sc.Line, sc.Col)
	}
	return &r, nil
}
