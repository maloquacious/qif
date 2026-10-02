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

// Package stdlib holds small helpers shared by the other packages.
package stdlib

import (
	"fmt"
)

// Date translates QIF date to a string with the date formatted as yyyy/mm/dd.
// The QIF date is a string which looks like
//
//	digit digit? slash (space | digit) digit tic digit digit
//
// It is is formatted as mm/dd'yy. The month can be one or two digits
// (eg, January is `1` while October is `10`). The day must be two characters,
// but the first character may be a space instead of a zero. For example,
// `01` and ` 1` are both the first day of the month. The year must be two
// digits, and we're assuming it is always in the 21st century (eg, `16` is
// converted to 2016, not 1916).
func Date(b []byte) string {
	atoi := func(b []byte) int {
		return int(b[0] - '0')
	}
	isdigit := func(b []byte) bool {
		return len(b) != 0 && '0' <= b[0] && b[0] <= '9'
	}

	var cc, yy, mm, dd int

	if isdigit(b) { // digit
		b, mm = b[1:], atoi(b)
	} else {
		return "****/**/**"
	}

	if isdigit(b) { // digit?
		b, mm = b[1:], mm*10+atoi(b)
	}

	if len(b) > 0 && b[0] == '/' { // slash
		b = b[1:]
	} else {
		return "****/**/**"
	}

	// space | digit
	if len(b) > 0 && b[0] == ' ' { // space
		b = b[1:]
	} else if isdigit(b) { // digit
		b, dd = b[1:], atoi(b)
	} else {
		return "****/**/**"
	}

	if isdigit(b) { // digit
		b, dd = b[1:], dd*10+atoi(b)
	} else {
		return "****/**/**"
	}

	if len(b) > 0 && b[0] == '\'' { // tic
		b, cc = b[1:], 2000
	} else {
		return "****/**/**"
	}

	if isdigit(b) { // digit
		b, yy = b[1:], atoi(b)
	} else {
		return "****/**/**"
	}

	if isdigit(b) { // digit
		yy = yy*10 + atoi(b)
	} else {
		return "****/**/**"
	}

	yyyy := cc + yy
	switch mm {
	case 1, 3, 5, 7, 8, 10, 12: // 31 days
		if !(1 <= dd && dd <= 31) {
			return "****/**/**"
		}
	case 4, 6, 9, 11: // 30 days
		if !(1 <= dd && dd <= 30) {
			return "****/**/**"
		}
	case 2: // february
		if yyyy%4 == 0 { // leap year (years are always 2000-2099)
			if !(1 <= dd && dd <= 29) {
				return "****/**/**"
			}
		} else {
			if !(1 <= dd && dd <= 28) {
				return "****/**/**"
			}
		}
	default: // invalid month
		return "****/**/**"
	}

	return fmt.Sprintf("%4d/%02d/%02d", yyyy, mm, dd)
}

// FlipSign changes the sign of an amount
func FlipSign(amount string) string {
	if amount == "" || amount == "0.00" {
		return "0.00"
	} else if amount[0] == '-' {
		return amount[1:]
	} else if amount[0] == '+' {
		return "-" + amount[1:]
	}
	return "-" + amount
}
