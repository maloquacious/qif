// Copyright © 2018 MICHAEL D HENDERSON <mdhender@mdhender.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qif

// File contains the data imported from the QIF file.
type File struct {
	Account               AccountDetail
	Accounts              []*AccountDetail
	Banks                 []*BankDetail
	Categories            []*CategoryDetail
	CreditCards           []*CreditCardDetail
	Investments           []*InvestmentDetail
	MemorizedTransactions []*MemorizedTransactionDetail
	Tags                  []*TagDetail
}

// AccountDetail is
type AccountDetail struct {
	Name                 string
	Type                 string
	CreditLimit          int
	Descr                string
	StatementBalance     int
	StatementBalanceDate string
}

// BankDetail is
type BankDetail struct {
	Address       []string // Up to five lines (the sixth line is an optional message)
	AmountTCode   int
	AmountUCode   int
	Category      string // Category/Subcategory/Transfer/Class
	ClearedStatus string
	Date          string
	Memo          string
	Num           string // (check or reference number)
	Payee         string
	Split         []*Split
}

// CategoryDetail is
type CategoryDetail struct {
	Name        string // Category/Subcategory/Transfer/Class
	Descr       string
	IsExpense   bool
	TaxRelated  bool
	TaxSchedule string
}

// CreditCardDetail is
type CreditCardDetail struct {
	Address       []string // Up to five lines (the sixth line is an optional message)
	AmountTCode   int
	AmountUCode   int
	Category      string // Category/Subcategory/Transfer/Class
	ClearedStatus string
	Date          string
	Memo          string
	Num           string // (check or reference number)
	Payee         string
	Split         []*Split
}

// InvestmentDetail is
type InvestmentDetail struct {
}

// MemorizedTransactionDetail is
type MemorizedTransactionDetail struct {
	Type                     string
	Address                  []string // Up to five lines (the sixth line is an optional message)
	AmountTCode              int
	AmountUCode              int
	Category                 string // Category/Subcategory/Transfer/Class
	ClearedStatus            string
	Date                     string
	Memo                     string
	MemorizedTransactionType string
	Num                      string // (check or reference number)
	Payee                    string
	Split                    []*Split
}

// Split allows a detail line to be split into multiple transfers
type Split struct {
	Amount   int    // Dollar amount of split
	Category string // Category in split (Category/Transfer/Class)
	Memo     string // in split
}

// TagDetail is
type TagDetail struct {
	Label string
}
