package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Recorder struct {
	mu       sync.Mutex
	counters map[string]int64
	gauges   map[string]int64
}

func NewRecorder() *Recorder {
	return &Recorder{counters: make(map[string]int64), gauges: make(map[string]int64)}
}

func (r *Recorder) Inc(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[key(name, labels)]++
}

func (r *Recorder) SetGauge(name string, labels map[string]string, value int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[key(name, labels)] = value
}

func (r *Recorder) Text() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	lines := make([]string, 0, len(r.counters)+len(r.gauges))
	for k, v := range r.counters {
		lines = append(lines, fmt.Sprintf("%s %d", k, v))
	}
	for k, v := range r.gauges {
		lines = append(lines, fmt.Sprintf("%s %d", k, v))
	}
	sort.Strings(lines)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func key(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	names := make([]string, 0, len(labels))
	for k := range labels {
		names = append(names, k)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, k := range names {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
	}
	return name + "{" + strings.Join(parts, ",") + "}"
}
