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

package transaction

import (
	"fmt"
	"github.com/maloquacious/qif/scanner"
	"github.com/maloquacious/qif/stdlib"
	"strings"
)

type Record struct {
	Line          int
	Col           int
	Account       string
	Address       []string // Up to five lines (the sixth line is an optional message)
	AmountTCode   string
	AmountUCode   string
	BudgetAmount  []string
	Category      string // Category[:Subcategory...], without the class
	Class         string // the class from L, after the "/"
	ClearedStatus string
	Commission    string
	Date          string
	Interest      string
	Memo          string
	MemorizedFlag string
	Quantity      string
	Payee         string
	Price         string
	RefNo         string // (check or reference number)
	Split         []*Split
	Ticker        string
	ToAccount     string // if category is [xxxx], then ToAccount is 'xxxx'
	Type          string
}

type Split struct {
	Line     int    `json:"-"`
	Col      int    `json:"-"`
	Account  string `json:"account,omitempty"`
	Amount   string `json:"amount,omitempty"`
	Category string `json:"category,omitempty"`
	Class    string `json:"class,omitempty"`
	Memo     string `json:"memo,omitempty"`
}

// parseCategory splits the value of an L or S field, which has the form
// Category[:Subcategory...][/Class] or [TransferAccount][/Class], into its
// category, transfer account and class. The class starts at the first "/",
// since Quicken does not allow "/" in category names. Only one of category
// and account is set.
func parseCategory(s string) (category, account, class string) {
	if i := strings.Index(s, "/"); i != -1 {
		s, class = s[:i], s[i+1:]
	}
	if strings.HasPrefix(s, "[") {
		return "", strings.TrimSuffix(s[1:], "]"), class
	}
	return s, "", class
}

func ReadRecord(sc scanner.Scanner, account, accountType string) (*Record, scanner.Scanner, error) {
	saved, sname, record := sc, "transaction", Record{Line: sc.Line, Col: sc.Col, Account: account, Type: accountType}

	var found bool
	var category, cleared, commission, date, interest, memo, memorized, payee, qty, refNo, ticker, tcode, ucode []byte
	var split *Split
	for {
		if addrLine, bb := sc.Field("A"); addrLine != nil {
			found, record.Address = true, append(record.Address, string(addrLine))
			sc = bb
			continue
		}
		if cleared == nil {
			if cleared, sc = sc.Field("C"); cleared != nil {
				found, record.ClearedStatus = true, string(cleared)
				continue
			}
		}
		if commission == nil {
			if commission, sc = sc.Field("O"); commission != nil {
				found, record.Commission = true, string(commission)
				continue
			}
		}
		if date == nil {
			if date, sc = sc.Date("D"); date != nil {
				found, record.Date = true, string(date)
				continue
			}
		}
		if interest == nil {
			if interest, sc = sc.Field("I"); interest != nil {
				found, record.Interest = true, string(interest)
				continue
			}
		}
		if memo == nil {
			if memo, sc = sc.Field("M"); memo != nil {
				found, record.Memo = true, string(memo)
				continue
			}
		}
		if memorized == nil {
			if memorized, sc = sc.Field("K"); memorized != nil {
				found, record.MemorizedFlag = true, string(memorized)
				continue
			}
		}
		if payee == nil {
			if payee, sc = sc.Field("P"); payee != nil {
				found, record.Payee = true, string(payee)
				continue
			}
		}
		if price, bb := sc.Field("\""); price != nil {
			lexeme := strings.TrimRight(string(price), "\"")
			if fields := strings.Split(lexeme, "\""); len(fields) == 3 {
				found, record.Ticker = true, fields[0]
				record.Price = strings.ReplaceAll(fields[1], ",", "")
				date = []byte(fields[2])
				record.Date = stdlib.Date(date)
				sc = bb
				continue
			}
		}
		if qty == nil {
			if qty, sc = sc.Field("Q"); qty != nil {
				found, record.Quantity = true, string(qty)
				continue
			}
		}
		if refNo == nil {
			if refNo, sc = sc.Field("N"); refNo != nil {
				found, record.RefNo = true, string(refNo)
				continue
			}
		}
		if splitAmount, bb := sc.Field("$"); splitAmount != nil {
			if split == nil {
				split = &Split{Line: sc.Line, Col: sc.Col}
				record.Split = append(record.Split, split)
			}
			found, split.Amount = true, string(splitAmount)
			sc = bb
			continue
		}
		if splitCategory, bb := sc.Field("S"); splitCategory != nil {
			split = &Split{Line: sc.Line, Col: sc.Col}
			found, record.Split = true, append(record.Split, split)
			split.Category, split.Account, split.Class = parseCategory(string(splitCategory))
			sc = bb
			continue
		}
		if splitMemo, bb := sc.Field("E"); splitMemo != nil {
			if split == nil {
				split = &Split{Line: sc.Line, Col: sc.Col}
				record.Split = append(record.Split, split)
			}
			found, split.Memo = true, string(splitMemo)
			sc = bb
			continue
		}
		if tcode == nil {
			if tcode, sc = sc.Field("T"); tcode != nil {
				found, record.AmountTCode = true, string(tcode)
				continue
			}
		}
		if ticker == nil {
			if ticker, sc = sc.Field("Y"); ticker != nil {
				found, record.Ticker = true, string(ticker)
				continue
			}
		}
		if ucode == nil {
			if ucode, sc = sc.Field("U"); ucode != nil {
				found, record.AmountUCode = true, string(ucode)
				continue
			}
		}
		if category == nil {
			if category, sc = sc.Field("L"); category != nil {
				found = true
				record.Category, record.ToAccount, record.Class = parseCategory(string(category))
				continue
			}
		}

		// add budget amount logic here
		if record.Type == "Memorized" {
			var budgetAmount []byte
			for _, flag := range []string{"1", "2", "3", "4", "5", "6", "7"} {
				if budgetAmount == nil {
					budgetAmount, sc = sc.Field(flag)
				}
			}
			if budgetAmount != nil {
				found, record.BudgetAmount = true, append(record.BudgetAmount, string(budgetAmount))
				continue
			}
		}

		break
	}

	if !found { // no fields found
		return nil, saved, nil
	}

	// check for required fields
	switch record.Type {
	case "Memorized":
		if memorized == nil {
			return nil, saved, fmt.Errorf("%d: %s: missing field %q", record.Line, sname, "memorized")
		}
	default:
		if date == nil {
			return nil, saved, fmt.Errorf("%d: %s: missing field %q", record.Line, sname, "date")
		}
	}

	eor, bb := sc.EndOfRecord()
	if eor == nil {
		return nil, saved, fmt.Errorf("%d: %s: missing record terminator", sc.Line, sname)
	}
	sc = bb

	return &record, sc, nil
}
