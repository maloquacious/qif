/*
 * qif - a package to convert QIF data
 *
 * Copyright (c) 2026 Michael D Henderson
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

package qif_test

import (
	"fmt"
	"log"
	"os"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/scanner"
	"github.com/maloquacious/qif/writer/ledger"
)

// This example reads a QIF export with one bank account and writes it as
// Ledger text. The other writers, csv and json, are used the same way.
func Example() {
	input := []byte(`!Account
NChecking
TBank
^
!Type:Bank
D1/ 2'24
T-12.50
PCoffee Shop
LDining:Coffee
^
`)

	sc, err := scanner.New(input)
	if err != nil {
		log.Fatal(err)
	}
	r, err := reader.Read(sc)
	if err != nil {
		log.Fatal(err)
	}

	// sections missing from the file are nil
	fmt.Println("categories:", r.Categories == nil)
	t := r.Transactions[0]
	fmt.Println(t.Account, t.Type, t.Date, t.AmountTCode, t.Category)

	l, err := ledger.Translate(r, nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := l.Write(os.Stdout); err != nil {
		log.Fatal(err)
	}
	// Output:
	// categories: true
	// Checking Bank 2024/01/02 -12.50 Dining:Coffee
	// 2024/01/02   Coffee Shop                                               ;;      6 Bank    Checking
	//     Dining:Coffee                                               $12.50 ;;      6 category
	//     Checking
}
