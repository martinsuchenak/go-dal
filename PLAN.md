# GO-DAL Development Plan

## Project Overview
GO-DAL is a lightweight, interface-driven database abstraction layer for Go that allows developers to write database-agnostic SQL queries across MySQL, PostgreSQL, SQLite, and SQL Server.

---

## Phase 1: Project Setup
- [x] Initialize Go module (go 1.26.2)
- [x] Create project structure (`pkg/dal`, `pkg/mysql`, etc.)
- [x] Set up configuration files (.gitignore, Taskfile.yml)
- [x] Remove standalone app artifacts (main.go, cmd/) -- this is a library only

## Phase 2: Core Types and Interfaces
- [x] Define `DBInterface` (Exec, Query, QueryRow, BeginTx, Ping, Close)
- [x] Define RETURNING hooks: `AppendReturning`, `PrependReturning` function fields
- [x] Define query types (`SelectQuery`, `InsertQuery`, `UpdateQuery`, `DeleteQuery`)
- [x] Use ordered slices instead of maps for deterministic column ordering
- [x] Table is set at creation time (no redundant `Into()`/`Table()` methods)
- [x] Define sentinel errors: ErrEmptyTable, ErrEmptyColumns, ErrEmptyInValues, ErrReturningNotSupported, ErrBatchRowLength, ErrTooManyInValues

## Phase 3: Query Builder Implementation
- [x] SELECT with FROM, WHERE, ORDER BY, LIMIT, OFFSET
- [x] INSERT with ordered SET pairs
- [x] UPDATE with ordered SET pairs and WHERE
- [x] DELETE with WHERE
- [x] Quote-aware placeholder replacement (skips `?` inside string literals)
- [x] Automatic placeholder numbering across multiple WHERE clauses
- [x] OR support in WHERE (OrWhere) and HAVING (OrHaving)
- [x] DISTINCT support
- [x] IN-clause expansion (dal.In) with MaxInValues cap (1000)
- [x] Batch INSERT (Columns + Values)
- [x] INSERT RETURNING / OUTPUT
- [x] UPDATE RETURNING / OUTPUT
- [x] DELETE RETURNING / OUTPUT
- [x] SelectAll() alias
- [x] WHERE groups: WhereGroup, OrWhereGroup (parenthesized conditions)
- [x] WHERE shortcuts: WhereIsNull, WhereIsNotNull, WhereBetween
- [x] Build() returns (string, []interface{}, error) — always check errors
- [x] In() returns (InValues, error) — validates non-empty and max count

## Phase 4: Database Driver Implementations
- [x] MySQL driver wrapper (`pkg/mysql`) — `?` placeholders, backtick quoting, BackslashEscapes
- [x] PostgreSQL driver wrapper (`pkg/postgres`) — `$1, $2, ...` placeholders, double-quote quoting, RETURNING
- [x] SQLite driver wrapper (`pkg/sqlite`) — `?` placeholders, double-quote quoting, RETURNING
- [x] SQL Server driver wrapper (`pkg/mssql`) — `@p1, @p2, ...` placeholders, bracket quoting, OUTPUT
- [x] Compile-time interface compliance checks for all drivers
- [x] Each driver provides `NewQueryBuilder()` pre-configured with correct dialect
- [x] Shared `BaseDB` in `pkg/dal` eliminates duplication across drivers

## Phase 5: Logging Layer
- [x] Define `Logger` interface (compatible with `fortix/go-libs/logger`)
- [x] `NoopLogger` for silent operation (default when nil passed)
- [x] `BaseDB` with logging for Exec, Query, QueryRow, BeginTx, Ping, Close
- [x] Log query, args, and duration at Debug level
- [x] Log errors at Error level with query and duration
- [x] Required logger via constructor: `NewXxxDB(db, logger)` — pass nil to disable
- [x] Runtime logger change via `SetLogger(logger)` — pass nil to disable
- [x] `SetLogArgs(bool)` to control argument redaction (default: redacted)
- [x] Access underlying `*sql.DB` via `DB()` method
- [x] `Tx` wrapper with logging for transaction-scoped Exec, Query, QueryRow, Commit, Rollback

