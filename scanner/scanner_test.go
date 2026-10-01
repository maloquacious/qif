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

package scanner_test

import (
	"testing"

	"github.com/maloquacious/qif/scanner"
)

func TestDateRejectsInvalidMonth(t *testing.T) {
	// When the date flag is followed by an invalid month
	// Then Date does not match and the scanner is not advanced
	input := "D13/01'16\n"
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("input of %q: unexpected error %v\n", input, err)
	}
	date, rest := sc.Date("D")
	if date != nil {
		t.Errorf("input of %q yields %q: expected no match\n", input, date)
	}
	if string(rest.Buffer) != input {
		t.Errorf("input of %q advanced the scanner to %q\n", input, rest.Buffer)
	}

	// When the date flag is followed by a valid date
	// Then Date returns the date formatted as yyyy/mm/dd
	input, expected := "D12/31'16\n", "2016/12/31"
	sc, err = scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("input of %q: unexpected error %v\n", input, err)
	}
	if date, _ = sc.Date("D"); string(date) != expected {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, date, expected)
	}
}

func TestLiteralReturnsRestOfLine(t *testing.T) {
	// When the literal is followed by text and more lines
	// Then Literal returns only the rest of the line and advances past it
	input := "!Type:Bank extra\nD12/31'16\n^\n"
	sc, err := scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("input of %q: unexpected error %v\n", input, err)
	}
	lexeme, rest := sc.Literal("!Type:Bank")
	if expected := " extra"; string(lexeme) != expected {
		t.Errorf("input of %q yields %q: expected value is %q\n", input, lexeme, expected)
	}
	if expected := "D12/31'16\n^\n"; string(rest.Buffer) != expected {
		t.Errorf("input of %q advanced the scanner to %q: expected %q\n", input, rest.Buffer, expected)
	}
	if rest.Line != 2 {
		t.Errorf("input of %q: line is %d: expected 2\n", input, rest.Line)
	}

	// When the literal is the whole line
	// Then Literal returns an empty, non-nil lexeme and advances past the line
	input = "!Account\nNChecking\n"
	sc, err = scanner.New([]byte(input))
	if err != nil {
		t.Fatalf("input of %q: unexpected error %v\n", input, err)
	}
	lexeme, rest = sc.Literal("!Account")
	if lexeme == nil || len(lexeme) != 0 {
		t.Errorf("input of %q yields %#v: expected empty, non-nil lexeme\n", input, lexeme)
	}
	if expected := "NChecking\n"; string(rest.Buffer) != expected {
		t.Errorf("input of %q advanced the scanner to %q: expected %q\n", input, rest.Buffer, expected)
	}

	// When the literal does not match
	// Then Literal returns nil and the scanner is not advanced
	lexeme, rest = sc.Literal("!Type:Cat")
	if lexeme != nil {
		t.Errorf("input of %q yields %q: expected no match\n", input, lexeme)
	}
	if string(rest.Buffer) != input {
		t.Errorf("input of %q advanced the scanner to %q\n", input, rest.Buffer)
	}
}
