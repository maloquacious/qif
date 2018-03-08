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
	"fmt"

	"github.com/mdhender/qif/qif"
)

func creditCardDetails(buf []byte, lineNo int) (totalBytesConsumed int, linesConsumed int, details []*qif.CreditCardDetail, err error) {
	var detail *qif.CreditCardDetail

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
			detail = &qif.CreditCardDetail{}
		}

		switch input[0] {
		case 'A': // Address (up to five lines; the sixth line is an optional message)
			detail.Address = append(detail.Address, string(input[1:]))
		case 'C': // Cleared status
			detail.ClearedStatus = string(input[1:])
		case 'D': // Date
			detail.Date = bufToDate(input[1:])
		case 'E': // Memo in split
			if detail.Split == nil {
				detail.Split = append(detail.Split, &qif.Split{})
			}
			detail.Split[len(detail.Split)-1].Memo = string(input[1:])
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
			detail.Split = append(detail.Split, &qif.Split{Category: string(input[1:])})
		case 'U': // Amount (TODO: what type?)
			detail.AmountUCode = bufToCurrency(input[1:])
		case '$': // Dollar amount of split
			if detail.Split == nil {
				detail.Split = append(detail.Split, &qif.Split{})
			}
			detail.Split[len(detail.Split)-1].Amount = bufToCurrency(input[1:])
		default:
			return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented code ccard/%q", lineNo+linesConsumed, string(input))
		}
	}

	if detail != nil {
		details = append(details, detail)
	}
	return totalBytesConsumed, linesConsumed, details, nil
}
