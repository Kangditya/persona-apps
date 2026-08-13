package ratelimit

import (
    "errors"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

const (
    defaultMaxEntries      = 10_000
    defaultIdleTTL         = 10 * time.Minute
    defaultCleanupInterval = time.Minute
    maxKeyLength           = 64
    fullLimiterRetryAfter  = time.Minute
)

type PublicLimiter struct {
    mutex           sync.Mutex
    entries         map[string]*entry
    rate            rate.Limit
    burst           int
    maxEntries      int
    idleTTL         time.Duration
    cleanupInterval time.Duration
    lastCleanup     time.Time
    now             func() time.Time
}

type entry struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

func NewPublicLimiter(perMinute, burst int) (*PublicLimiter, error) {
    return newPublicLimiter(perMinute, burst, time.Now)
}

func newPublicLimiter(perMinute, burst int, now func() time.Time) (*PublicLimiter, error) {
    if perMinute < 1 {
        return nil, errors.New("public rate limit per minute must be positive")
    }
    if burst < 1 {
        return nil, errors.New("public rate limit burst must be positive")
    }
    if now == nil {
        return nil, errors.New("public rate limiter clock is required")
    }
    return &PublicLimiter{
        entries:         make(map[string]*entry),
        rate:            rate.Limit(float64(perMinute) / 60),
        burst:           burst,
        maxEntries:      defaultMaxEntries,
        idleTTL:         defaultIdleTTL,
        cleanupInterval: defaultCleanupInterval,
        now:             now,
    }, nil
}

func (limiter *PublicLimiter) Allow(key string) (bool, time.Duration) {
    if len(key) > maxKeyLength {
        key = ""
    }

    now := limiter.now()
    limiter.mutex.Lock()
    defer limiter.mutex.Unlock()

    limiter.cleanup(now)
    current, found := limiter.entries[key]
    if !found {
        if len(limiter.entries) >= limiter.maxEntries {
            return false, fullLimiterRetryAfter
        }
        current = &entry{limiter: rate.NewLimiter(limiter.rate, limiter.burst)}
        limiter.entries[key] = current
    }
    current.lastSeen = now

    reservation := current.limiter.ReserveN(now, 1)
    if !reservation.OK() {
        return false, fullLimiterRetryAfter
    }
    delay := reservation.DelayFrom(now)
    if delay <= 0 {
        return true, 0
    }
    reservation.CancelAt(now)
    return false, delay
}

func (limiter *PublicLimiter) cleanup(now time.Time) {
    if !limiter.lastCleanup.IsZero() && now.Sub(limiter.lastCleanup) < limiter.cleanupInterval && len(limiter.entries) < limiter.maxEntries {
        return
    }
    for key, current := range limiter.entries {
        if now.Sub(current.lastSeen) >= limiter.idleTTL {
            delete(limiter.entries, key)
        }
    }
    limiter.lastCleanup = now
}
