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

package ledger

import (
	"fmt"
	"io"
	"strings"
)

type Lines []*Line

type Line struct {
	Line     int
	Source   string
	Category string
	Amount   string
	IsZero   bool
}

func (l Lines) Len() int {
	return len(l)
}

func (l Lines) Less(i, j int) bool {
	return l[i].Line < l[j].Line
}

func (l Lines) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *Line) Write(w io.Writer) error {
	category := l.Category
	if strings.Index(category, "  ") != -1 || strings.HasPrefix(category, "check") {
		category = strings.ReplaceAll(category, " ", "_")
	}
	_, err := fmt.Fprintf(w, "    %-49s  %15s ;; %6d %s\n", category, "$"+l.Amount, l.Line, l.Source)
	return err
}
