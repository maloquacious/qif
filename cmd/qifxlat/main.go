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

// Package main implements a command line tool to convert QIF data to CSV.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/maloquacious/qif"
	"github.com/maloquacious/qif/reader"
	"github.com/maloquacious/qif/scanner"
	cdata "github.com/maloquacious/qif/writer/csv"
	jdata "github.com/maloquacious/qif/writer/json"
	ldata "github.com/maloquacious/qif/writer/ledger"
)

func main() {
	// until the configuration is read, errors are logged with the defaults
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config()
	if err != nil {
		logger.Error("qifxlat: configuration", "err", err)
		os.Exit(2)
	}

	if cfg.Show.Version {
		fmt.Println(qif.Version().Short())
		return
	}

	if logger, err = newLogger(os.Stderr, cfg.Log.Level, cfg.Log.Format); err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("qifxlat: configuration", "err", err)
		os.Exit(2)
	}

	if err = run(cfg, logger); err != nil {
		logger.Error("qifxlat: failed", "err", err)
		os.Exit(2)
	}
}

func run(cfg *Config, logger *slog.Logger) error {
	started := time.Now()

	logger.Debug("qifxlat: settings",
		"version", qif.Version().String(),
		"input", cfg.Input.QIF,
		"csv", cfg.Output.CSV,
		"json", cfg.Output.JSON,
		"ledger", cfg.Output.Ledger,
		"log_level", cfg.Log.Level,
		"log_format", cfg.Log.Format)
	if cfg.Output.CSV == "" && cfg.Output.JSON == "" && cfg.Output.Ledger == "" {
		logger.Warn("qifxlat: no output files specified; validating the QIF data only")
	}

	input, err := os.ReadFile(cfg.Input.QIF)
	if err != nil {
		return err
	}

	sc, err := scanner.New(input)
	if err != nil {
		return err
	}

	r, err := reader.Read(sc)
	if err != nil {
		return err
	}

	// Accounts, Categories, Securities and Tags are nil when the file lacks the section
	var accounts, categories, securities, tags int
	if r.Accounts != nil {
		accounts = len(r.Accounts.Records)
	}
	if r.Categories != nil {
		categories = len(r.Categories.Records)
	}
	if r.Securities != nil {
		securities = len(r.Securities.Records)
	}
	if r.Tags != nil {
		tags = len(r.Tags.Records)
	}
	logger.Info("import: read complete",
		"accounts", accounts,
		"categories", categories,
		"memorized", len(r.Memorized),
		"prices", len(r.Prices),
		"securities", securities,
		"tags", tags,
		"transactions", len(r.Transactions))
	logger.Debug("import: finished", "duration", time.Since(started))

	if cfg.Output.CSV != "" {
		started := time.Now()
		data, err := cdata.Translate(r, logger)
		if err != nil {
			return err
		}
		if err = writeFile(cfg.Output.CSV, data.Write); err != nil {
			return err
		}
		logger.Debug("csv: finished", "duration", time.Since(started))
	}

	if cfg.Output.JSON != "" {
		started := time.Now()
		data, err := jdata.Translate(r, logger)
		if err != nil {
			return err
		}
		if err = writeFile(cfg.Output.JSON, data.Write); err != nil {
			return err
		}
		logger.Debug("json: finished", "duration", time.Since(started))
	}

	if cfg.Output.Ledger != "" {
		started := time.Now()
		data, err := ldata.Translate(r, logger)
		if err != nil {
			return err
		}
		if err = writeFile(cfg.Output.Ledger, data.Write); err != nil {
			return err
		}
		logger.Debug("ledger: finished", "duration", time.Since(started))
	}

	logger.Info("qifxlat: run complete", "duration", time.Since(started))

	return nil
}

// writeFile calls write to produce the contents of the file at path.
// It writes to a temporary file in the same directory and renames it
// onto path only after write succeeds, so an error never leaves a
// partial or empty file and never replaces an existing one.
func writeFile(path string, write func(io.Writer) error) error {
	fp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := fp.Name()
	if err = write(fp); err == nil {
		err = fp.Chmod(0644)
	}
	if err != nil {
		_ = fp.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err = fp.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
