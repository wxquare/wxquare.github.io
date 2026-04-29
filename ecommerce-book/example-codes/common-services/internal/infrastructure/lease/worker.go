package lease

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Store interface {
	AcquireWorker(ctx context.Context, regionID int64, datacenterCode, instanceID string, ttl time.Duration) (workerID int64, leaseToken string, err error)
	RenewWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string, ttl time.Duration) error
	ReleaseWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string) error
}

type Manager struct {
	mu             sync.RWMutex
	store          Store
	regionID       int64
	datacenterCode string
	instanceID     string
	workerID       int64
	leaseToken     string
	ttl            time.Duration
	heartbeatEvery time.Duration
	ready          bool
	stop           chan struct{}
	stopOnce       sync.Once
}

func NewManager(store Store, regionID int64, datacenterCode, instanceID string, ttl, heartbeatEvery time.Duration) *Manager {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	if heartbeatEvery <= 0 {
		heartbeatEvery = 10 * time.Second
	}
	return &Manager{
		store: store, regionID: regionID, datacenterCode: datacenterCode,
		instanceID: instanceID, workerID: -1, ttl: ttl, heartbeatEvery: heartbeatEvery,
		stop: make(chan struct{}),
	}
}

func (m *Manager) Ready() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ready
}

func (m *Manager) WorkerID() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.workerID
}

func (m *Manager) RegionID() int64 {
	return m.regionID
}

func (m *Manager) Start(ctx context.Context) error {
	workerID, token, err := m.store.AcquireWorker(ctx, m.regionID, m.datacenterCode, m.instanceID, m.ttl)
	if err != nil {
		m.markNotReady()
		return err
	}
	m.mu.Lock()
	m.workerID = workerID
	m.leaseToken = token
	m.ready = true
	m.mu.Unlock()
	go m.heartbeat()
	return nil
}

func (m *Manager) Stop(ctx context.Context) {
	m.stopOnce.Do(func() {
		close(m.stop)
	})
	m.mu.RLock()
	workerID, token := m.workerID, m.leaseToken
	m.mu.RUnlock()
	if workerID >= 0 && token != "" {
		_ = m.store.ReleaseWorker(ctx, m.regionID, workerID, m.instanceID, token)
	}
	m.markNotReady()
}

func (m *Manager) heartbeat() {
	ticker := time.NewTicker(m.heartbeatEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			workerID, token := m.workerID, m.leaseToken
			m.mu.RUnlock()
			if err := m.store.RenewWorker(context.Background(), m.regionID, workerID, m.instanceID, token, m.ttl); err != nil {
				m.markNotReady()
			}
		case <-m.stop:
			return
		}
	}
}

func (m *Manager) markNotReady() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ready = false
}

func NewLeaseToken() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf[:])
}
