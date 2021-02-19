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

package ledger

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Entry struct {
	Line        int
	IsZero      bool
	Account     string
	AccountType string
	Date        string
	Cleared     string
	RefNo       string
	Payee       string
	Memo        string
	Lines       Lines
}

func (e *Entry) Sort() {
	sort.Sort(e.Lines)
}

func (e *Entry) Write(w io.Writer) error {
	payee := e.Payee
	if payee == "" {
		payee = "Missing Payee"
	}

	var crp string
	if e.Cleared != "" && e.RefNo != "" {
		crp = fmt.Sprintf("%s  (%s) %s", e.Cleared, e.RefNo, payee)
	} else if e.Cleared != "" && e.RefNo == "" {
		crp = fmt.Sprintf("%s  %s", e.Cleared, payee)
	} else if e.Cleared == "" && e.RefNo != "" {
		crp = fmt.Sprintf("  (%s) %s", e.RefNo, payee)
	} else if e.Cleared == "" {
		crp = fmt.Sprintf("  %s", payee)
	}

	_, err := fmt.Fprintf(w, "%s %-59s ;; %6d %-7s %s\n", e.Date, crp, e.Line, e.AccountType, e.Account)
	if err != nil {
		return err
	}

	if e.Memo != "" {
		_, err := fmt.Fprintf(w, "    ; %s\n", e.Memo)
		if err != nil {
			return err
		}
	}
	//amount = "$" + amount

	for _, l := range e.Lines {
		// don't write lines with no amount
		//if l.IsZero {
		//	continue
		//}

		err := l.Write(w)
		if err != nil {
			return err
		}
	}

	// add a bucket to balance
	bucket := e.Account
	if e.Payee == "Opening Balance" && len(e.Lines) == 1 {
		bucket = "Equity:Opening Balances"
	} else if strings.Index(bucket, "  ") != -1 {
		bucket = fmt.Sprintf("%q", bucket)
	}
	_, err = fmt.Fprintf(w, "    %s\n", bucket)
	if err != nil {
		return err
	}

	return nil
}
