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
	"io"
	"log/slog"
	"sort"
)

type LEDGER struct {
	Entries []*Entry
	logger  *slog.Logger
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
	skipped, written, err := l.write(w)
	if err != nil {
		return err
	}

	l.log().Info("ledger: write complete", "written", written, "skipped", skipped)

	return nil
}

// log returns the logger, or one that discards everything if there is none.
func (l *LEDGER) log() *slog.Logger {
	if l.logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return l.logger
}

// write writes the entries to w and returns the number of entries
// skipped and written.
func (l *LEDGER) write(w io.Writer) (skipped, written int, err error) {
	for _, e := range l.Entries {
		// don't write entries that are missing amounts or that only
		// duplicate transfers written by another entry
		if e.IsZero || e.IsLinked {
			skipped++
			continue
		}

		if err := e.Write(w); err != nil {
			return skipped, written, err
		}
		written++
	}

	return skipped, written, nil
}
