// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

//go:build !wasm
// +build !wasm

package node

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
)

func TestOpenKeyValueDatabaseTreeDBFreshAutodetectReopen(t *testing.T) {
	dir := t.TempDir()
	db, err := openKeyValueDatabase(InternalOpenOptions{
		Directory: dir,
		DbEngine:  rawdb.DBTreedb,
	})
	if err != nil {
		t.Fatalf("open TreeDB: %v", err)
	}
	if err := db.Put([]byte("key"), []byte("value")); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := db.SyncKeyValue(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if got := rawdb.PreexistingDatabase(dir); got != rawdb.DBTreedb {
		t.Fatalf("PreexistingDatabase = %q, want %q", got, rawdb.DBTreedb)
	}

	reopen, err := openKeyValueDatabase(InternalOpenOptions{Directory: dir})
	if err != nil {
		t.Fatalf("autodetect reopen TreeDB: %v", err)
	}
	defer reopen.Close()
	got, err := reopen.Get([]byte("key"))
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if !bytes.Equal(got, []byte("value")) {
		t.Fatalf("get after reopen = %q, want value", got)
	}
}

func TestOpenKeyValueDatabaseTreeDBEngineConflict(t *testing.T) {
	dir := t.TempDir()
	db, err := openKeyValueDatabase(InternalOpenOptions{Directory: dir, DbEngine: rawdb.DBTreedb})
	if err != nil {
		t.Fatalf("open TreeDB: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err = openKeyValueDatabase(InternalOpenOptions{Directory: dir, DbEngine: rawdb.DBPebble})
	if err == nil {
		t.Fatal("expected conflict opening Pebble on existing TreeDB")
	}
	want := "db.engine choice was pebble but found pre-existing treedb database"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("conflict error = %q, want substring %q", err, want)
	}
}

func TestOpenKeyValueDatabaseTreeDBNotFoundClassification(t *testing.T) {
	db, err := openKeyValueDatabase(InternalOpenOptions{Directory: t.TempDir(), DbEngine: rawdb.DBTreedb})
	if err != nil {
		t.Fatalf("open TreeDB: %v", err)
	}
	defer db.Close()
	_, err = db.Get([]byte("missing"))
	if !rawdb.IsDbErrNotFound(err) {
		t.Fatalf("IsDbErrNotFound(%v) = false, want true", err)
	}
}

func TestOpenKeyValueDatabaseTreeDBDatabaseSuite(t *testing.T) {
	dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
		db, err := openKeyValueDatabase(InternalOpenOptions{Directory: t.TempDir(), DbEngine: rawdb.DBTreedb})
		if err != nil {
			t.Fatalf("open TreeDB: %v", err)
		}
		return db
	})
}
