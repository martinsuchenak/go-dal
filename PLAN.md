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
- [x] Define `DBInterface` (Exec, Query, QueryRow, BeginTx, Close)
- [x] Define `PlaceholderStyle` (QuestionMark, DollarNumber, AtPNumber)
- [x] Define query types (`SelectQuery`, `InsertQuery`, `UpdateQuery`, `DeleteQuery`)
- [x] Use ordered slices instead of maps for deterministic column ordering
- [x] Table is set at creation time (no redundant `Into()`/`Table()` methods)

## Phase 3: Query Builder Implementation
- [x] SELECT with FROM, WHERE, ORDER BY, LIMIT, OFFSET
- [x] INSERT with ordered SET pairs
- [x] UPDATE with ordered SET pairs and WHERE
- [x] DELETE with WHERE
- [x] Quote-aware placeholder replacement (skips `?` inside string literals)
- [x] Automatic placeholder numbering across multiple WHERE clauses

## Phase 4: Database Driver Implementations
- [x] MySQL driver wrapper (`pkg/mysql`) -- `?` placeholders
- [x] PostgreSQL driver wrapper (`pkg/postgres`) -- `$1, $2, ...` placeholders
- [x] SQLite driver wrapper (`pkg/sqlite`) -- `?` placeholders
- [x] SQL Server driver wrapper (`pkg/mssql`) -- `@p1, @p2, ...` placeholders
- [x] Compile-time interface compliance checks for all drivers
- [x] Each driver provides `NewQueryBuilder()` pre-configured with correct dialect
- [x] Shared `BaseDB` in `pkg/dal` eliminates duplication across drivers

## Phase 5: Logging Layer
- [x] Define `Logger` interface (compatible with `fortix/go-libs/logger`)
- [x] `NoopLogger` for silent operation (default when no logger provided)
- [x] `BaseDB` with logging for Exec, Query, QueryRow, BeginTx, Close
- [x] Log query, args, and duration at Debug level
- [x] Log errors at Error level with query and duration
- [x] Optional logger via constructor: `NewXxxDB(db, logger)`
- [x] Runtime logger change via `SetLogger(logger)` -- pass nil to disable
- [x] Access underlying `*sql.DB` via `DB()` method

## Phase 6: Dialect Architecture
- [x] `Dialect` interface with BuildSelect/BuildInsert/BuildUpdate/BuildDelete
- [x] `BaseDialect` with configurable PlaceholderStyle + LimitStyle
- [x] `LimitStyle`: LimitOffsetStyle (MySQL/PostgreSQL/SQLite) and FetchNextStyle (MSSQL)
- [x] Each driver provides `NewDialect()` returning a configured `BaseDialect`
- [x] Query structs hold a `Dialect` reference, `Build()` delegates to dialect
- [x] Extensible: new databases = implement Dialect, zero changes to shared code

## Phase 7: Testing
- [x] Unit tests for SELECT, INSERT, UPDATE, DELETE (all placeholder styles)
- [x] Unit tests for quote-aware replacement
- [x] Unit tests for all 4 driver packages
- [x] Unit tests for logging layer
- [x] Unit tests for MSSQL LIMIT/OFFSET dialect
- [x] Integration tests across SQLite, MySQL, PostgreSQL, MSSQL (CRUD, JOINs, aggregation, transactions)
- [x] Runnable Go examples (query builder, SQLite CRUD, placeholder styles)

## Phase 8: Documentation
- [x] README with installation, quick start, query builder, logging, dialect examples
- [x] API reference in `docs/api.md`
- [x] GoDoc comments on all exported types and methods

## Phase 9: Finalization
- [x] Dialect architecture refactored from monolithic Build()
- [x] Cross-database compatibility (MSSQL FETCH NEXT dialect)
- [x] Full test suite passing (58 unit + 8 examples + 19 integration)
- [ ] Tag initial release
