// Hand-authored accessor — lets the decisions package attach its schema
// and reuse the same *sql.DB rather than opening a second connection against the
// same SQLite file (which would fight the WAL journaling).
//
// This file is not generated.

package store

import "database/sql"

// DB returns the underlying *sql.DB. Callers must not Close it — Store owns the lifecycle.
func (s *Store) DB() *sql.DB { return s.db }
