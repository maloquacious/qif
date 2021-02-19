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

package scanner

import (
	"bytes"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Scanner
type Scanner struct {
	Line   int
	Col    int
	Buffer []byte
}

// New returns a new scanner with a copy of the input.
func New(input []byte) (Scanner, error) {
	b, offset, line, col := make([]byte, 0, len(input)+1), 0, 1, 0
	for offset < len(input) {
		r, w := utf8.DecodeRune(input[offset:])
		if r == utf8.RuneError {
			return Scanner{}, fmt.Errorf("utf8: import: invalid utf-8 character on line %d, col %d", line, col)
		} else if r == '\r' {
			offset += w
			continue
		} else if r == '\n' {
			line, col = line+1, 0
		}
		b, offset, col = append(b, input[offset:offset+w]...), offset+w, col+1
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	return Scanner{Buffer: b, Line: 1}, nil
}

// Date will accept a date string which looks like
//    digit digit? slash (space digit) digit tic digit digit
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

	if r, w = utf8.DecodeRune(buf.Buffer[length:]); !(r == ' ' || unicode.IsDigit(r)) { // (space digit)
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

	lexeme := bdate(buf.Buffer[:length])
	if lexeme == "****/**/**" {
		return nil, saved
	}

	// consume to the end of the line
	_, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return []byte(lexeme), buf
}

// Field will accept text to the end of the line only if the flag matches.
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

// EndOfLine will accept \r\n and \n.
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

// EndOfRecord will accept '^' or end-of-input.
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

// EndOfSection will accept '!' or end-of-input.
// It does not actually consume the marker.
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

// Literal will accept a literal.
// The buffer's line and col variables will be hosed if the literal has an embedded newline.
func (buf Scanner) Literal(lit string) ([]byte, Scanner) {
	if !bytes.HasPrefix(buf.Buffer, []byte(lit)) {
		return nil, buf
	}

	lexeme := bdup(buf.Buffer[len(lit):])

	// consume to the end of the line
	_, buf = buf.ToEndOfLine()

	// return the lexeme and updated buffer
	return lexeme, buf
}

// ToEndOfLine will consume all the text up to (and including) the next end of line.
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

// bdate translates QIF date to a string with the date formatted as yyyy/mm/dd
// The QIF date is formatted as mm/dd'yy. The month can be one or two digits
// (eg, January is `1` while October is `10`). The day must be two characters,
// but the first character may be a space instead of a zero. For example,
// `01` and ` 1` are both the first day of the month. The year must be two
// digits, and we're assuming it is always in the 21st century (eg, `16` is
// converted to 2016, not 1916).
func bdate(b []byte) string {
	// 9/ 3'16 -> 2016/09/03
	// 9/13'16 -> 2016/09/13
	if len(b) == 7 && b[1] == '/' && b[4] == '\'' {
		mm, dd, yy := b[0:1], b[2:4], b[5:]
		return fmt.Sprintf("%4d/%02d/%02d", bint(yy)+2000, bint(mm), bint(dd))
	}

	// 12/ 9'16 -> 2016/12/09
	// 12/19'16 -> 2016/12/19
	if len(b) == 8 && b[2] == '/' && b[5] == '\'' {
		mm, dd, yy := b[0:2], b[3:6], b[6:]
		return fmt.Sprintf("%4d/%02d/%02d", bint(yy)+2000, bint(mm), bint(dd))
	}

	// invalid date
	return "****/**/**"
}

func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// bint converts a slice to an int.
func bint(b []byte) (i int) {
	for pos := 0; pos < len(b); pos++ {
		if '0' <= b[pos] && b[pos] <= '9' {
			i = i*10 + int(b[pos]) - '0'
		}
	}
	return i
}
