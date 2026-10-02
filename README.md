# qif

`github.com/maloquacious/qif` reads QIF (Quicken Interchange Format) exports and translates them to CSV, JSON and [Ledger](https://ledger-cli.org) text. The `qifxlat` command line tool does the same for a file.

It does **not** write QIF, does not parse amounts as numbers, and does not check that transactions balance.

This README is the reference for the API, the QIF it accepts and the output formats. Every identifier, path and flag named here exists in the code. Each exported identifier's doc comment (`go doc ./...`) describes that symbol in more detail.

## Quick start

### CLI

```sh
go run ./cmd/qifxlat -input file.qif \
    -output-csv-filename out.csv \
    -output-json-filename out.json \
    -output-ledger-filename out.ledger
```

With no output flags, `qifxlat` only checks that the file parses. See [CLI reference](#cli-reference).

### Library

```go
sc, err := scanner.New(data)         // data is the []byte of a .qif file
if err != nil { ... }
r, err := reader.Read(sc)            // *reader.Reader
if err != nil { ... }
j, err := json.Translate(r, logger)  // writer/json; a nil *slog.Logger discards the log
if err != nil { ... }
err = j.Write(os.Stdout)
```

The runnable version of this is `Example` in [`example_test.go`](example_test.go). It writes Ledger text and is checked by `go test`.

## Package layout

Data flows one way:

```
[]byte → scanner.New → reader.Read → *reader.Reader → normalizer.Transactions → {csv,json,ledger}.Translate → Write(io.Writer)
```

| Package | Path | Purpose |
|---|---|---|
| `qif` | `version.go` | `Version()` only. |
| `scanner` | `scanner/` | `Scanner`, an immutable cursor over the input. |
| `reader` | `reader/` | `Read`, which parses a whole file into a `Reader`. |
| `account` | `reader/account/` | `!Account` records, and the account type → `!Type:` header table. |
| `category` | `reader/category/` | `!Type:Cat` records. |
| `security` | `reader/security/` | `!Type:Security` records. |
| `tag` | `reader/tag/` | `!Type:Tag` records. |
| `transaction` | `reader/transaction/` | Transaction records for account, `Memorized` and `Prices` sections. |
| `section` | `reader/internal/section/` | The generic header → records → end-of-section reader that the packages above share. Internal. |
| `normalizer` | `normalizer/` | Flattens transactions for the writers and marks duplicate transfer halves. |
| `csv` | `writer/csv/` | CSV writer. |
| `json` | `writer/json/` | JSON writer. |
| `ledger` | `writer/ledger/` | Ledger writer. |
| `stdlib` | `stdlib/` | `Date` and `FlipSign` helpers. |
| `main` | `cmd/qifxlat/` | The `qifxlat` command. |

## API

### scanner

- `scanner.New(input []byte) (Scanner, error)` copies the input, **removes every `\r`**, appends a final `\n` if one is missing, and returns an error for invalid UTF-8.
- `Scanner` is a **value type** with exported `Line`, `Col` (both 1-based) and `Buffer`. Its methods never change the receiver. Each returns `(lexeme []byte, next Scanner)`.
  - A nil lexeme means no match, and `next` is the receiver unchanged. To backtrack, keep the old value.
  - `Field(flag)` matches the flag and returns the rest of the line.
  - `Date(flag)` does the same, but only matches a valid QIF date, and returns it as `yyyy/mm/dd`.
  - `Literal(lit)` matches a literal (a section header) and returns the rest of the line.
  - `EndOfRecord()` matches `^`. `EndOfSection()` matches `!` without consuming it. Both also match the end of input.
  - `EndOfLine()` and `ToEndOfLine()` match the end of the line.

### reader

`reader.Read(sc scanner.Scanner) (*reader.Reader, error)` parses the whole input. It returns **only the first error**, and never a partial `Reader`.

`reader.Reader` fields:

| Field | Type | Contents |
|---|---|---|
| `Accounts` | `*account.Section` | The account list. **nil** if the file has none. |
| `Categories` | `*category.Section` | **nil** if the file has no `!Type:Cat` section. |
| `Securities` | `*security.Section` | The records of every `!Type:Security` section, combined. **nil** if there are none. |
| `Tags` | `*tag.Section` | **nil** if the file has no `!Type:Tag` section. |
| `Transactions` | `[]*transaction.Record` | Transactions from every account section, in input order. |
| `Memorized` | `[]*transaction.Record` | `!Type:Memorized` records. `Type` is `"Memorized"`; `Account` is empty. |
| `Prices` | `[]*transaction.Record` | `!Type:Prices` records. `Type` is `"Prices"`; `Account` is empty. |

A section with no records counts as missing. **Check the four pointer fields for nil before using them.** Each `Section` has `Line`, `Col` (where its header is) and `Records []*Record`.

**How a transaction gets its `Account` and `Type`.** A QIF transaction record doesn't name its account, so `Read` keeps an *active account* and copies its name and type onto each transaction:

- A one-record `!Account` section directly followed by an account transaction header (`!Type:Bank` etc.) makes that record the active account. If no account list has been seen yet, it also becomes `Accounts`.
- Otherwise the first non-empty `!Account` section is the account list. Quicken wraps it in `!Option:AutoSwitch` … `!Clear:AutoSwitch`.
- After the list, a one-record `!Account` section sets the active account. A second list with more than one record is an error.
- A transaction section is only accepted if its header is the active account's transaction header. `account.TransactionType` maps account types to headers. `Port` and `401(k)/403(b)` accounts use `!Type:Invst`, but their records keep the account's own `Type`. A transaction section before any account is active is an error.

Record types (every one also has `Line` and `Col`):

| Type | Fields |
|---|---|
| `account.Record` | `Name`, `Type`, `Description`, `CreditLimit`, `StatementBalance`, `StatementBalanceDate` |
| `category.Record` | `Name`, `Description`, `IsIncome`, `IsTaxRelated`, `TaxSchedule`, `BudgetAmount []string` |
| `security.Record` | `Name`, `Ticker`, `Type`, `Risk`, `Description` |
| `tag.Record` | `Name`, `Description` |
| `transaction.Record` | `Account`, `Type`, `Date`, `AmountTCode`, `AmountUCode`, `ClearedStatus`, `RefNo`, `Payee`, `Memo`, `Address []string`, `Category`, `Class`, `ToAccount`, `Split []*Split`, `Commission`, `Interest`, `Quantity`, `Price`, `Ticker`, `MemorizedFlag`, `BudgetAmount []string` |
| `transaction.Split` | `Category`, `Account` (transfer account), `Class`, `Memo`, `Amount` |

Which QIF field fills each struct field is in [QIF support](#qif-support).

Each `reader/<kind>` package also exports `ReadSection` and `ReadRecord`. `Read` calls them; you rarely need them directly. `account.TransactionType(accountType)` and `account.TransactionTypes()` hold the only list of known account types.

### normalizer

- `normalizer.Transactions([]*transaction.Record) []*normalizer.Transaction` flattens records, keeping their order:
  - Every transaction has **at least one `Split`**. A record with no splits becomes one split built from `T` (amount), `L` (category or transfer account) and `M`. The memo moves from the transaction to that split.
  - `IsZero` is set when an amount is `""` or `"0.00"`. On a `Transaction`, it is set when every split is zero.
  - Quicken records each transfer twice, once in each account. Two splits are halves of the same transfer when they have the same date, each names the other's account, and their amounts are opposites (after removing `,` and `+`). Halves are paired in input order. The duplicate half gets `IsLinked`: the `Oth L` half if exactly one half is in an `Oth L` account, otherwise the half found later. A transfer to the account itself (an opening balance), or one with no matching half, is never linked.
  - `Transaction.IsLinked` is true only when every split is linked.
- `normalizer.ByAccount(*reader.Reader) ([]*normalizer.Account, error)` runs `Transactions` and groups the result under each account record (`Record`, `Transactions`), in account-list order. It includes accounts with no transactions. It returns an error for a name listed twice, and for a transaction whose account isn't in the list.

### writers

Every writer has `Translate(r *reader.Reader, logger *slog.Logger) (*T, error)` and `(*T).Write(w io.Writer) error`. `Write` logs a summary at `Info` to the logger passed to `Translate`. A **nil logger discards** it. Library packages never print and never use `slog.Default()`.

| | `writer/csv` | `writer/json` | `writer/ledger` |
|---|---|---|---|
| Type | `*csv.CSV` | `*json.JSON` | `*ledger.Ledger` |
| Needs an account list | yes (via `ByAccount`) | no | no |
| Order | date, account name, line | input order | date, line |
| Linked (duplicate) transfers | skipped | **included** | skipped |
| Zero amounts | zero splits and transactions skipped | included | transactions with no amount skipped |
| Class | dropped | dropped | dropped |
| Amount signs | as in QIF, except `Oth A`/`Oth L` opening balances | as in QIF | flipped, except single-line opening balances |
| `Translate` errors | unknown account type, `ByAccount` errors | unknown account type | single-line opening balance in an unknown account type |

**CSV.** One header row, then one row per written split:

| Column | Value |
|---|---|
| `LINE` | Line of the transaction |
| `SEQ` | 1-based number of the split among those written for the transaction |
| `DATE` | `yyyy/mm/dd` |
| `STATUS` | Cleared status (`C`) |
| `REFNO` | Check or reference number (`N`) |
| `PAYEE` | Payee (`P`) |
| `MEMO` | Transaction memo; empty when there were no splits (the memo is then in `SMEMO`) |
| `ALINE` | Line of the account record |
| `ATYPE` | `BNK`, `CCD`, `CSH`, `ASS` (`Oth A`), `LBT` (`Oth L`), `INV` (`Invst`), `BRK` (`Port`), `RET` (`401(k)/403(b)`) |
| `ANAME` | Account name |
| `SLINE` | Line of the split |
| `TOACCT` | Transfer account |
| `CATEGORY` | Category |
| `SMEMO` | Split memo |
| `AMOUNT` | Amount with `,` removed |
| `FLIPPED` | `true` when the sign was flipped: a one-split `Opening Balance` in an `ASS` or `LBT` account |

**JSON.** An indented object `{"accounts": [...], "categories": [...], "transactions": [...]}`. A section the file lacks is `null`.
- Account: `type` (`bank`, `creditCard`, `cash`, `asset`, `liability`, `investment`, `brokerage`, `retirement`), `name`, `credit_limit`, `descr`, `balance`, `statement_date`.
- Category: `name`, `descr`, `income`, `tax_related`, `tax_schedule`.
- Transaction: `line`, `type` (QIF account type), `date`, `account`, `cleared_status`, `memo`, `payee`, `ref_no`, `lines`.
- Split, in `lines`: `line`, `account`, `amount`, `category`, `memo`.

Empty values are omitted. Securities, tags, memorized transactions and prices are not written.

**Ledger.** One entry per transaction:
- The header line is `date [cleared]  [(refno)] payee`, with a `;; line type account` comment. A missing payee is written as `Missing Payee`.
- The memo follows as a `; memo` comment.
- Then there is a posting per split that isn't linked. The posting's account is the first non-empty one of: transfer account, category, ticker, memo, `Missing Category`.
- The entry ends with a balancing posting with no amount: `Equity:Opening Balances` for a single-line `Opening Balance`, the entry's own account otherwise.
- An account or category name that contains a double space, or starts with `check`, has its spaces replaced by `_`.

### stdlib

- `stdlib.Date(b []byte) string` converts `m/dd'yy` to `yyyy/mm/dd`, and returns `"****/**/**"` for anything invalid.
- `stdlib.FlipSign(amount string) string` toggles a leading `-`, treating a leading `+` as positive. `""` and `"0.00"` become `"0.00"`.

## QIF support

### Section headers

Each header must start its line exactly as shown.

| Header | Read by | Notes |
|---|---|---|
| `!Option:AutoSwitch`, `!Clear:AutoSwitch` | `reader.Read` | Ignored. |
| `!Account` | `account.ReadSection` | The account list, or the active-account header (see [reader](#reader)). |
| `!Type:Cat` | `category.ReadSection` | At most one non-empty section. |
| `!Type:Security` | `security.ReadSection` | May repeat; the records are combined. |
| `!Type:Tag` | `tag.ReadSection` | At most one non-empty section. |
| `!Type:Bank`, `!Type:Cash`, `!Type:CCard`, `!Type:Invst`, `!Type:Oth A`, `!Type:Oth L` | `transaction.ReadSection` | Must match the active account's type. |
| `!Type:Memorized` | `transaction.ReadSection` | |
| `!Type:Prices` | `transaction.ReadSection` | |

Any other header, including `!Type:Class` and `!Type:Invoice`, is an error.

### Field codes

A record is a run of field lines ending with a `^` line. The first character of each line is the field code, and the rest of the line is its value. Unless the table says it repeats, a field may appear once; a second occurrence ends the record early, which causes a "missing record terminator" error. An unknown code causes the same error.

To support a new field code, add a branch to the loop in that package's `ReadRecord`, in `reader/<kind>/record.go`.

**`!Account`** (`account.Record`). `N` is required.

| Code | Field |
|---|---|
| `N` | `Name` |
| `T` | `Type`: `Bank`, `Cash`, `CCard`, `Invst`, `Oth A`, `Oth L`, `Port`, `401(k)/403(b)` |
| `D` | `Description` |
| `L` | `CreditLimit` |
| `$` | `StatementBalance` |
| `/` | `StatementBalanceDate` (a date) |

The reader accepts any `T` value, but a transaction section for an unknown type is never matched, and the CSV and JSON writers return an error for one.

**`!Type:Cat`** (`category.Record`). `N` is required.

| Code | Field |
|---|---|
| `N` | `Name` |
| `D` | `Description` |
| `I` | `IsIncome = true` (value ignored) |
| `E` | `IsIncome = false` (value ignored; this is the default) |
| `T` | `IsTaxRelated = true` (value ignored) |
| `R` | `TaxSchedule` |
| `B` | `BudgetAmount` (repeats) |

**`!Type:Security`** (`security.Record`). `N` is required.

| Code | Field |
|---|---|
| `N` | `Name` |
| `S` | `Ticker` |
| `T` | `Type` |
| `G` | `Risk` (goal) |
| `D` | `Description` |

**`!Type:Tag`** (`tag.Record`). `N` is required.

| Code | Field |
|---|---|
| `N` | `Name` |
| `D` | `Description` |

**Transactions** (`transaction.Record`), used for account sections, `!Type:Memorized` and `!Type:Prices`. `D` is required, except in a `Memorized` record, which requires `K`; a price line also satisfies the requirement.

| Code | Field |
|---|---|
| `D` | `Date` (a date) |
| `T` | `AmountTCode` |
| `U` | `AmountUCode` |
| `C` | `ClearedStatus` |
| `N` | `RefNo` (check number; the action in investment accounts) |
| `P` | `Payee` |
| `M` | `Memo` |
| `A` | `Address` (repeats) |
| `L` | `Category`, `ToAccount` and `Class`, from `Category[:Sub]` or `[Account]`, then an optional `/Class` |
| `S` | Starts a new `Split`, and fills its `Category`, `Account` and `Class` like `L` (repeats) |
| `E` | `Memo` of the current split, starting one if there is none (repeats) |
| `$` | `Amount` of the current split, starting one if there is none (repeats) |
| `Y` | `Ticker` |
| `Q` | `Quantity` |
| `I` | `Interest` |
| `O` | `Commission` |
| `K` | `MemorizedFlag` |
| `1` to `7` | `BudgetAmount`, appended (`Memorized` records only; repeats) |
| `"` | A price line, `"SYMBOL",price,"date"`: sets `Ticker`, `Price` (with `,` removed) and `Date` (repeats) |

### Errors

`Read` stops at the first problem. Its messages start with the position, and errors from inside a section are prefixed with that section's line and name:

```
1:1: unexpected input
1:1: transaction section "!Type:Bank" found before any !Account header
1: categories: 2: category: missing field "name"
5: transactions: 8: transaction: missing record terminator
5: transactions: 6:1: unexpected input
```

Input that is rejected:
- empty input
- blank lines
- unknown headers
- unknown field codes
- a repeated single-occurrence field
- a record without its required field
- a date that doesn't match the grammar or doesn't exist
- a transaction section for an account type other than the active one
- a second account, category or tag list
- invalid UTF-8

Inputs that used to panic (#3, #8, #9, #12, #17, #18) now return errors, and tests cover them.

## Data conventions

- **Dates** are `yyyy/mm/dd` strings, so comparing them as strings sorts them chronologically. QIF dates are `m/dd'yy`:
  - The month is one or two digits, with no leading space.
  - The day is two characters, where the first may be a space (` 1`).
  - The year is two digits and is **always 2000–2099**.
  - An impossible date (`2/30'21`, month `13`) doesn't match, so its record fails to parse. The exception is a price line, where the date becomes `****/**/**` with no error.
  - Text after the year is ignored.
- **Amounts are strings** exactly as in the file, and may contain `,` thousands separators (`-1,000.00`). Nothing parses them as numbers. Only the CSV `AMOUNT` column and a transaction's `Price` have the commas removed.
- **Transfers** are written `[Account]` in `L` or `S`. The name goes in `ToAccount` (or `Split.Account`), and `Category` is empty.
- **Opening balances** are transactions with payee `Opening Balance`. A single-line one is a transfer to the account itself, so it is never linked. The CSV and Ledger writers treat its sign specially (see [writers](#writers)).
- **Line and Col** fields are where an item starts in the source file. Errors and output comments use them.

## CLI reference

`qifxlat` reads one QIF file and writes each requested output. Diagnostics go to stderr via `log/slog`; only `-version` writes to stdout. On any error it exits with status 2. Outputs are written to a temporary file and renamed into place, so a failed run never leaves a partial file.

| Flag | Environment variable | Default | Meaning |
|---|---|---|---|
| `-input` | `QIFXLAT_INPUT` | | QIF file to read (required) |
| `-output-csv-filename` | `QIFXLAT_OUTPUT_CSV_FILENAME` | | Write CSV here |
| `-output-json-filename` | `QIFXLAT_OUTPUT_JSON_FILENAME` | | Write JSON here |
| `-output-ledger-filename` | `QIFXLAT_OUTPUT_LEDGER_FILENAME` | | Write Ledger here |
| `-log-level` | `QIFXLAT_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. Stage durations are logged at `debug`. |
| `-log-format` | `QIFXLAT_LOG_FORMAT` | `text` | `text` or `json` |
| `-version` | `QIFXLAT_VERSION` | `false` | Print the version and exit |

Every flag's variable is `QIFXLAT_` plus the flag name in upper case with `-` replaced by `_`. A flag on the command line overrides its variable. `-h` lists both. There is no config file.

## Known limitations

There are no open issues for these.

- QIF is read, never written.
- Only the headers in [Section headers](#section-headers) are supported. Classes (`!Type:Class`), invoices and other lists are rejected.
- Securities, tags, memorized transactions and prices are read, but no writer outputs them.
- No writer outputs investment details (`Y`, `Q`, `I`, `O`, `U`, `Price`) or the class.
- Two-digit years are always 2000–2099.
- Empty input and blank lines are errors.

## Development

```sh
gofmt -l .        # must print nothing
go vet ./...
go test ./...
```

- **Go version:** `go.mod` declares `go 1.24.0`, so don't use features from later releases. Raise it only in a change of its own.
- **Tests** use only the standard library. Prefer table-driven tests, and add a regression test with each fix.
- **License header:** every `.go` file starts with the MIT license block. Copy it from an existing file.
- **Issues and branches:** work is tracked in GitHub issues.
  - Each issue gets its own branch, `fix/<n>-<slug>`.
  - The PR body includes `Fixes #<n>`.
  - PRs are squash-merged.
  - Fix open `bug` issues before starting `enhancement` ones.
- **Version:** every change bumps `version.go` (see [Versioning](#versioning)).

## Versioning

The module version is defined in `version.go` (`qif.Version()`), and `qifxlat -version` prints it.

Every change to code or documentation bumps the version in `version.go`, in the same commit or PR:

- **Minor** (`0.x.0`): a new feature, a behavior change, or a change to an exported API. Reset patch to 0.
- **Patch** (`0.x.y`): a bug fix, a documentation change, refactoring, or tests with no change in behavior.
- **Major** stays `0` until the API is declared stable.

Releases are tagged `vMAJOR.MINOR.PATCH` (e.g. `v0.3.0`). The older `v0.1` and `v0.2` tags predate this scheme.
