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
	Account               AccountDetail                 `json:"-"`
	Accounts              []*AccountDetail              `json:",omitempty"`
	Banks                 []*BankDetail                 `json:"-"`
	Budget                []*BudgetDetail               `json:"-"`
	Cash                  []*CashDetail                 `json:"-"`
	Categories            []*CategoryDetail             `json:",omitempty"`
	CreditCards           []*CreditCardDetail           `json:"-"`
	Investments           []*InvestmentDetail           `json:"-"`
	MemorizedTransactions []*MemorizedTransactionDetail `json:"-"`
	OtherAssets           []*OtherAssetDetail           `json:"-"`
	OtherLiabilities      []*OtherLiabilityDetail       `json:"-"`
	Prices                []*PriceDetail                `json:"-"`
	Securities            []*SecurityDetail             `json:"-"`
	Tags                  []*TagDetail                  `json:",omitempty"`
}

// AccountDetail is
type AccountDetail struct {
	Name                 string
	Type                 string
	CreditLimit          int
	Descr                string
	StatementBalance     int
	StatementBalanceDate string
	Banks                []*BankDetail           `json:",omitempty"`
	Budget               []*BudgetDetail         `json:",omitempty"`
	Cash                 []*CashDetail           `json:",omitempty"`
	CreditCards          []*CreditCardDetail     `json:",omitempty"`
	Investments          []*InvestmentDetail     `json:",omitempty"`
	OtherAssets          []*OtherAssetDetail     `json:",omitempty"`
	OtherLiabilities     []*OtherLiabilityDetail `json:",omitempty"`
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

// BudgetDetail is
type BudgetDetail struct {
	Raw []string
}

// CashDetail is
type CashDetail struct {
	Raw []string
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
	Raw []string
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

// OtherAssetDetail is
type OtherAssetDetail struct {
	Raw []string
}

// OtherLiabilityDetail is
type OtherLiabilityDetail struct {
	Raw []string
}

// PriceDetail is
type PriceDetail struct {
	Raw    []string
	Symbol string
	Price  string
	Date   string
}

// Split allows a detail line to be split into multiple transfers
type Split struct {
	Amount   int    // Dollar amount of split
	Category string // Category in split (Category/Transfer/Class)
	Memo     string // in split
}

// SecurityDetail is
type SecurityDetail struct {
	Description string
	Label       string
	Risk        string
	Symbol      string
	Type        string
}

// TagDetail is
type TagDetail struct {
	Description string
	Label       string
}
