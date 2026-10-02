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

// Package scanner implements an immutable cursor over QIF input.
package scanner

import (
	"bytes"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/maloquacious/qif/stdlib"
)

// Scanner is an immutable cursor over the input.
//
// Line and Col give the position of the start of Buffer. Both are 1-based:
// the first character of every line is in column 1. Col counts runes, but
// Date, Field and Literal advance it by the byte length of their flag, so
// flags and literals must be ASCII.
type Scanner struct {
	Line   int
	Col    int
	Buffer []byte
}

// New returns a scanner positioned at the start of a copy of the input.
// It removes every carriage return ("\r"), appends a final newline if the
// input doesn't end with one, and returns an error for invalid UTF-8.
func New(input []byte) (Scanner, error) {
	b, offset, line, col := make([]byte, 0, len(input)+1), 0, 1, 1
	for offset < len(input) {
		r, w := utf8.DecodeRune(input[offset:])
		if r == utf8.RuneError {
			return Scanner{}, fmt.Errorf("utf8: import: invalid utf-8 character on line %d, col %d", line, col)
		} else if r == '\r' {
			offset += w
			continue
		}
		b, offset = append(b, input[offset:offset+w]...), offset+w
		if r == '\n' {
			line, col = line+1, 1
		} else {
			col++
		}
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	return Scanner{Buffer: b, Line: 1, Col: 1}, nil
}

// Date accepts the flag followed by a QIF date, which looks like
//
//	digit digit? slash (space | digit) digit tic digit digit
//
// and consumes the rest of the line. The lexeme is the date as yyyy/mm/dd
// (see stdlib.Date), not the original text. An invalid date, such as month
// 13, is no match.
func (buf Scanner) Date(flag string) ([]byte, Scanner) {
	saved := buf

	if !bytes.HasPrefix(buf.Buffer, []byte(flag)) {
		return nil, buf
	}
	// skip the flag (we don't return it as part of the lexeme)
	buf.Buffer, buf.Col = buf.Buffer[len(flag):], buf.Col+len(flag)

	var length, w int
	var r rune

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !unicode.IsDigit(r) { // digit
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); unicode.IsDigit(r) { // digit?
		length += w
	}

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); r != '/' { // slash
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !(r == ' ' || unicode.IsDigit(r)) { // (space | digit)
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !unicode.IsDigit(r) { // digit
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); r != '\'' { // tic
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !unicode.IsDigit(r) { // digit
		return nil, saved
	}
	length += w

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !unicode.IsDigit(r) { // digit
		return nil, saved
	}
	length += w

	lexeme := stdlib.Date(buf.Buffer[:length])
	if lexeme == "****/**/**" {
		return nil, saved
	}

	// consume to the end of the line
	_, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return []byte(lexeme), buf
}

// Field accepts the flag and consumes the rest of the line. The lexeme is
// the text after the flag, without the newline; it is empty, not nil, when
// nothing follows the flag.
func (buf Scanner) Field(flag string) ([]byte, Scanner) {
	if !bytes.HasPrefix(buf.Buffer, []byte(flag)) {
		return nil, buf
	}
	// skip the flag (we don't return it as part of the lexeme)
	buf.Buffer, buf.Col = buf.Buffer[len(flag):], buf.Col+len(flag)

	// read the lexeme and consume to the end of the line
	var lexeme []byte
	lexeme, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return lexeme, buf
}

// EndOfLine accepts a newline. It also accepts "\r\n", but New removes
// every "\r", so only a Scanner built some other way can contain one.
func (buf Scanner) EndOfLine() ([]byte, Scanner) {
	if len(buf.Buffer) == 0 {
		return nil, buf
	}
	if buf.Buffer[0] == '\n' {
		lexeme := bdup([]byte{'\n'})
		buf.Buffer, buf.Line, buf.Col = buf.Buffer[1:], buf.Line+1, 1
		// return the lexeme and updated buffer
		return lexeme, buf
	}
	if len(buf.Buffer) > 1 && buf.Buffer[0] == '\r' && buf.Buffer[1] == '\n' {
		lexeme := bdup([]byte{'\n'})
		buf.Buffer, buf.Line, buf.Col = buf.Buffer[2:], buf.Line+1, 1
		// return the lexeme and updated buffer
		return lexeme, buf
	}
	return nil, buf
}

// EndOfRecord accepts the record terminator '^' and consumes the rest of
// its line. At the end of the input it returns "^" without consuming
// anything.
func (buf Scanner) EndOfRecord() ([]byte, Scanner) {
	if len(buf.Buffer) == 0 {
		return bdup([]byte{'^'}), buf
	} else if buf.Buffer[0] != '^' {
		return nil, buf
	}

	lexeme := bdup(buf.Buffer[:1])

	// consume to the end of the line
	_, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return lexeme, buf
}

// EndOfSection accepts the '!' that starts the next section header, or
// the end of the input. It never consumes anything: the returned scanner
// is the receiver.
func (buf Scanner) EndOfSection() ([]byte, Scanner) {
	if len(buf.Buffer) == 0 {
		return bdup([]byte{'!'}), buf
	} else if buf.Buffer[0] != '!' {
		return nil, buf
	}
	lexeme := bdup(buf.Buffer[:1])
	// return the lexeme and original buffer
	return lexeme, buf
}

// Literal accepts a literal and consumes the rest of the line.
// On a match, the lexeme is the text between the literal and the end of the line;
// it is never nil, even when empty. On no match, the lexeme is nil.
// The literal must not contain a newline, since Line and Col are not adjusted for one.
func (buf Scanner) Literal(lit string) ([]byte, Scanner) {
	if !bytes.HasPrefix(buf.Buffer, []byte(lit)) {
		return nil, buf
	}
	// skip the literal (we don't return it as part of the lexeme)
	buf.Buffer, buf.Col = buf.Buffer[len(lit):], buf.Col+len(lit)

	// read the lexeme and consume to the end of the line
	var lexeme []byte
	lexeme, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return lexeme, buf
}

// ToEndOfLine consumes the text up to and including the next newline. The
// lexeme is the text before the newline; it is never nil.
func (buf Scanner) ToEndOfLine() ([]byte, Scanner) {
	// consume to the end of the line
	var length int
	r, w := utf8.DecodeRune(buf.Buffer[length:])
	for r != utf8.RuneError && r != '\n' {
		length, buf.Col = length+w, buf.Col+1
		r, w = utf8.DecodeRune(buf.Buffer[length:])
	}

	lexeme := bdup(buf.Buffer[:length])

	if r == '\n' {
		buf.Line, buf.Col = buf.Line+1, 1
		length++
	}
	buf.Buffer = buf.Buffer[length:]

	// return the lexeme and updated buffer
	return lexeme, buf
}

func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
