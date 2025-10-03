// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package remote

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/staranto/tfctlgo/internal/config"
)

// CacheEntry represents a single entry in the cache.
type CacheEntry struct {
	// Key is the clear-text key used to identify the cache entry. For example,
	// the state version id or the HSDU.
	Key string
	// EncodedKey is the encoded version of the key, used as the filename in the
	// cache directory.
	EncodedKey string
	// Path is the full path to the cache entry on disk.
	Path string
	// Data is the raw data stored in the cache.
	Data []byte
}

// CacheEntryPath returns the path to the cache entry for the given key, if it
// exists. The cache is organized first by the backend hostname
// (app.terraform.io) and then by the organization name. The key is hashed and
// used as the filename.
func CacheEntryPath(be *BackendRemote, key string) (string, bool) {

	var cacheDir string
	if c, ok := os.LookupEnv("TFCTL_CACHE_DIR"); ok {
		cacheDir = c
	} else {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "tfctl")
	}

	hostname, organization := getOverrides(be)

	cacheFilePath := filepath.Join(cacheDir, hostname, organization, key)
	if _, err := os.Stat(cacheFilePath); err == nil {
		return cacheFilePath, true
	}

	return "", false
}

// CacheReader reads the cache entry for the given key, if it exists. If the
// cache is disabled, or the entry does not exist, the second return value will
// be false.
func CacheReader(be *BackendRemote, key string) (*CacheEntry, bool) {
	if !isCacheEnabled() {
		return nil, false
	}

	cacheKey := encodeKey(key)

	// IF the cache entry exists, read it and slam it into a CacheEntry.
	if cacheFilePath, ok := CacheEntryPath(be, cacheKey); ok {
		data, err := os.ReadFile(cacheFilePath)
		if err != nil {
			return nil, false
		}
		data = bytes.TrimSpace(data)

		entry := &CacheEntry{
			Key:        key,
			EncodedKey: cacheKey,
			Data:       data,
			Path:       cacheFilePath,
		}

		return entry, true
	}

	// Cache entry doesn't exist yet.
	return nil, false
}

func CacheWriter(be *BackendRemote, key string, data []byte) error {
	if !isCacheEnabled() {
		return nil
	}

	hostname, organization := getOverrides(be)

	// Grab the cache directory from the environment, and fail back to the
	// default.
	cacheDir := os.Getenv("TFCTL_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "tfctl")
	}

	// Make sure the entire cache directory tree exists
	if err := os.MkdirAll(filepath.Join(cacheDir, hostname, organization), 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheKey := encodeKey(key)

	cacheFilePath := filepath.Join(cacheDir, hostname, organization, cacheKey)
	if err := os.WriteFile(cacheFilePath, data, os.FileMode(0o600)); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write to cache: %w", err)
	}

	return nil
}

// encodeKey returns the MD5 hash of the given key, encoded as a hex string.
func encodeKey(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func getOverrides(be *BackendRemote) (hostname, organization string) {
	hostname = be.Backend.Config.Hostname
	if h, ok := os.LookupEnv("TFE_HOSTNAME"); ok {
		hostname = h
	}

	organization = be.Backend.Config.Organization
	if org, ok := os.LookupEnv("TFE_ORGANIZATION"); ok {
		organization = org
	}

	return
}

// isCacheEnabled returns true if the cache is enabled. By default it is, and
// can only be disabled by setting the TFCTL_CACHE environment variable to "0"
// or "false".
func isCacheEnabled() bool {
	enabled, _ := os.LookupEnv("TFCTL_CACHE")
	return enabled == "" || (enabled != "0" && enabled != "false")
}

func PurgeCache() error {
	var cacheDir string
	if c, ok := os.LookupEnv("TFCTL_CACHE_DIR"); ok {
		cacheDir = c
	} else {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "tfctl")
	}

	cleanHours, _ := config.GetInt("cache.clean")
	if cleanHours <= 0 {
		log.Debug("cache cleaning disabled")
		return nil
	}

	if err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && time.Since(info.ModTime()) > time.Duration(cleanHours)*time.Hour {
			if err := os.Remove(path); err == nil {
				log.Debugf("removed cache file %s", path)
			} else {
				log.WithError(err).Warnf("failed to remove cache file %s", path)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to write workspace ID to cache: %w", err)
	}

	return nil
}
