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
- [x] Each driver provides `NewQueryBuilder()` pre-configured with correct style
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

## Phase 6: Testing
- [x] Unit tests for SELECT (basic, star, multiple WHERE, offset)
- [x] Unit tests for INSERT (basic, empty)
- [x] Unit tests for UPDATE (basic, multiple SET, empty)
- [x] Unit tests for DELETE (basic, all, multiple WHERE)
- [x] Unit tests for `$1, $2...` placeholder style (SELECT, INSERT, UPDATE, DELETE)
- [x] Unit tests for `@p1, @p2...` placeholder style (SELECT, INSERT, UPDATE, DELETE)
- [x] Unit tests for quote-aware replacement (single, double, escaped quotes, multi-clause)
- [x] Unit tests for all 4 driver packages (placeholder style, interface compliance)
- [x] Unit tests for logging (Exec, Query, QueryRow, BeginTx, Close, error logging)
- [x] Unit tests for NoopLogger, nil defaults, SetLogger, duration logging
- [x] Integration tests using real SQLite (via modernc.org/sqlite)
- [ ] Integration tests for MySQL, PostgreSQL, SQL Server

## Phase 7: Documentation
- [x] README with installation, quick start, query builder, and logging examples
- [x] API reference in `docs/api.md`
- [ ] Add GoDoc comments to all exported types and methods
- [ ] Add runnable examples for each database driver

## Phase 8: Finalization
- [ ] Review code quality and performance
- [ ] Ensure cross-database compatibility (SQL dialect differences beyond placeholders)
- [ ] Run full test suite
- [ ] Tag initial release
