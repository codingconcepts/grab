package models

import (
	"net/http"
	"strconv"
	"time"
)

// ParseRateLimitResetTime parses a non-200 response's headers and if it contains
// fields which would suggest a user is being rate-limted, it'll tell them when
// they can try again.
func ParseRateLimitResetTime(h http.Header) (time.Duration, bool) {
	rateLimitResetRaw := h["X-Ratelimit-Reset"][0]
	if rateLimitResetRaw == "" {
		return 0, false
	}

	rateLimitReset, err := strconv.ParseInt(rateLimitResetRaw, 10, 64)
	if err != nil {
		return 0, false
	}

	now := time.Now().UTC()
	return time.Unix(rateLimitReset, 0).Sub(now), true
}
