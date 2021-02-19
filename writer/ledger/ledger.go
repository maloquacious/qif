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
	"sort"
)

type LEDGER struct {
	Entries []*Entry
}

func (l *LEDGER) Len() int {
	return len(l.Entries)
}

func (l *LEDGER) Less(i, j int) bool {
	if l.Entries[i].Date < l.Entries[j].Date {
		return true
	}
	if l.Entries[i].Date > l.Entries[j].Date {
		return false
	}
	return l.Entries[i].Line < l.Entries[j].Line
}

func (l *LEDGER) Sort() {
	sort.Sort(l)
	for _, e := range l.Entries {
		e.Sort()
	}
}

func (l *LEDGER) Swap(i, j int) {
	l.Entries[i], l.Entries[j] = l.Entries[j], l.Entries[i]
}

func (l *LEDGER) Write(w io.Writer) error {
	var skipped, written int

	for _, e := range l.Entries {
		// don't write entries that are missing amounts
		if e.IsZero {
			continue
		}

		err := e.Write(w)
		if err != nil {
			return err
		}
		written++
	}

	fmt.Printf("ledger: skipped   %8d entries\n", skipped)
	fmt.Printf("ledger: wrote     %8d entries\n", written)

	return nil
}
