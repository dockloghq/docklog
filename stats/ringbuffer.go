package stats

import (
	"sync"
	"time"
)

// MetricPoint represents a single point in time for a metric
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	Memory    int64     `json:"memory"`
}

// RingBuffer is a thread-safe circular buffer for storing metric points in memory
type RingBuffer struct {
	capacity int
	points   []MetricPoint
	head     int
	size     int
	mu       sync.RWMutex
}

// NewRingBuffer creates a ring buffer with a fixed maximum capacity
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 240 // 2 hours of 30s snapshots (frontend charts query max 200 points; older metrics persist in SQLite)
	}
	return &RingBuffer{
		capacity: capacity,
		points:   make([]MetricPoint, capacity),
	}
}

// Push appends a new data point, overwriting the oldest point if full
func (rb *RingBuffer) Push(cpu float64, memory int64, ts time.Time) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.points[rb.head] = MetricPoint{
		Timestamp: ts,
		CPU:       cpu,
		Memory:    memory,
	}
	rb.head = (rb.head + 1) % rb.capacity
	if rb.size < rb.capacity {
		rb.size++
	}
}

// Last returns the most recent data point
func (rb *RingBuffer) Last() (MetricPoint, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return MetricPoint{}, false
	}
	lastIdx := (rb.head - 1 + rb.capacity) % rb.capacity
	return rb.points[lastIdx], true
}

// GetSince returns all points recorded after the given time in chronological order
func (rb *RingBuffer) GetSince(since time.Time) []MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return nil
	}

	start := (rb.head - rb.size + rb.capacity) % rb.capacity
	var result []MetricPoint

	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.capacity
		pt := rb.points[idx]
		if pt.Timestamp.After(since) || pt.Timestamp.Equal(since) {
			result = append(result, pt)
		}
	}
	return result
}

// GetRange returns all points between from and to in chronological order
func (rb *RingBuffer) GetRange(from, to time.Time) []MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return nil
	}

	start := (rb.head - rb.size + rb.capacity) % rb.capacity
	var result []MetricPoint

	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.capacity
		pt := rb.points[idx]
		if (pt.Timestamp.After(from) || pt.Timestamp.Equal(from)) &&
			(pt.Timestamp.Before(to) || pt.Timestamp.Equal(to)) {
			result = append(result, pt)
		}
	}
	return result
}

// RingBufferRegistry manages ring buffers for system metrics and individual containers
type RingBufferRegistry struct {
	systemBuffer     *RingBuffer
	containerBuffers map[string]*RingBuffer
	mu               sync.RWMutex
}

var Registry = &RingBufferRegistry{
	systemBuffer:     NewRingBuffer(240),
	containerBuffers: make(map[string]*RingBuffer),
}

// GetAllPoints returns all current points stored in the ring buffer in chronological order
func (rb *RingBuffer) GetAllPoints() []MetricPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return nil
	}

	start := (rb.head - rb.size + rb.capacity) % rb.capacity
	result := make([]MetricPoint, rb.size)
	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.capacity
		result[i] = rb.points[idx]
	}
	return result
}

// RestorePoints bulk-populates points into the ring buffer
func (rb *RingBuffer) RestorePoints(pts []MetricPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for _, pt := range pts {
		rb.points[rb.head] = pt
		rb.head = (rb.head + 1) % rb.capacity
		if rb.size < rb.capacity {
			rb.size++
		}
	}
}

// SnapshotData represents the serialized checkpoint of ring buffers
type SnapshotData struct {
	Timestamp  time.Time                  `json:"timestamp"`
	System     []MetricPoint              `json:"system"`
	Containers map[string][]MetricPoint   `json:"containers"`
}

// ExportSnapshot extracts all points within 24h for serialization
func (r *RingBufferRegistry) ExportSnapshot() SnapshotData {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snap := SnapshotData{
		Timestamp:  time.Now().UTC(),
		System:     r.systemBuffer.GetAllPoints(),
		Containers: make(map[string][]MetricPoint),
	}

	for id, buf := range r.containerBuffers {
		pts := buf.GetAllPoints()
		if len(pts) > 0 {
			snap.Containers[id] = pts
		}
	}

	return snap
}

// RestoreSnapshot restores points from a saved snapshot (filtering out points older than 24h)
func (r *RingBufferRegistry) RestoreSnapshot(snap SnapshotData) {
	cutoff := time.Now().Add(-24 * time.Hour)

	var validSys []MetricPoint
	for _, p := range snap.System {
		if p.Timestamp.After(cutoff) {
			validSys = append(validSys, p)
		}
	}
	if len(validSys) > 0 {
		r.systemBuffer.RestorePoints(validSys)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for id, pts := range snap.Containers {
		var validPts []MetricPoint
		for _, p := range pts {
			if p.Timestamp.After(cutoff) {
				validPts = append(validPts, p)
			}
		}
		if len(validPts) > 0 {
			buf := NewRingBuffer(240)
			buf.RestorePoints(validPts)
			r.containerBuffers[id] = buf
		}
	}
}

func (r *RingBufferRegistry) PushSystem(cpu float64, memory int64, ts time.Time) {
	r.systemBuffer.Push(cpu, memory, ts)
}

func (r *RingBufferRegistry) GetSystemSince(since time.Time) []MetricPoint {
	return r.systemBuffer.GetSince(since)
}

func (r *RingBufferRegistry) PushContainer(containerID string, cpu float64, memory int64, ts time.Time) {
	r.mu.Lock()
	buf, ok := r.containerBuffers[containerID]
	if !ok {
		buf = NewRingBuffer(240)
		r.containerBuffers[containerID] = buf
	}
	r.mu.Unlock()

	buf.Push(cpu, memory, ts)
}

func (r *RingBufferRegistry) GetContainerSince(containerID string, since time.Time) []MetricPoint {
	r.mu.RLock()
	buf, ok := r.containerBuffers[containerID]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	return buf.GetSince(since)
}

func (r *RingBufferRegistry) GetContainerRange(containerID string, from, to time.Time) []MetricPoint {
	r.mu.RLock()
	buf, ok := r.containerBuffers[containerID]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	return buf.GetRange(from, to)
}

func (r *RingBufferRegistry) GetContainerLast(containerID string) (MetricPoint, bool) {
	r.mu.RLock()
	buf, ok := r.containerBuffers[containerID]
	r.mu.RUnlock()
	if !ok {
		return MetricPoint{}, false
	}
	return buf.Last()
}

// PruneRemovedContainers frees ring buffers for containers that have been deleted from Docker.
func (r *RingBufferRegistry) PruneRemovedContainers(existingIDs map[string]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id := range r.containerBuffers {
		if !existingIDs[id] {
			delete(r.containerBuffers, id)
		}
	}
}

