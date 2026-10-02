# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`github.com/maloquacious/qif` parses QIF (Quicken Interchange Format) exports and converts them to CSV, JSON and [Ledger](https://ledger-cli.org) text. `cmd/qifxlat` is the only binary. It does not write QIF. The root package `qif` contains only `Version()` (`version.go`).

## Commands

```sh
go build ./...
go vet ./...
gofmt -l .                                  # must print nothing
go test ./...
go test -run TestReadSingleAccountHeader ./reader/   # a single test
go run ./cmd/qifxlat -input file.qif -output-json-filename out.json \
    -output-csv-filename out.csv -output-ledger-filename out.ledger
```

`qifxlat` flags can also come from `QIFXLAT_*` environment variables (e.g. `QIFXLAT_INPUT`) or from a plain-text file given with `-config` (via `peterbourgon/ff`). With no output flags, it only validates the input. `-version` prints the version and exits.

## Constraints

- `go.mod` declares `go 1.24.0`, so don't use features from later Go releases (e.g. `sync.WaitGroup.Go` or `testing.T.Output` from 1.25). Raise the version only in a change of its own.
- **Every change to code or docs bumps the version in `version.go`**, in the same commit or PR. Minor for features, behavior changes or exported-API changes (reset patch to 0); patch for fixes, docs, refactors and tests. Major stays 0. See README "Versioning".
- Every `.go` file starts with the MIT license header block. Copy it from an existing file.
- Tests use only the standard library.

## Architecture

Data flows one way:

```
[]byte → scanner.Scanner → reader.Read → *reader.Reader → normalizer.Transactions → writer/{csv,json,ledger}.Translate → .Write(io.Writer)
```

### scanner: an immutable cursor

`scanner.Scanner` is a **value type** holding the remaining buffer plus line/col. Each method (`Field`, `Date`, `Literal`, `EndOfRecord`, `EndOfSection`) returns `(lexeme, newScanner)`.
- A nil lexeme means "no match", and the returned scanner is unchanged.
- Backtracking is just keeping the old value (`saved := sc`).
- `New` strips every `\r` and rejects invalid UTF-8.
- `Date` validates through `stdlib.Date` and returns `yyyy/mm/dd`.

### reader: hand-written recursive descent

Each `reader/<kind>` package (`account`, `category`, `security`, `tag`, `transaction`) has the same two functions:
- `ReadSection` matches its `!Type:` / `!Account` header with `Literal`, calls `ReadRecord` until it returns nil, then requires `EndOfSection` (`!` or EOF).
- `ReadRecord` loops over the field codes:
  - Each single-occurrence field is guarded by `if x == nil`, so a repeated field ends the loop. Then `^` is required, otherwise it's a "missing record terminator" error.
  - Repeatable fields (address `A`, splits `S`/`E`/`$`, budget `B`) append instead.
  - A transaction's `L` value and a split's `S` value go through `parseCategory`, which separates `Category[:Sub]` or `[TransferAccount]` from the optional `/Class`.

  To support a new field code, add a branch to that loop.

`reader.Read` repeatedly tries each section reader in a fixed order: AutoSwitch markers, accounts, categories, securities, tags, active-account transactions, Memorized, Prices. If none match, it returns a `line:col: …` error. Errors are only returned, never collected; the first one wins.

**Active account:** transaction records don't name their own account. `Read` keeps `r.active.{account,accountType}` and stamps them onto each transaction, using these rules:
- A one-record `!Account` section directly followed by an account transaction header (`!Type:Bank|Cash|CCard|Invst|Oth A|Oth L`) is the active-account header. If no account list has been seen yet, it also becomes `r.Accounts`.
- Otherwise the first `!Account` section is the account list (normally wrapped in `!Option:AutoSwitch` … `!Clear:AutoSwitch`).
- A transaction section is only recognized when its header matches the active account's type, via `account.TransactionType` (`Port` and `401(k)/403(b)` accounts use `!Type:Invst`; records keep the account's own type). Before any account is active it's an error.
- The set of known account types and their transaction headers lives only in `reader/account` (`TransactionType`, `TransactionTypes`); the reader derives its header checks from it.

`Reader.Accounts`, `Categories`, `Securities` and `Tags` are **pointers that stay nil** when the file lacks that section. Check for nil before using them.

### normalizer

`normalizer.Transactions` flattens `transaction.Record`s for the writers:
- Every transaction gets at least one split; a transaction with no splits becomes one split from `T`/`L`.
- It sets `IsZero` (no non-zero amount).
- Quicken records each transfer twice, once in each account. `linkTransfers` pairs the two halves (same date, each names the other's account, opposite amounts after removing `,` and `+`) in input order, and sets `IsLinked` on the duplicate half: the `Oth L` half if exactly one side is `Oth L`, otherwise the half found later. A transfer to the account itself (an opening balance) or with no matching half (e.g. to an account missing from the export) is never linked. `Transaction.IsLinked` is true only when every split is linked.
- The CSV and ledger writers skip linked splits, and skip a transaction whose splits are all linked.

### writers

Each writer package has `Translate(*reader.Reader) (*T, error)` and `(*T).Write(io.Writer) error`. Each one prints its own progress counts to stdout (moving to slog is #22).
- `csv` and `json` map QIF account types to their own codes and return an error on unknown types.
- `ledger` flips amount signs except for single-line opening balances (`doFlipSign`), and adds a balancing posting to the source account.

### stdlib

`stdlib` holds small shared helpers: `Date` (QIF `m/d'yy` → `yyyy/mm/dd`, years assumed to be 2000–2099, invalid dates → `****/**/**`) and `FlipSign`.

## Data conventions

- Amounts are kept as the original **strings** (they may contain `,`). Nothing parses them to numbers.
- Dates are `yyyy/mm/dd` strings, so comparing them as strings sorts them chronologically.
- `Line`/`Col` fields record where each item appears in the source file, for error messages and output comments.

## Workflow

- Bugs and planned work are tracked as GitHub issues; check `gh issue list` before changing anything nearby.
- **Bugs before features.** While any issue labeled `bug` is open, don't start an `enhancement` issue unless the human has explicitly approved working on that specific issue. Fix the open bugs first (`gh issue list --label bug`). When asked for a feature while bugs are open, list the open bugs and ask before proceeding. Approval for one feature doesn't extend to the next.
- Work on an issue goes on its own branch (`fix/<n>-<slug>`). The PR body includes `Fixes #<n>`, and PRs are squash-merged.
- Changes not tied to an issue (repository upkeep, agent docs) are committed directly to `main` and pushed. Don't open a branch or PR for them.
- Add a regression test with each fix.
