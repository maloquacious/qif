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

// env returns a getenv function that looks up names in m.
func env(m map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		v, ok := m[name]
		return v, ok
	}
}

// TestEnvName is a regression test for issue #34: a flag's environment
// variable is QIFXLAT_ plus the flag name uppercased, with "-" as "_".
func TestEnvName(t *testing.T) {
	for flagName, want := range map[string]string{
		"input":                  "QIFXLAT_INPUT",
		"output-csv-filename":    "QIFXLAT_OUTPUT_CSV_FILENAME",
		"output-json-filename":   "QIFXLAT_OUTPUT_JSON_FILENAME",
		"output-ledger-filename": "QIFXLAT_OUTPUT_LEDGER_FILENAME",
		"log-level":              "QIFXLAT_LOG_LEVEL",
		"log-format":             "QIFXLAT_LOG_FORMAT",
		"version":                "QIFXLAT_VERSION",
	} {
		if got := envName(flagName); got != want {
			t.Errorf("envName(%q): want %q, got %q", flagName, want, got)
		}
	}
}

// TestParseConfigEnv is a regression test for issue #34: every setting can
// come from the environment.
func TestParseConfigEnv(t *testing.T) {
	cfg, err := parseConfig(nil, env(map[string]string{
		"QIFXLAT_INPUT":                  "f.qif",
		"QIFXLAT_OUTPUT_CSV_FILENAME":    "out.csv",
		"QIFXLAT_OUTPUT_JSON_FILENAME":   "out.json",
		"QIFXLAT_OUTPUT_LEDGER_FILENAME": "out.ledger",
		"QIFXLAT_LOG_LEVEL":              "debug",
		"QIFXLAT_LOG_FORMAT":             "json",
	}))
	if err != nil {
		t.Fatalf("parseConfig: unexpected error: %v", err)
	}
	for _, tc := range []struct{ name, got, want string }{
		{"input", cfg.Input.QIF, "f.qif"},
		{"csv", cfg.Output.CSV, "out.csv"},
		{"json", cfg.Output.JSON, "out.json"},
		{"ledger", cfg.Output.Ledger, "out.ledger"},
		{"log level", cfg.Log.Level, "debug"},
		{"log format", cfg.Log.Format, "json"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: want %q, got %q", tc.name, tc.want, tc.got)
		}
	}

	// -version from the environment skips the -input check
	cfg, err = parseConfig(nil, env(map[string]string{"QIFXLAT_VERSION": "true"}))
	if err != nil {
		t.Fatalf("parseConfig: QIFXLAT_VERSION: unexpected error: %v", err)
	}
	if !cfg.Show.Version {
		t.Errorf("QIFXLAT_VERSION=true: want Show.Version true")
	}
}

// TestParseConfigPrecedence is a regression test for issue #34: a flag on
// the command line overrides its environment variable, which overrides the
// default.
func TestParseConfigPrecedence(t *testing.T) {
	cfg, err := parseConfig(
		[]string{"-input", "flag.qif", "-log-level", "warn"},
		env(map[string]string{
			"QIFXLAT_INPUT":                "env.qif",
			"QIFXLAT_OUTPUT_JSON_FILENAME": "env.json",
		}))
	if err != nil {
		t.Fatalf("parseConfig: unexpected error: %v", err)
	}
	if cfg.Input.QIF != "flag.qif" {
		t.Errorf("input: want flag to win, got %q", cfg.Input.QIF)
	}
	if cfg.Output.JSON != "env.json" {
		t.Errorf("json: want env value, got %q", cfg.Output.JSON)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("log level: want flag value, got %q", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("log format: want default, got %q", cfg.Log.Format)
	}

	// -version=false on the command line overrides QIFXLAT_VERSION
	_, err = parseConfig([]string{"-version=false"}, env(map[string]string{"QIFXLAT_VERSION": "1"}))
	if err == nil {
		t.Errorf("-version=false: want missing input error, got nil")
	}
}

// TestParseConfigInvalidEnv is a regression test for issue #34: an invalid
// environment value is an error that names the variable.
func TestParseConfigInvalidEnv(t *testing.T) {
	_, err := parseConfig([]string{"-input", "f.qif"}, env(map[string]string{"QIFXLAT_VERSION": "maybe"}))
	if err == nil {
		t.Fatalf("QIFXLAT_VERSION=maybe: want error, got nil")
	}
	if !strings.Contains(err.Error(), "QIFXLAT_VERSION") {
		t.Errorf("error should name the variable, got %q", err)
	}
}

// TestParseConfigMissingInput checks that -input is required unless
// -version is set.
func TestParseConfigMissingInput(t *testing.T) {
	if _, err := parseConfig(nil, env(nil)); err == nil {
		t.Errorf("no input: want error, got nil")
	}
	if _, err := parseConfig([]string{"-version"}, env(nil)); err != nil {
		t.Errorf("-version: unexpected error: %v", err)
	}
}
