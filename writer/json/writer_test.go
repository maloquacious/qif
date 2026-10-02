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

package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/reader/account"
	"github.com/maloquacious/qif/reader/category"
	"github.com/maloquacious/qif/writer/json"
)

// TestTranslateNilSections verifies that Translate does not dereference
// sections that the reader left nil because the file did not contain them.
func TestTranslateNilSections(t *testing.T) {
	for _, tc := range []struct {
		name string
		r    *reader.Reader
	}{
		{"empty", &reader.Reader{}},
		{"categories only", &reader.Reader{
			Categories: &category.Section{Records: []*category.Record{{Name: "Groceries"}}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("Translate panicked: %v", p)
				}
			}()
			if _, err := json.Translate(tc.r, nil); err != nil {
				t.Fatalf("Translate returned error: %v", err)
			}
		})
	}
}

// TestTranslateUnknownAccountType verifies that an unknown account type is
// returned as an error instead of a panic (#12).
func TestTranslateUnknownAccountType(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 12, Name: "Foo", Type: "Mutual"}}},
	}
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Translate panicked: %v", p)
			}
		}()
		_, err = json.Translate(r, nil)
	}()
	if err == nil {
		t.Fatal("Translate: expected error, got nil")
	}
	if want := `12: account "Foo": unknown account type "Mutual"`; err.Error() != want {
		t.Errorf("Translate: want error %q, got %q", want, err)
	}
}

// TestTranslateInvestmentAccount verifies that an Invst account is
// translated (#12).
func TestTranslateInvestmentAccount(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 3, Name: "Broker", Type: "Invst"}}},
	}
	var got *json.JSON
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("Translate panicked: %v", p)
			}
		}()
		got, err = json.Translate(r, nil)
	}()
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if want := "investment"; len(got.Accounts) != 1 || got.Accounts[0].Type != want {
		t.Errorf("Translate: want one account of type %q, got %+v", want, got.Accounts)
	}
}

// TestWriteLogs is a regression test for issue #22: Write logs its counts
// to the logger given to Translate instead of printing them.
func TestWriteLogs(t *testing.T) {
	r := &reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "Checking", Type: "Bank"}}},
	}
	var log bytes.Buffer
	j, err := json.Translate(r, slog.New(slog.NewTextHandler(&log, nil)))
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if err := j.Write(io.Discard); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if want := `level=INFO msg="json: write complete" accounts=1 categories=0 transactions=0`; !strings.Contains(log.String(), want) {
		t.Errorf("log: want %q, got %q", want, log.String())
	}
	// the logger is not part of the JSON output
	var out bytes.Buffer
	if err := j.Write(&out); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if strings.Contains(out.String(), "logger") {
		t.Errorf("output contains the logger: %s", out.String())
	}
}

// TestWriteValidJSON is a regression test for issue #21: Write produces one
// JSON document ending with a newline.
func TestWriteValidJSON(t *testing.T) {
	j, err := json.Translate(&reader.Reader{
		Accounts: &account.Section{Records: []*account.Record{{Line: 2, Name: "Checking", Type: "Bank"}}},
	}, nil)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	var out bytes.Buffer
	if err := j.Write(&out); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if !strings.HasSuffix(out.String(), "}\n") {
		t.Errorf("output should end with a newline: %q", out.String())
	}
	var got map[string]any
	if err := stdjson.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if _, ok := got["accounts"]; !ok {
		t.Errorf("output missing accounts: %s", out.String())
	}
}
