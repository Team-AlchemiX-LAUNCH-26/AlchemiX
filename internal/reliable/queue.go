package reliable

import "sync"

// SafeQueue holds packets when no valid route exists.
type SafeQueue struct {
	mu      sync.Mutex
	packets []*ReliablePacket
	maxSize int
}

// NewSafeQueue creates a queue with the given max size.
func NewSafeQueue(maxSize int) *SafeQueue {
	if maxSize < 1 {
		maxSize = 100
	}
	return &SafeQueue{packets: make([]*ReliablePacket, 0), maxSize: maxSize}
}

// Enqueue adds a packet to the queue. Returns false if queue is full.
func (q *SafeQueue) Enqueue(pkt *ReliablePacket) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.packets) >= q.maxSize {
		return false
	}
	pkt.Status = StatusQueued
	q.packets = append(q.packets, pkt)
	return true
}

// Dequeue removes and returns the highest priority packet.
func (q *SafeQueue) Dequeue() *ReliablePacket {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.packets) == 0 {
		return nil
	}
	pkt := q.packets[0]
	q.packets = q.packets[1:]
	return pkt
}

// Size returns the current queue length.
func (q *SafeQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.packets)
}

// ExpireTTL removes packets whose TTL has reached zero.
func (q *SafeQueue) ExpireTTL() []*ReliablePacket {
	q.mu.Lock()
	defer q.mu.Unlock()
	var expired []*ReliablePacket
	var remaining []*ReliablePacket
	for _, pkt := range q.packets {
		if pkt.TTL <= 0 {
			pkt.Status = StatusExpired
			expired = append(expired, pkt)
		} else {
			remaining = append(remaining, pkt)
		}
	}
	q.packets = remaining
	return expired
}

// DecrementTTL reduces TTL on all queued packets by 1.
func (q *SafeQueue) DecrementTTL() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, pkt := range q.packets {
		pkt.TTL--
	}
}

// DrainAll removes and returns all packets (for recovery after routes restore).
func (q *SafeQueue) DrainAll() []*ReliablePacket {
	q.mu.Lock()
	defer q.mu.Unlock()
	all := q.packets
	q.packets = make([]*ReliablePacket, 0)
	return all
}
