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

import "fmt"

func accountDetails(buf []byte, lineNo int) (totalBytesConsumed int, linesConsumed int, details []*AccountDetail, err error) {
	var detail *AccountDetail

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
			detail = &AccountDetail{}
		}

		switch input[0] {
		case '/': // statement balance date
			detail.StatementBalanceDate = bufToDate(input[1:])
		case '$': // statement balance
			detail.StatementBalance = bufToCurrency(input[1:])
		case 'D': // description
			detail.Descr = string(input[1:])
		case 'L': // credit limit (only for credit card account)
			detail.CreditLimit = bufToCurrency(input[1:])
		case 'N': // name
			detail.Name = string(input[1:])
		case 'T': // type of account
			detail.Type = string(input[1:])
		default:
			return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented code account/%q", lineNo+linesConsumed, string(input))
		}
	}

	if detail != nil {
		details = append(details, detail)
	}
	return totalBytesConsumed, linesConsumed, details, nil
}
