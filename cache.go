package geckoterminal

import (
	"container/list"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Cache is an optional HTTP response cache installed via WithCache.
//
// Get returns the stored entry for a key. Implementations must populate
// CacheEntry.ExpiresAt so the client can decide freshness; expired
// entries are still returned so their ETag can be revalidated.
// Set stores an entry for at most ttl.
type Cache interface {
	Get(key string) (CacheEntry, bool)
	Set(key string, entry CacheEntry, ttl time.Duration)
}

// CacheEntry is a cached HTTP response.
type CacheEntry struct {
	// Body is the raw response body.
	Body []byte
	// Header is a copy of the response headers.
	Header http.Header
	// ETag is the response ETag, used for revalidation.
	ETag string
	// ExpiresAt is populated by the cache implementation when the entry
	// is returned from Get.
	ExpiresAt time.Time
}

// fresh reports whether the entry is still within its TTL.
func (e CacheEntry) fresh(now time.Time) bool {
	return now.Before(e.ExpiresAt)
}

// MemoryCache is a bounded in-memory LRU Cache. It is safe for
// concurrent use.
type MemoryCache struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element
	now      func() time.Time
}

type memoryCacheItem struct {
	key       string
	entry     CacheEntry
	expiresAt time.Time
}

// NewMemoryCache returns an in-memory LRU cache holding at most
// capacity entries (256 when capacity <= 0).
func NewMemoryCache(capacity int) *MemoryCache {
	if capacity <= 0 {
		capacity = 256
	}
	return &MemoryCache{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[string]*list.Element),
		now:      time.Now,
	}
}

// Get returns the cached entry for key, or false when absent. Expired
// entries are returned so their ETag may be revalidated.
func (c *MemoryCache) Get(key string) (CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return CacheEntry{}, false
	}
	c.ll.MoveToFront(el)
	item := el.Value.(*memoryCacheItem)
	entry := item.entry
	entry.ExpiresAt = item.expiresAt
	return entry, true
}

// Set stores the entry under key with the given TTL. Non-positive TTLs
// are ignored.
func (c *MemoryCache) Set(key string, entry CacheEntry, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	item := &memoryCacheItem{
		key:       key,
		entry:     entry,
		expiresAt: c.now().Add(ttl),
	}
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		el.Value = item
		return
	}
	c.items[key] = c.ll.PushFront(item)
	for c.ll.Len() > c.capacity {
		back := c.ll.Back()
		if back == nil {
			break
		}
		delete(c.items, back.Value.(*memoryCacheItem).key)
		c.ll.Remove(back)
	}
}

// cacheKey derives a cache key from the API version and the full
// request URL, which already contains the host (public vs Pro) and the
// sorted query string.
func cacheKey(apiVersion, url string) string {
	return apiVersion + " " + url
}

// cacheTTL extracts a TTL from Cache-Control. s-maxage wins over
// max-age; defaultTTL applies when neither is present. A negative
// result means the response must not be stored.
func cacheTTL(h http.Header, defaultTTL time.Duration) time.Duration {
	cc := h.Get("Cache-Control")
	if cc == "" {
		return defaultTTL
	}
	for _, dir := range strings.Split(cc, ",") {
		dir = strings.TrimSpace(strings.ToLower(dir))
		switch {
		case dir == "no-store" || dir == "no-cache" || dir == "private":
			return -1
		case strings.HasPrefix(dir, "s-maxage="):
			if secs, err := strconv.Atoi(strings.TrimPrefix(dir, "s-maxage=")); err == nil {
				return time.Duration(secs) * time.Second
			}
		}
	}
	for _, dir := range strings.Split(cc, ",") {
		dir = strings.TrimSpace(strings.ToLower(dir))
		if strings.HasPrefix(dir, "max-age=") {
			if secs, err := strconv.Atoi(strings.TrimPrefix(dir, "max-age=")); err == nil {
				return time.Duration(secs) * time.Second
			}
		}
	}
	return defaultTTL
}
