// SPDX-License-Identifier: BSD-3-Clause
// Copyright 2026 Fyne contributors.

package cache

import (
	"testing"
	"time"
)

func TestExpiringCacheConcurrentRefresh(t *testing.T) {
	entry := &expiringCache{}
	expiresAt := time.Unix(0, cachedNow.Load()).Add(ValidDuration)
	entry.setAlive()

	const workers = 2
	const iterations = 1024
	start := make(chan struct{})
	done := make(chan struct{}, workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			<-start
			for i := 0; i < iterations; i++ {
				entry.setAlive()
			}
			done <- struct{}{}
		}()
	}

	close(start)
	for i := 0; i < iterations; i++ {
		if entry.isExpired(expiresAt.Add(-time.Nanosecond)) || entry.isExpired(expiresAt) {
			t.Error("refresh expired an entry at or before its expiration boundary")
			break
		}
		if !entry.isExpired(expiresAt.Add(time.Nanosecond)) {
			t.Error("entry remained alive after its expiration boundary")
			break
		}
	}
	for worker := 0; worker < workers; worker++ {
		<-done
	}
}
