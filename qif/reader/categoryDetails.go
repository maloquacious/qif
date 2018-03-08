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

func categoryDetails(buf []byte, lineNo int) (totalBytesConsumed int, linesConsumed int, details []*qif.CategoryDetail, err error) {
	var detail *qif.CategoryDetail

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
			detail = &qif.CategoryDetail{
				IsExpense: true,
			}
		}

		switch input[0] {
		case 'B': // Budget amount (only in a Budget Amounts QIF file)
			// TODO: stop ignoring this?
		case 'D': // Description
			detail.Descr = string(input[1:])
		case 'E': // Expense category (if category is unspecified, Quicken assumes expense type)
			detail.IsExpense = true
		case 'I': // Income category
			detail.IsExpense = false
		case 'N': // Category name:subcategory name
			detail.Name = string(input[1:])
		case 'R': // Tax schedule information
			detail.TaxSchedule = string(input[1:])
		case 'T': // Tax related if included, not tax related if omitted
			detail.TaxRelated = true
		default:
			return totalBytesConsumed, linesConsumed, details, fmt.Errorf("%d: unimplemented code category/%q", lineNo+linesConsumed, string(input))
		}
	}

	if detail != nil {
		details = append(details, detail)
	}
	return totalBytesConsumed, linesConsumed, details, nil
}
