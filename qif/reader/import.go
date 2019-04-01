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
	f := qif.File{
		Accounts: make(map[string]*qif.AccountDetail),
	}

	// state will be
	//    0 => never saw flag
	//    1 => set flag first time
	//    2 => cleared flag
	//    3 => set flag two or more times
	autoSwitchState := 0
	accountDefineMode := false
	accountAccumulateMode := false

	var defaultAccount *qif.AccountDetail
	lineNo := 0
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
			if accountDefineMode {
				// there should never be duplicate accounts.
				for _, acct := range details {
					if f.Accounts[acct.Name] != nil {
						return nil, fmt.Errorf("%d: duplicate account %q", lineNo, acct.Name)
					}
					f.Accounts[acct.Name] = acct
				}
			} else if accountAccumulateMode {
				if len(details) == 0 {
					return nil, fmt.Errorf("%d: missing account", lineNo)
				} else if len(details) != 1 {
					return nil, fmt.Errorf("%d: unexpected account %q", lineNo, details[1].Name)
				}
				defaultAccount = details[0]
			} else {
				// update the global account
				if len(details) == 0 {
					return nil, fmt.Errorf("%d: expected global account", lineNo)
				} else if len(details) != 1 {
					return nil, fmt.Errorf("%d: unexpected global account %q", lineNo, details[1].Name)
				}
				f.Account.Name = details[0].Name
				f.Account.Type = details[0].Type
				f.Account.CreditLimit = details[0].CreditLimit
				f.Account.Descr = details[0].Descr
				f.Account.StatementBalance = details[0].StatementBalance
				f.Account.StatementBalanceDate = details[0].StatementBalanceDate
			}
		case bytes.Compare(input, []byte("!Clear:AutoSwitch")) == 0:
			// when auto switch is false, the current account never changes.
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			switch autoSwitchState {
			case 0:
				// odd. didn't set before clearing.
				autoSwitchState = 2
			case 1:
				accountDefineMode = false
				accountAccumulateMode = false
			case 2:
				// odd. didn't set before clearing.
				accountDefineMode = false
				accountAccumulateMode = false
			case 3:
				autoSwitchState = 2
				accountDefineMode = false
				accountAccumulateMode = false
			}
		case bytes.Compare(input, []byte("!Option:AutoSwitch")) == 0:
			// There are two modes for AutoSwitch.
			// If we haven't encountered Clear:AutoSwitch, then this
			// declares a list of accounts in the file.
			// That list ends with Clear:AutoSwitch (or the first non-Account
			// line that follows).
			// The other mode occurs when we have already declared the list
			// of accounts. Now every reference to !Account changes the default
			// account that transactions are saved to.
			if debug {
				log.Printf("%5d: %s\n", lineNo, string(input))
			}
			switch autoSwitchState {
			case 0:
				// first time setting
				autoSwitchState = 1
				accountDefineMode = true
				accountAccumulateMode = false
			case 1:
				// odd. never cleared the previous. jump.
				autoSwitchState = 3
				accountDefineMode = false
				accountAccumulateMode = true
			case 2:
				autoSwitchState = 3
				accountDefineMode = false
				accountAccumulateMode = false
			case 3:
				// odd. never cleared the previous.
				accountDefineMode = false
				accountAccumulateMode = true
			}
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
			if defaultAccount == nil {
				f.Banks = append(f.Banks, details...)
			} else {
				defaultAccount.Banks = append(defaultAccount.Banks, details...)
			}
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
			if defaultAccount == nil {
				f.Budget = append(f.Budget, details...)
			} else {
				defaultAccount.Budget = append(defaultAccount.Budget, details...)
			}
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
			if defaultAccount == nil {
				f.Cash = append(f.Cash, details...)
			} else {
				defaultAccount.Cash = append(defaultAccount.Cash, details...)
			}
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
			f.Categories = append(f.Categories, details...)
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
			if defaultAccount == nil {
				f.CreditCards = append(f.CreditCards, details...)
			} else {
				defaultAccount.CreditCards = append(defaultAccount.CreditCards, details...)
			}
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
			if defaultAccount == nil {
				f.Investments = append(f.Investments, details...)
			} else {
				defaultAccount.Investments = append(defaultAccount.Investments, details...)
			}
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
			f.MemorizedTransactions = append(f.MemorizedTransactions, details...)
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
			f.OtherAssets = append(f.OtherAssets, details...)
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
			f.OtherLiabilities = append(f.OtherLiabilities, details...)
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
			f.Prices = append(f.Prices, details...)
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
			f.Securities = append(f.Securities, details...)
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
			f.Tags = append(f.Tags, details...)
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
