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
package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewLogger is a regression test for issue #22: the logger honors the
// configured level and format.
func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger, err := newLogger(&buf, "warn", "json")
	if err != nil {
		t.Fatalf("newLogger: unexpected error: %v", err)
	}
	logger.Info("hidden")
	logger.Warn("shown", "count", 3)
	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Errorf("level warn: info record was written: %s", out)
	}
	if !strings.Contains(out, `"msg":"shown"`) || !strings.Contains(out, `"count":3`) {
		t.Errorf("format json: want a JSON warn record, got %s", out)
	}

	buf.Reset()
	if logger, err = newLogger(&buf, "debug", "text"); err != nil {
		t.Fatalf("newLogger: unexpected error: %v", err)
	}
	logger.Debug("detail", "count", 3)
	if out := buf.String(); !strings.Contains(out, "level=DEBUG msg=detail count=3") {
		t.Errorf("format text: want a text debug record, got %q", out)
	}
}

// TestNewLoggerInvalid is a regression test for issue #22: an unknown level
// or format is an error.
func TestNewLoggerInvalid(t *testing.T) {
	for _, tc := range []struct{ level, format string }{
		{"verbose", "text"},
		{"info", "xml"},
		{"", "text"},
	} {
		if _, err := newLogger(&bytes.Buffer{}, tc.level, tc.format); err == nil {
			t.Errorf("newLogger(%q, %q): want error, got nil", tc.level, tc.format)
		}
	}
}
