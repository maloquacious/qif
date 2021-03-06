/*
 * qif - a package to convert QIF data
 *
 * Copyright (c) 2021 Michael D Henderson
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 */

package stdlib_test

import (
	"github.com/mdhender/qif/stdlib"
	"testing"
)

func TestDate(t *testing.T) {
	// Specification: Date

	// When "9/ 3'16" is converted
	// Then it has the value "2016/09/03"
	input, expected := "9/ 3'16", "2016/09/03"
	yields := stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "9/ 3'16" is converted
	// Then it has the value "2016/09/13"
	input, expected = "9/13'16", "2016/09/13"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "12/ 9'16" is converted
	// Then it has the value "2016/12/09"
	input, expected = "12/ 9'16", "2016/12/09"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "12/19'16" is converted
	// Then it has the value "2016/12/19"
	input, expected = "12/19'16", "2016/12/19"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "ab/cd'ee" is converted
	// Then it has the value "****/**/**"
	input, expected = "ab/cd'ee", "****/**/**"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "1/32'00" is converted
	// Then it has the value "****/**/**"
	input, expected = "1/32'00", "****/**/**"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "2/29'00" is converted
	// Then it has the value "****/**/**"
	input, expected = "2/29'00", "****/**/**"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "2/29'01" is converted
	// Then it has the value "****/**/**"
	input, expected = "2/29'01", "****/**/**"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "2/29'04" is converted
	// Then it has the value "2004/02/29"
	input, expected = "2/29'04", "2004/02/29"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}

	// When "11/31'00" is converted
	// Then it has the value "****/**/**"
	input, expected = "11/31'00", "****/**/**"
	yields = stdlib.Date([]byte(input))
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, yields, expected)
	}
}

func TestFlipSign(t *testing.T) {
	// Specification: Amounts

	// When "0.00" is flipped
	// Then it has the value "0.00"
	amount, expected := "0.00", "0.00"
	yields := stdlib.FlipSign(amount)
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", amount, yields, expected)
	}

	// When "12.45" is flipped
	// Then it has the value "-12.45"
	amount, expected = "12.45", "-12.45"
	yields = stdlib.FlipSign(amount)
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", amount, yields, expected)
	}

	// When "-12.45" is flipped
	// Then it has the value "12.45"
	amount, expected = "-12.45", "12.45"
	yields = stdlib.FlipSign(amount)
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", amount, yields, expected)
	}

	// When "" is flipped
	// Then it has the value "0.00"
	amount, expected = "", "0.00"
	yields = stdlib.FlipSign(amount)
	if expected != yields {
		t.Errorf("input of %q yields %q: expected value is %q\n", amount, yields, expected)
	}
}
