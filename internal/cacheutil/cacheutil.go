// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package cacheutil

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
)

// Entry represents a cached artifact on disk.
// Key is the clear-text key; EncodedKey is the hashed filename.
type Entry struct {
	Key        string
	EncodedKey string
	Path       string
	Data       []byte
}

// Dir resolves the base cache directory.
// Precedence:
//  1. TFCTL_CACHE_DIR, if set and non-empty
//  2. os.UserCacheDir()/tfctl
//
// Returns ("", false) if a base cannot be resolved (treat as disabled).
func Dir() (string, bool) {
	if c, ok := os.LookupEnv("TFCTL_CACHE_DIR"); ok && c != "" {
		return c, true
	}
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "tfctl"), true
	}
	return "", false
}

// Enabled returns true unless TFCTL_CACHE explicitly disables it ("0"/"false").
func Enabled() bool {
	enabled, _ := os.LookupEnv("TFCTL_CACHE")
	return enabled == "" || (enabled != "0" && enabled != "false")
}

// EnsureBaseDir creates the base cache directory if caching is enabled and
// a base path can be resolved. Returns the path, whether it is usable, and an
// error if creation failed.
func EnsureBaseDir() (string, bool, error) {
	if !Enabled() {
		return "", false, nil
	}
	base, ok := Dir()
	if !ok {
		return "", false, nil
	}
	if err := os.MkdirAll(base, 0o755); err != nil { //nolint:mnd
		return base, false, fmt.Errorf("failed to create cache base directory: %w", err)
	}
	return base, true, nil
}

// EntryPath returns the absolute path where a cache entry would live given
// subdirectory components and the clear-text key. It also returns true if a
// file currently exists at that path.
func EntryPath(subdirs []string, clearKey string) (string, bool) {
	base, ok := Dir()
	if !ok {
		return "", false
	}
	encoded := encodeKey(clearKey)
	p := filepath.Join(append([]string{base}, append(subdirs, encoded)...)...)
	if _, err := os.Stat(p); err == nil {
		return p, true
	}
	return p, false
}

// Purge removes files older than the provided number of hours.
// If hours <= 0 or the cache dir cannot be resolved, it is a no-op.
func Purge(hours int) error {
	if hours <= 0 {
		log.Debug("cache cleaning disabled")
		return nil
	}
	base, ok := Dir()
	if !ok {
		return nil
	}
	maxAge := time.Duration(hours) * time.Hour
	if err := filepath.Walk(base, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err == nil {
				log.Debugf("removed cache file %s", path)
			} else {
				log.WithError(err).Warnf("failed to remove cache file %s", path)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to purge cache: %w", err)
	}
	return nil
}

// Read attempts to read a cached entry.
func Read(subdirs []string, clearKey string) (*Entry, bool) {
	if !Enabled() {
		return nil, false
	}
	p, ok := EntryPath(subdirs, clearKey)
	if !ok {
		return nil, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	b = bytes.TrimSpace(b)
	return &Entry{
		Key:        clearKey,
		EncodedKey: encodeKey(clearKey),
		Path:       p,
		Data:       b,
	}, true
}

// Write stores data for the given key beneath subdirs. Creates directories as needed.
func Write(subdirs []string, clearKey string, data []byte) error {
	if !Enabled() {
		return nil // treat as disabled.
	}
	base, ok := Dir()
	if !ok {
		return nil // treat as disabled.
	}
	encoded := encodeKey(clearKey)
	dir := filepath.Join(append([]string{base}, subdirs...)...)
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	p := filepath.Join(dir, encoded)
	if err := os.WriteFile(p, data, os.FileMode(0o600)); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write to cache: %w", err)
	}
	return nil
}

// encodeKey hashes k with MD5 and returns the hex string.
func encodeKey(k string) string {
	h := md5.New()
	_, _ = h.Write([]byte(k))
	return hex.EncodeToString(h.Sum(nil))
}
