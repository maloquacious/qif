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
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteFile is a regression test for issue #19: writeFile writes the
// content produced by the write func to the target path.
func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	err := writeFile(path, func(w io.Writer) error {
		_, err := io.WriteString(w, "new content\n")
		return err
	})
	if err != nil {
		t.Fatalf("writeFile: unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != "new content\n" {
		t.Errorf("content: want %q, got %q", "new content\n", got)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if mode := fi.Mode().Perm(); mode != 0644 {
		t.Errorf("mode: want %v, got %v", os.FileMode(0644), mode)
	}
	assertOnlyFiles(t, dir, "out.csv")
}

// TestWriteFileError is a regression test for issue #19: when the write func
// fails, writeFile leaves neither the target nor a temporary file behind.
func TestWriteFileError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	wantErr := errors.New("translate failed")
	err := writeFile(path, func(w io.Writer) error {
		_, _ = io.WriteString(w, "partial")
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("writeFile: want error %v, got %v", wantErr, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("target: want not to exist, got stat error %v", err)
	}
	assertOnlyFiles(t, dir)
}

// TestWriteFileErrorPreservesExisting is a regression test for issue #19:
// when the write func fails, an existing target keeps its old content.
func TestWriteFileErrorPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	if err := os.WriteFile(path, []byte("old content\n"), 0644); err != nil {
		t.Fatalf("write existing target: %v", err)
	}
	wantErr := errors.New("translate failed")
	err := writeFile(path, func(w io.Writer) error {
		_, _ = io.WriteString(w, "partial")
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("writeFile: want error %v, got %v", wantErr, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "old content\n" {
		t.Errorf("content: want %q, got %q", "old content\n", got)
	}
	assertOnlyFiles(t, dir, "out.csv")
}

// assertOnlyFiles fails the test unless dir contains exactly the named files.
func assertOnlyFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
	}
	if len(got) != len(names) {
		t.Fatalf("dir: want files %q, got %q", names, got)
	}
	for i := range names {
		if got[i] != names[i] {
			t.Fatalf("dir: want files %q, got %q", names, got)
		}
	}
}
