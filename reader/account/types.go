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

package account

// accountTypes is the single list of QIF account types that the reader and
// writers know about. Each entry maps an account type (the `T` field of an
// `!Account` record) to the type in the `!Type:` header that Quicken writes
// that account's transactions under. The investment-style accounts share
// the `!Type:Invst` header.
var accountTypes = []struct {
	accountType     string
	transactionType string
}{
	{"Bank", "Bank"},
	{"Cash", "Cash"},
	{"CCard", "CCard"},
	{"Invst", "Invst"},
	{"Oth A", "Oth A"},
	{"Oth L", "Oth L"},
	{"Port", "Invst"},
	{"401(k)/403(b)", "Invst"},
}

// TransactionType returns the type in the `!Type:` header of the section
// that holds the transactions for an account of the given type. For example,
// TransactionType("Port") returns "Invst", true. It returns "", false if the
// account type is not known.
func TransactionType(accountType string) (string, bool) {
	for _, t := range accountTypes {
		if t.accountType == accountType {
			return t.transactionType, true
		}
	}
	return "", false
}

// TransactionTypes returns the distinct types used in the `!Type:` headers
// of account transaction sections, in a fixed order. It does not include
// the `Memorized` and `Prices` sections, which don't belong to an account.
func TransactionTypes() []string {
	var types []string
	seen := map[string]bool{}
	for _, t := range accountTypes {
		if !seen[t.transactionType] {
			seen[t.transactionType] = true
			types = append(types, t.transactionType)
		}
	}
	return types
}