## Phase 6: Dialect Architecture
- [x] `Dialect` interface with BuildSelect/BuildInsert/BuildUpdate/BuildDelete/QuoteIdentifier/SupportsReturning (all return error)
- [x] `BaseDialect` with configurable function fields (`Placeholder`, `AppendLimit`, `AppendReturning`, `PrependReturning`) + style flags (`QuoteStyle`, `BackslashEscapes`)
- [x] `Placeholder` function field: QuestionMarkPlaceholder, DollarPlaceholder, AtPPlaceholder
- [x] `AppendLimit` function field: LimitOffset, FetchNextLimit
- [x] `QuoteStyle`: NoQuoting, BacktickQuoting, DoubleQuoteQuoting, BracketQuoting
- [x] `AppendReturning` / `PrependReturning` function fields: `WriteReturning`, `WriteOutput`
- [x] Each driver provides `NewDialect()` returning a configured `BaseDialect`
- [x] Query structs hold a `Dialect` reference, `Build()` delegates to dialect
- [x] `buildClauses` helper eliminates repeated WHERE-building code
- [x] `replaceAndCount` merges placeholder replacement + counting into single pass
- [x] `SafeIdentifier()` validates identifier names
- [x] Extensible: new databases = implement Dialect, zero changes to shared code

## Phase 7: Testing
- [x] Unit tests for SELECT, INSERT, UPDATE, DELETE (all placeholder styles)
- [x] Unit tests for quote-aware replacement
- [x] Unit tests for all 4 driver packages (with identifier quoting)
- [x] Unit tests for logging layer (including Tx wrapper, Ping)
- [x] Unit tests for MSSQL LIMIT/OFFSET dialect
- [x] Unit tests for OR, DISTINCT, IN, batch INSERT, RETURNING
- [x] Unit tests for WHERE groups, BETWEEN, IS NULL/IS NOT NULL, OrHaving
- [x] Unit tests for error returns from Build() and In()
- [x] Integration tests across SQLite, MySQL, PostgreSQL, MSSQL (CRUD, JOINs, aggregation, transactions)
- [x] Runnable Go examples

## Phase 8: Portability Expression Helpers
- [x] 6 portable expression methods on `Dialect` interface: ConcatExpr, LengthExpr, CurrentTimestamp, BoolLiteral, StringAggExpr, RandExpr
- [x] `BaseDialect` default implementations in `pkg/dal/expressions.go`
- [x] `QueryBuilder` convenience wrappers (`qb.ConcatExpr()`, etc.) — no driver-specific imports needed
- [x] MSSQL overrides all 6 via `mssqlDialect` struct (embeds `*dal.BaseDialect`)
- [x] PostgreSQL overrides `StringAggExpr`, `RandExpr` via `postgresDialect`
- [x] SQLite overrides `CurrentTimestamp`, `BoolLiteral`, `StringAggExpr`, `RandExpr` via `sqliteDialect`
- [x] MySQL uses plain `*dal.BaseDialect` (all defaults work)
- [x] Unit tests in `pkg/dal/expressions_test.go`
- [x] Integration tests in `tests/integration/expressions_test.go`
- [x] Removed stale driver-level expression files (`pkg/*/expressions.go`, `pkg/*/expressions_test.go`)
- [x] Updated documentation (`docs/usage.md`, `AGENTS.md`)

## Phase 9: Documentation
- [x] README with installation, quick start, query builder, logging, dialect, portability notes
- [x] API reference in `docs/usage.md`
- [x] Contributing guide in `docs/contributing.md`
- [x] GoDoc comments on all exported types and methods

## Phase 10: Finalization
- [x] Removed 80 lines of redundant driver forwarding methods
- [x] Merged replacePlaceholders + countPlaceholders into single-pass replaceAndCount
- [x] Cross-database compatibility (MSSQL FETCH NEXT, identifier quoting)
- [x] Full test suite passing

## Phase 11: Comprehensive Review Fixes
- [x] Fix WhereIsNull/WhereIsNotNull/WhereBetween to quote column names
- [x] Fix BuildInsert missing empty table validation
- [x] Fix findFirstUnquotedPlaceholder to handle backslash escapes (MySQL)
- [x] Fix WriteOutput to use DELETED. prefix for DELETE queries
- [x] Fix race condition on SetLogger/SetLogArgs with sync.RWMutex
- [x] Remove dead code (ErrNotImplemented, redundant double-checks)
- [x] Fix all documentation compilation errors (Build() 3-value, constructor 2-param, In() error)
- [x] Fix Dialect interface docs to include expression methods
- [x] Fix contributing.md to include expressions.go
- [x] Fix PLAN.md duplicate Phase 9 numbering
- [ ] Tag initial release
