// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/apex/log"
)

// TODO Doesn't belong in this package.
// THINK Needs to take a CacheEntry.
func Hitter(be *BackendRemote, url string) (bytes.Buffer, error) {

	if err := PurgeCache(); err != nil {
		log.Warnf("failed to purge cache: %w", errors.New("TODO"))
	}

	if entry, ok := CacheReader(be, url); ok {
		log.Debugf("cache hit: %s", entry.Path)
		return *bytes.NewBuffer(entry.Data), nil
	}

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to create request: %w", err)
	}

	//nolint:forcetypeassert
	req.Header.Set("Authorization", "Bearer "+be.Backend.Config.Token.(string))

	http := &http.Client{}
	resp, err := http.Do(req)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var doc bytes.Buffer
	if _, err := doc.ReadFrom(resp.Body); err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to read response: %w", err)
	}

	if err := CacheWriter(be, url, doc.Bytes()); err != nil {
		log.Warnf("failed to write state to cache: %w", err)
	}

	return doc, nil
}
