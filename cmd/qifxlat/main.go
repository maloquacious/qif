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
	"github.com/mdhender/qif/reader"
	"github.com/mdhender/qif/scanner"
	cdata "github.com/mdhender/qif/writer/csv"
	jdata "github.com/mdhender/qif/writer/json"
	ldata "github.com/mdhender/qif/writer/ledger"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	cfg, err := config()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(2)
	}

	if err = run(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(2)
	}
}

func run(cfg *Config) error {
	started := time.Now()

	input, err := ioutil.ReadFile(cfg.Input.QIF)
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

	var totalRecords int
	if r.Accounts != nil {
		fmt.Printf("import: read %8d accounts\n", len(r.Accounts.Records))
		totalRecords += len(r.Accounts.Records)
	}
	if r.Categories != nil {
		fmt.Printf("import: read %8d categories\n", len(r.Categories.Records))
		totalRecords += len(r.Categories.Records)
	}
	fmt.Printf("import: read %8d memorized\n", len(r.Memorized))
	totalRecords += len(r.Memorized)
	fmt.Printf("import: read %8d prices\n", len(r.Prices))
	totalRecords += len(r.Prices)
	if r.Securities != nil {
		fmt.Printf("import: read %8d securities\n", len(r.Securities.Records))
		totalRecords += len(r.Securities.Records)
	}
	if r.Tags != nil {
		fmt.Printf("import: read %8d tags\n", len(r.Tags.Records))
	}
	fmt.Printf("import: read %8d transactions\n", len(r.Transactions))
	totalRecords += len(r.Transactions)

	if cfg.Show.Timing {
		duration := time.Now().Sub(started)
		fmt.Printf("import: finished in %v\n", duration)
	}

	if cfg.Output.CSV != "" {
		started := time.Now()

		fp, err := os.Create(cfg.Output.CSV)
		if err != nil {
			return err
		}
		data, err := cdata.Translate(r)
		if err != nil {
			return err
		}
		err = data.Write(fp)
		if err != nil {
			return err
		}
		err = fp.Close()
		if err != nil {
			return err
		}

		if cfg.Show.Timing {
			duration := time.Now().Sub(started)
			fmt.Printf("csv: finished in %v\n", duration)
		}
	}

	if cfg.Output.JSON != "" {
		started := time.Now()

		fp, err := os.Create(cfg.Output.JSON)
		if err != nil {
			return err
		}
		data, err := jdata.Translate(r)
		if err != nil {
			return err
		}
		err = data.Write(fp)
		if err != nil {
			return err
		}
		err = fp.Close()
		if err != nil {
			return err
		}

		if cfg.Show.Timing {
			duration := time.Now().Sub(started)
			fmt.Printf("json: finished in %v\n", duration)
		}
	}

	if cfg.Output.Ledger != "" {
		started := time.Now()

		fp, err := os.Create(cfg.Output.Ledger)
		if err != nil {
			return err
		}
		data, err := ldata.Translate(r)
		if err != nil {
			return err
		}
		err = data.Write(fp)
		if err != nil {
			return err
		}
		err = fp.Close()
		if err != nil {
			return err
		}

		if cfg.Show.Timing {
			duration := time.Now().Sub(started)
			fmt.Printf("ledger: finished in %v\n", duration)
		}
	}

	if cfg.Show.Timing {
		duration := time.Now().Sub(started)
		fmt.Printf("qif: finished run  in %v\n", duration)
	}

	return nil
}
