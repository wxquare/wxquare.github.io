package snowflake

import (
	"context"
	"sync"
	"time"

	"common-services/internal/idgen"
)

const (
	regionBits   = 5
	workerBits   = 5
	sequenceBits = 12
	maxRegion    = int64(1<<regionBits - 1)
	maxWorker    = int64(1<<workerBits - 1)
	maxSequence  = int64(1<<sequenceBits - 1)
	workerShift  = sequenceBits
	regionShift  = workerBits + sequenceBits
	timeShift    = regionBits + workerBits + sequenceBits
)

type Config struct {
	Epoch           time.Time
	MaxWaitRollback time.Duration
}

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type Lease interface {
	Ready() bool
	WorkerID() int64
	RegionID() int64
}

type RealClock struct{}

func (RealClock) Now() time.Time          { return time.Now() }
func (RealClock) Sleep(d time.Duration)   { time.Sleep(d) }
func DefaultEpoch() time.Time             { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
func MaxWorkerID() int64                  { return maxWorker }
func MaxRegionID() int64                  { return maxRegion }
func MaxSequencePerMillisecond() int64    { return maxSequence + 1 }
func TimestampShiftForDiagnostics() uint8 { return timeShift }
func RegionShiftForDiagnostics() uint8    { return regionShift }
func WorkerShiftForDiagnostics() uint8    { return workerShift }

type Generator struct {
	mu            sync.Mutex
	cfg           Config
	clock         Clock
	lease         Lease
	lastTimestamp int64
	sequence      int64
}

func NewGenerator(cfg Config, clock Clock, lease Lease) *Generator {
	if cfg.Epoch.IsZero() {
		cfg.Epoch = DefaultEpoch()
	}
	if cfg.MaxWaitRollback == 0 {
		cfg.MaxWaitRollback = 5 * time.Millisecond
	}
	return &Generator{cfg: cfg, clock: clock, lease: lease, lastTimestamp: -1}
}

func (g *Generator) Next(ctx context.Context, ns idgen.NamespaceConfig) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.lease.Ready() {
		return 0, idgen.NewError(idgen.ErrWorkerLeaseLost, ns.Namespace, "worker lease is not ready", true)
	}
	workerID := g.lease.WorkerID()
	regionID := g.lease.RegionID()
	if workerID < 0 || workerID > maxWorker || regionID < 0 || regionID > maxRegion {
		return 0, idgen.NewError(idgen.ErrWorkerLeaseLost, ns.Namespace, "worker or region is out of range", true)
	}

	timestamp := g.timestamp()
	if timestamp < g.lastTimestamp {
		delta := time.Duration(g.lastTimestamp-timestamp) * time.Millisecond
		if delta > g.cfg.MaxWaitRollback {
			return 0, idgen.NewError(idgen.ErrClockRollback, ns.Namespace, "clock moved backwards", false)
		}
		g.clock.Sleep(delta)
		timestamp = g.timestamp()
		if timestamp < g.lastTimestamp {
			return 0, idgen.NewError(idgen.ErrClockRollback, ns.Namespace, "clock moved backwards after wait", false)
		}
	}

	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			timestamp = g.waitNextMillis(g.lastTimestamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = timestamp
	return (timestamp << timeShift) | (regionID << regionShift) | (workerID << workerShift) | g.sequence, nil
}

func (g *Generator) timestamp() int64 {
	return g.clock.Now().Sub(g.cfg.Epoch).Milliseconds()
}

func (g *Generator) waitNextMillis(last int64) int64 {
	ts := g.timestamp()
	for ts <= last {
		g.clock.Sleep(time.Millisecond)
		ts = g.timestamp()
	}
	return ts
}

type FakeClock struct {
	NowValue time.Time
}

func (c *FakeClock) Now() time.Time        { return c.NowValue }
func (c *FakeClock) Sleep(d time.Duration) { c.NowValue = c.NowValue.Add(d) }

type StaticLease struct {
	ReadyValue    bool
	WorkerIDValue int64
	RegionIDValue int64
}

func (l *StaticLease) Ready() bool     { return l.ReadyValue }
func (l *StaticLease) WorkerID() int64 { return l.WorkerIDValue }
func (l *StaticLease) RegionID() int64 { return l.RegionIDValue }
