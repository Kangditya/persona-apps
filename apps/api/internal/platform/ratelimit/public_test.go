package ratelimit

import (
    "strconv"
    "strings"
    "sync"
    "testing"
    "time"
)

func TestPublicLimiterAllowsBurstAndRecovers(t *testing.T) {
    now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
    limiter, err := newPublicLimiter(60, 2, func() time.Time { return now })
    if err != nil {
        t.Fatal(err)
    }

    for attempt := 0; attempt < 2; attempt++ {
        allowed, retryAfter := limiter.Allow("192.0.2.1")
        if !allowed || retryAfter != 0 {
            t.Fatalf("attempt %d = allowed %t, retry after %s", attempt, allowed, retryAfter)
        }
    }
    allowed, retryAfter := limiter.Allow("192.0.2.1")
    if allowed || retryAfter <= 0 {
        t.Fatalf("exhausted limiter = allowed %t, retry after %s", allowed, retryAfter)
    }

    now = now.Add(time.Second)
    allowed, retryAfter = limiter.Allow("192.0.2.1")
    if !allowed || retryAfter != 0 {
        t.Fatalf("recovered limiter = allowed %t, retry after %s", allowed, retryAfter)
    }
}

func TestPublicLimiterBoundsAndCleansEntries(t *testing.T) {
    now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
    limiter, err := newPublicLimiter(60, 1, func() time.Time { return now })
    if err != nil {
        t.Fatal(err)
    }
    limiter.maxEntries = 2
    limiter.idleTTL = time.Minute
    limiter.cleanupInterval = time.Minute

    if allowed, _ := limiter.Allow("192.0.2.1"); !allowed {
        t.Fatal("first key was rejected")
    }
    if allowed, _ := limiter.Allow("192.0.2.2"); !allowed {
        t.Fatal("second key was rejected")
    }
    if allowed, _ := limiter.Allow("192.0.2.3"); allowed {
        t.Fatal("unseen key bypassed the entry bound")
    }
    if len(limiter.entries) != 2 {
        t.Fatalf("entries = %d, want 2", len(limiter.entries))
    }

    now = now.Add(2 * time.Minute)
    if allowed, _ := limiter.Allow("192.0.2.3"); !allowed {
        t.Fatal("idle entries were not cleaned before admitting a new key")
    }
    if len(limiter.entries) != 1 {
        t.Fatalf("entries after cleanup = %d, want 1", len(limiter.entries))
    }

    oversizedKey := strings.Repeat("x", maxKeyLength+1)
    if allowed, _ := limiter.Allow(oversizedKey); !allowed {
        t.Fatal("oversized key was unexpectedly rejected")
    }
    if _, found := limiter.entries[""]; !found {
        t.Fatal("oversized key was retained instead of using the bounded fallback key")
    }
}

func TestNewPublicLimiterRejectsUnsafeConfiguration(t *testing.T) {
    for _, test := range []struct {
        perMinute int
        burst     int
    }{
        {perMinute: 0, burst: 1},
        {perMinute: 1, burst: 0},
    } {
        if _, err := NewPublicLimiter(test.perMinute, test.burst); err == nil {
            t.Fatalf("NewPublicLimiter(%d, %d) succeeded", test.perMinute, test.burst)
        }
    }
}

func TestPublicLimiterConcurrentAccess(t *testing.T) {
    limiter, err := NewPublicLimiter(10_000, 10_000)
    if err != nil {
        t.Fatal(err)
    }
    var group sync.WaitGroup
    for worker := 0; worker < 32; worker++ {
        group.Add(1)
        go func(worker int) {
            defer group.Done()
            if allowed, _ := limiter.Allow("192.0.2." + strconv.Itoa(worker+1)); !allowed {
                t.Errorf("worker %d was unexpectedly rate limited", worker)
            }
        }(worker)
    }
    group.Wait()
}
