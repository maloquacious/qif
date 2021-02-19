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
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"os"
)

type Config struct {
	Input struct {
		QIF string
	}
	Output struct {
		CSV    string
		JSON   string
		Ledger string
	}
	Show struct {
		Timing bool
	}
}

func config() (*Config, error) {
	cfg := Config{}
	cfg.Show.Timing = true

	fs := flag.NewFlagSet("qifxlat", flag.ExitOnError)
	fs.StringVar(&cfg.Input.QIF, "input", "", "QIF file to translate")
	fs.StringVar(&cfg.Output.CSV, "output-csv-filename", cfg.Output.CSV, "file to write CSV data to")
	fs.StringVar(&cfg.Output.JSON, "output-json-filename", cfg.Output.JSON, "file to write JSON data to")
	fs.StringVar(&cfg.Output.Ledger, "output-ledger-filename", cfg.Output.Ledger, "file to write Ledger data to")
	fs.BoolVar(&cfg.Show.Timing, "show-timing", cfg.Show.Timing, "display timing of stages")
	_ = fs.String("config", "", "config file (optional)")

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("QIFXLAT"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.PlainParser)); err != nil {
		return nil, err
	}

	if cfg.Input.QIF == "" {
		return nil, fmt.Errorf("please provide the name of the QIF file to translate\n")
	}
	fmt.Printf("%-30s == %q\n", "QIFXLAT_INPUT", cfg.Input.QIF)
	outputFileSpecified := false
	if cfg.Output.CSV != "" {
		fmt.Printf("%-30s == %q\n", "QIFXLAT_OUTPUT_CSV_FILENAME", cfg.Output.CSV)
		outputFileSpecified = true
	}
	if cfg.Output.JSON != "" {
		fmt.Printf("%-30s == %q\n", "QIFXLAT_OUTPUT_JSON_FILENAME", cfg.Output.JSON)
		outputFileSpecified = true
	}
	if cfg.Output.Ledger != "" {
		fmt.Printf("%-30s == %q\n", "QIFXLAT_OUTPUT_LEDGER_FILENAME", cfg.Output.Ledger)
		outputFileSpecified = true
	}
	if !outputFileSpecified {
		fmt.Printf("warning: no output file(s) specified; will validate QIF data only\n")
	}
	if cfg.Show.Timing {
		fmt.Printf("%-30s == %v\n", "QIFXLAT_SHOW_TIMING", cfg.Show.Timing)
	}

	return &cfg, nil
}
