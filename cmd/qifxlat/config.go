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
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
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
	Log struct {
		Level  string
		Format string
	}
	Show struct {
		Version bool
	}
}

// envPrefix is prepended to a flag's name to form its environment variable.
const envPrefix = "QIFXLAT_"

// config reads the configuration from the command line and the environment.
func config() (*Config, error) {
	return parseConfig(os.Args[1:], os.LookupEnv)
}

// parseConfig reads the configuration from args and from the environment
// variables that getenv looks up. Each flag can also be set by the variable
// named by envName; a flag given in args overrides its variable.
func parseConfig(args []string, getenv func(string) (string, bool)) (*Config, error) {
	cfg := Config{}
	cfg.Log.Level, cfg.Log.Format = "info", "text"

	fs := flag.NewFlagSet("qifxlat", flag.ExitOnError)
	fs.StringVar(&cfg.Input.QIF, "input", "", "QIF file to translate")
	fs.StringVar(&cfg.Output.CSV, "output-csv-filename", cfg.Output.CSV, "file to write CSV data to")
	fs.StringVar(&cfg.Output.JSON, "output-json-filename", cfg.Output.JSON, "file to write JSON data to")
	fs.StringVar(&cfg.Output.Ledger, "output-ledger-filename", cfg.Output.Ledger, "file to write Ledger data to")
	fs.StringVar(&cfg.Log.Level, "log-level", cfg.Log.Level, "log level: debug, info, warn or error")
	fs.StringVar(&cfg.Log.Format, "log-format", cfg.Log.Format, "log format: text or json")
	fs.BoolVar(&cfg.Show.Version, "version", cfg.Show.Version, "display the version and exit")
	fs.Usage = func() { usage(fs) }

	// apply the environment first, so that the command line overrides it
	var err error
	fs.VisitAll(func(f *flag.Flag) {
		if err != nil {
			return
		}
		name := envName(f.Name)
		if value, ok := getenv(name); ok {
			if setErr := fs.Set(f.Name, value); setErr != nil {
				err = fmt.Errorf("%s=%q: %w", name, value, setErr)
			}
		}
	})
	if err != nil {
		return nil, err
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if cfg.Show.Version {
		return &cfg, nil
	}

	if cfg.Input.QIF == "" {
		return nil, errors.New("please provide the name of the QIF file to translate")
	}

	return &cfg, nil
}

// envName returns the environment variable for the flag: the flag name
// uppercased, with "-" replaced by "_", after envPrefix.
func envName(flagName string) string {
	return envPrefix + strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}

// usage writes the help text, listing each flag with its environment variable.
func usage(fs *flag.FlagSet) {
	w := fs.Output()
	fmt.Fprintf(w, "Usage of %s:\n", fs.Name())
	fs.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(w, "  -%s (env %s)\n    \t%s", f.Name, envName(f.Name), f.Usage)
		if f.DefValue != "" && f.DefValue != "false" {
			fmt.Fprintf(w, " (default %q)", f.DefValue)
		}
		fmt.Fprintln(w)
	})
	fmt.Fprintln(w, "A flag on the command line overrides its environment variable.")
}

// newLogger returns a logger that writes to w at the given level
// (debug, info, warn or error) in the given format (text or json).
func newLogger(w io.Writer, level, format string) (*slog.Logger, error) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		return nil, fmt.Errorf("log level %q: want debug, info, warn or error", level)
	}
	opts := &slog.HandlerOptions{Level: lvl}
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(w, opts)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(w, opts)), nil
	}
	return nil, fmt.Errorf("log format %q: want text or json", format)
}
