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
		capacity = 2880 // Default to 24 hours at 30s intervals
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
	systemBuffer:     NewRingBuffer(2880), // 24h of 30s snapshots
	containerBuffers: make(map[string]*RingBuffer),
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
		buf = NewRingBuffer(2880) // 24h
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

