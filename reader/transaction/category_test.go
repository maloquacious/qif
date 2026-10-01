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

package transaction

import "testing"

// TestParseCategory is a regression test for issue #14: the class after "/"
// must be split off before the brackets of a transfer account are removed.
func TestParseCategory(t *testing.T) {
	for _, tc := range []struct {
		in                       string
		category, account, class string
	}{
		{in: "Food", category: "Food"},
		{in: "Food:Groceries", category: "Food:Groceries"},
		{in: "Food/Business", category: "Food", class: "Business"},
		{in: "[Checking]", account: "Checking"},
		{in: "[Checking]/Business", account: "Checking", class: "Business"},
		{in: "[My Acct]/Biz", account: "My Acct", class: "Biz"},
		{in: ""},
		{in: "/Business", class: "Business"},
	} {
		category, account, class := parseCategory(tc.in)
		if category != tc.category || account != tc.account || class != tc.class {
			t.Errorf("parseCategory(%q): want (%q, %q, %q), got (%q, %q, %q)",
				tc.in, tc.category, tc.account, tc.class, category, account, class)
		}
	}
}
