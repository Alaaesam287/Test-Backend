package limiter

import (
	"sync"
	"time"
)

type clientLimiter struct {
	bucket   *TokenBucket
	lastSeen time.Time
}

type Manager struct {
	clients map[string]*clientLimiter
	mu      sync.Mutex

	rps     int
	burst   int
	cleanup time.Duration
}

func NewManager(rps, burst int, cleanup time.Duration) *Manager {
	m := &Manager{
		clients: make(map[string]*clientLimiter),
		rps:     rps,
		burst:   burst,
		cleanup: cleanup,
	}

	go m.cleanupLoop()
	return m
}

func (m *Manager) GetBucket(key string) *TokenBucket {
	m.mu.Lock()
	defer m.mu.Unlock()

	cl, exists := m.clients[key]
	if !exists {
		b := NewTokenBucket(m.rps, m.burst)
		m.clients[key] = &clientLimiter{
			bucket:   b,
			lastSeen: time.Now(),
		}
		return b
	}

	cl.lastSeen = time.Now()
	return cl.bucket
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.cleanup)
	for range ticker.C {
		m.mu.Lock()
		for key, cl := range m.clients {
			if time.Since(cl.lastSeen) > m.cleanup {
				delete(m.clients, key)
			}
		}
		m.mu.Unlock()
	}
}
