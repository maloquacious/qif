// Copyright © 2018 MICHAEL D HENDERSON <mdhender@mdhender.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mdhender/qif/qif"
)

// ImportBuffer does
func ImportBuffer(buf []byte) (*qif.File, error) {
	debug := false
	f := qif.File{}

	lineNo, autoSwitch := 0, false
	for len(buf) > 0 {
		lineNo++
		bytesConsumed, input, err := readline(buf)
		buf = buf[bytesConsumed:]

		if err != nil {
			return nil, err
		}

		switch {
		case bytes.Compare(input, []byte("!Account")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := accountDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			if autoSwitch {
				f.Accounts = append(f.Accounts, details)
			} else if len(details) == 1 {
				f.Account.Name = details[0].Name
				f.Account.Type = details[0].Type
				f.Account.CreditLimit = details[0].CreditLimit
				f.Account.Descr = details[0].Descr
				f.Account.StatementBalance = details[0].StatementBalance
				f.Account.StatementBalanceDate = details[0].StatementBalanceDate
			}
		case bytes.Compare(input, []byte("!Clear:AutoSwitch")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			autoSwitch = false
		case bytes.Compare(input, []byte("!Option:AutoSwitch")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			autoSwitch = true
		case bytes.Compare(input, []byte("!Type:Bank")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := bankDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Banks = append(f.Banks, details)
		case bytes.Compare(input, []byte("!Type:Budget")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := budgetDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Budget = append(f.Budget, details)
		case bytes.Compare(input, []byte("!Type:Cash")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := cashDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Cash = append(f.Cash, details)
		case bytes.Compare(input, []byte("!Type:Cat")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := categoryDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Categories = append(f.Categories, details)
		case bytes.Compare(input, []byte("!Type:CCard")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := creditCardDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.CreditCards = append(f.CreditCards, details)
		case bytes.Compare(input, []byte("!Type:Invst")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := investmentDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Investments = append(f.Investments, details)
		case bytes.Compare(input, []byte("!Type:Memorized")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := memorizedTransactionDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.MemorizedTransactions = append(f.MemorizedTransactions, details)
		case bytes.Compare(input, []byte("!Type:Oth A")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := otherAssetDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.OtherAssets = append(f.OtherAssets, details)
		case bytes.Compare(input, []byte("!Type:Oth L")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := otherLiabilityDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.OtherLiabilities = append(f.OtherLiabilities, details)
		case bytes.Compare(input, []byte("!Type:Prices")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := priceDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Prices = append(f.Prices, details)
		case bytes.Compare(input, []byte("!Type:Security")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := securityDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Securities = append(f.Securities, details)
		case bytes.Compare(input, []byte("!Type:Tag")) == 0:
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			bytesConsumed, linesConsumed, details, err := tagDetails(buf, lineNo)
			buf = buf[bytesConsumed:]
			lineNo += linesConsumed
			if err != nil {
				return nil, err
			}
			f.Tags = append(f.Tags, details)
		default:
			return nil, fmt.Errorf("%d: invalid section %q", lineNo, string(input))
		}
	}

	return &f, nil
}

// ImportFile does
func ImportFile(file string) (*qif.File, error) {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ImportBuffer(buf)
}
