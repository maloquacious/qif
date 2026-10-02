# qif
QIF (Quicken Interchange Format) data conversion

This package implements a QIF reader that loads data into a very simple view of accounts, categories, and transactions.

It includes some writers to show how to use the imported data.

## Versioning

The module version is defined in `version.go` (`qif.Version()`), and `qifxlat -version` prints it.

Every change to code or documentation bumps the version in `version.go`, in the same commit or PR:

- **Minor** (`0.x.0`): a new feature, a behavior change, or a change to an exported API. Reset patch to 0.
- **Patch** (`0.x.y`): a bug fix, a documentation change, refactoring, or tests with no change in behavior.
- **Major** stays `0` until the API is declared stable.

Releases are tagged `vMAJOR.MINOR.PATCH` (e.g. `v0.3.0`). The older `v0.1` and `v0.2` tags predate this scheme.
