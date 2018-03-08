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

import "fmt"

// bufToCurrency translates a dollar amount to an integer
func bufToCurrency(b []byte) int {
	var isNegative bool
	pos := 0
	if pos < len(b) && b[pos] == '-' {
		isNegative = true
		pos++
	}
	var dollars, cents []byte
	for ; pos < len(b) && b[pos] != '.'; pos++ {
		if '0' <= b[pos] && b[pos] <= '9' {
			dollars = append(dollars, b[pos])
		}
	}
	for ; pos < len(b); pos++ {
		if '0' <= b[pos] && b[pos] <= '9' {
			cents = append(cents, b[pos])
		}
	}
	if len(dollars) == 0 || len(cents) != 2 {
		// not currency
		return 0
	}
	amount := bufToInt(dollars)*100 + bufToInt(cents)
	if isNegative {
		return -1 * amount
	}
	return amount
}

// date translates something like 12/31'13 to 2013/12/31
func bufToDate(b []byte) string {
	// 9/ 3'16 -> 2016/09/03
	// 9/13'16 -> 2016/09/13
	if len(b) == 7 && b[1] == '/' && b[4] == '\'' {
		mm, dd, yy := b[0:1], b[2:4], b[5:]
		return fmt.Sprintf("%4d/%02d/%02d", bufToInt(yy)+2000, bufToInt(mm), bufToInt(dd))
	}

	// 12/ 9'16 -> 2016/12/09
	// 12/19'16 -> 2016/12/19
	if len(b) == 8 && b[2] == '/' && b[5] == '\'' {
		mm, dd, yy := b[0:2], b[3:6], b[6:]
		return fmt.Sprintf("%4d/%02d/%02d", bufToInt(yy)+2000, bufToInt(mm), bufToInt(dd))
	}

	// invalid date
	return "****/**/**"
}

func bufToInt(b []byte) (i int) {
	for pos := 0; pos < len(b); pos++ {
		if '0' <= b[pos] && b[pos] <= '9' {
			i = i*10 + int(b[pos]) - '0'
		}
	}
	return i
}
