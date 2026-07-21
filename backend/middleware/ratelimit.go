package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	RefillRate float64
	Tokens     float64
	buckets    map[string]*Bucket
	bucketsMu  sync.Mutex
}

func (r *RateLimiter) getBucket(key string) *Bucket {
	r.bucketsMu.Lock()
	defer r.bucketsMu.Unlock()

	if b, ok := r.buckets[key]; ok {
		return b
	}
	b := &Bucket{
		Tokens:     r.Tokens,
		LastRefill: time.Now(),
		Capacity:   10,
		RefillRate: r.RefillRate,
	}
	r.buckets[key] = b
	return b
}

func (r *RateLimiter) RateLimit(server *Server) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			b := r.getBucket(clientIP(req))

			if !b.allow() {
				http.Error(res, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next(res, req)
		}
	}
}

func CreateRateLimiter(refillRate float64, tokens float64) *RateLimiter {
	return &RateLimiter{
		Tokens:     tokens,
		RefillRate: refillRate,
		buckets:    make(map[string]*Bucket),
	}
}

func clientIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		return req.RemoteAddr
	}

	return host
}
