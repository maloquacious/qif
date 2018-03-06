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

package qif

import (
	"fmt"
)

// func (dtl *MemorizedTransaction) decode(d *data) error {
// 	switch d.code {
// 	case 'A': // Address (up to five lines; the sixth line is an optional message)
// 		dtl.Address = append(dtl.Address, string(d.val))
// 	case 'C': // Cleared status
// 		dtl.ClearedStatus = string(d.val)
// 	case 'D': // Date
// 		dtl.Date = d.date
// 	case 'E': // Memo in split
// 		if err := dtl.split(d); err != nil {
// 			return nil
// 		}
// 	case 'K':
// 		switch d.val[0] {
// 		case 'C':
// 			dtl.MemorizedTransactionType = "Check"
// 		case 'D':
// 			dtl.MemorizedTransactionType = "Deposit"
// 		case 'P':
// 			dtl.MemorizedTransactionType = "Payment"
// 		case 'I':
// 			dtl.MemorizedTransactionType = "Investment"
// 		case 'E':
// 			dtl.MemorizedTransactionType = "Electronic Payee"
// 		default:
// 			return fmt.Errorf("%d: unimplemented code memorized/%q/%s", d.line, string(d.code), string(d.val))
// 		}
// 	case 'L': // Category (Category/Subcategory/Transfer/Class)
// 		dtl.Category = string(d.val)
// 	case 'M': // Memo
// 		dtl.Memo = string(d.val)
// 	case 'N': // Num (check or reference number)
// 		dtl.Num = string(d.val)
// 	case 'P': // Payee
// 		dtl.Payee = string(d.val)
// 	case 'T': // Amount (TODO: what type?)
// 		dtl.AmountTCode = d.currency
// 	case 'S': // Category in split (Category/Transfer/Class)
// 		if err := dtl.split(d); err != nil {
// 			return nil
// 		}
// 	case 'U': // Amount (TODO: what type?)
// 		dtl.AmountUCode = d.currency
// 	case '$': // Dollar amount of split
// 		if err := dtl.split(d); err != nil {
// 			return nil
// 		}
// 	case '1': // Amortization
// 	case '2': // Amortization
// 	case '3': // Amortization
// 	case '4': // Amortization
// 	case '5': // Amortization
// 	case '6': // Amortization
// 	case '7': // Amortization
// 	default:
// 		return fmt.Errorf("%d: unimplemented code memorized/%q", d.line, string(d.code))
// 	}
// 	return nil
// }

// func (dtl *MemorizedTransaction) split(d *data) error {
// 	var split *Split
// 	if dtl.Split == nil {
// 		split = &Split{}
// 		dtl.Split = append(dtl.Split, split)
// 	} else {
// 		split = dtl.Split[len(dtl.Split)-1]
// 		if (d.code == '$' && split.setAmount) || (d.code == 'E' && split.setMemo) || (d.code == 'S' && split.setCategory) {
// 			split = &Split{}
// 			dtl.Split = append(dtl.Split, split)
// 		}
// 	}
// 	switch d.code {
// 	case '$':
// 		split.Amount = d.currency
// 		split.setAmount = true
// 	case 'E':
// 		split.Memo = string(d.val)
// 		split.setMemo = true
// 	case 'S':
// 		split.Category = string(d.val)
// 		split.setCategory = true
// 	default:
// 		return fmt.Errorf("%d: unimplemented split %q", d.line, string(d.code))
// 	}
// 	return nil
// }

func memorizedTransactionDetails(buf []byte, lineNo int) (totalBytesConsumed int, linesConsumed int, details []*MemorizedTransactionDetail, err error) {
	var detail *MemorizedTransactionDetail

	// the test for '!' stops us at the section
	for len(buf) > 0 && buf[0] != '!' {
		bytesConsumed, input, err := readline(buf)
		if err != nil {
			return totalBytesConsumed, linesConsumed, nil, err
		}
		linesConsumed++
		totalBytesConsumed += bytesConsumed
		buf = buf[bytesConsumed:]

		if len(input) == 0 {
			continue
		} else if input[0] == '^' {
			if detail != nil {
				details = append(details, detail)
				detail = nil
			}
			continue
		} else if detail == nil {
			detail = &MemorizedTransactionDetail{}
		}

		switch input[0] {
		case '1': // Budget slot 1
		case '2': // Budget slot 2
		case '3': // Budget slot 3
		case '4': // Budget slot 4
		case '5': // Budget slot 5
		case '6': // Budget slot 6
		case '7': // Budget slot 7
		case 'A': // Address (up to five lines; the sixth line is an optional message)
			if len(input) > 1 {
				detail.Address = append(detail.Address, string(input[1:]))
			}
		case 'C': // Cleared status
			detail.ClearedStatus = string(input[1:])
		case 'D': // Date
			detail.Date = bufToDate(input[1:])
		case 'E': // Memo in split
			if detail.Split == nil {
				detail.Split = append(detail.Split, &Split{})
			}
			detail.Split[len(detail.Split)-1].Memo = string(input[1:])
		case 'K':
			if len(input) > 1 {
				switch input[1] {
				case 'C': // Check transaction
					detail.Type = "CHECK"
				case 'D': // Deposit transaction
					detail.Type = "DEPOSIT"
				case 'E': // Electronic payee transaction
					detail.Type = "ELECTRONIC"
				case 'I': // Investment transaction
					detail.Type = "INVESTMENT"
				case 'P': // Payment transaction
					detail.Type = "PAYMENT"
				default:
					return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented memorizedTransaction/%q", lineNo+linesConsumed, string(input))
				}
			} else {
				return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented memorizedTransaction/%q", lineNo+linesConsumed, string(input))
			}
		case 'L': // Category (Category/Subcategory/Transfer/Class)
			detail.Category = string(input[1:])
		case 'M': // Memo
			detail.Memo = string(input[1:])
		case 'N': // Num (check or reference number)
			detail.Num = string(input[1:])
		case 'P': // Payee
			detail.Payee = string(input[1:])
		case 'T': // Amount (TODO: what type?)
			detail.AmountTCode = bufToCurrency(input[1:])
		case 'S': // Category in split (Category/Transfer/Class)
			detail.Split = append(detail.Split, &Split{Category: string(input[1:])})
		case 'U': // Amount (TODO: what type?)
			detail.AmountUCode = bufToCurrency(input[1:])
		case '$': // Dollar amount of split
			if detail.Split == nil {
				detail.Split = append(detail.Split, &Split{})
			}
			detail.Split[len(detail.Split)-1].Amount = bufToCurrency(input[1:])
		default:
			return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented code memorizedTransaction/%q", lineNo+linesConsumed, string(input))
		}
	}

	if detail != nil {
		details = append(details, detail)
	}
	return totalBytesConsumed, linesConsumed, details, nil
}
