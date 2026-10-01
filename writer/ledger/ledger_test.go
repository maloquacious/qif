/*
 * qif - a package to convert QIF data
 *
 * Copyright (c) 2026 Michael D Henderson
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
	"bytes"
	"strings"
	"testing"
)

// TestWriteCountsSkipped verifies that entries with no amounts are counted
// as skipped and are not written.
func TestWriteCountsSkipped(t *testing.T) {
	l := &LEDGER{Entries: []*Entry{
		{Line: 1, Date: "2026/01/01", Payee: "Kept One", Account: "Checking",
			Lines: Lines{{Line: 2, Category: "Groceries", Amount: "10.00"}}},
		{Line: 3, Date: "2026/01/02", Payee: "Zero Entry", Account: "Checking", IsZero: true,
			Lines: Lines{{Line: 4, Category: "Groceries", Amount: "0.00", IsZero: true}}},
		{Line: 5, Date: "2026/01/03", Payee: "Kept Two", Account: "Checking",
			Lines: Lines{{Line: 6, Category: "Dining", Amount: "20.00"}}},
	}}

	var buf bytes.Buffer
	skipped, written, err := l.write(&buf)
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}
	if skipped != 1 {
		t.Errorf("skipped: got %d, want 1", skipped)
	}
	if written != 2 {
		t.Errorf("written: got %d, want 2", written)
	}

	out := buf.String()
	for _, want := range []string{"Kept One", "Kept Two"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Zero Entry") {
		t.Errorf("output contains skipped entry:\n%s", out)
	}
}
