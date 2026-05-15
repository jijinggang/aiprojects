package main

import (
	"log"
	"sync"
)

// OutputLine represents a single line of process output
type OutputLine struct {
	Seq    int
	Text   string
	Stream string // "stdout" or "stderr"
}

// OutputBuf stores output lines with a max size limit
type OutputBuf struct {
	lines   []OutputLine
	maxSize int
	nextSeq int
	mu      sync.RWMutex
	closed  bool
}

// NewOutputBuf creates a new output buffer with the given max line count
func NewOutputBuf(maxSize int) *OutputBuf {
	return &OutputBuf{
		maxSize: maxSize,
	}
}

// Write appends a new output line and returns its assigned seq number
func (b *OutputBuf) Write(text string, stream string) int {
	b.mu.Lock()
	line := OutputLine{
		Seq:    b.nextSeq,
		Text:   text,
		Stream: stream,
	}
	b.nextSeq++
	if len(b.lines) >= b.maxSize {
		b.lines = b.lines[1:]
	}
	b.lines = append(b.lines, line)
	b.mu.Unlock()
	return line.Seq
}

// Replay returns all lines with seq >= fromSeq
func (b *OutputBuf) Replay(fromSeq int) []OutputLine {
	b.mu.RLock()
	result := make([]OutputLine, 0, len(b.lines))
	for _, l := range b.lines {
		if l.Seq >= fromSeq {
			result = append(result, l)
		}
	}
	b.mu.RUnlock()
	return result
}

// Close marks the buffer as closed
func (b *OutputBuf) Close() {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
}

// IsClosed returns whether the buffer is closed
func (b *OutputBuf) IsClosed() bool {
	b.mu.RLock()
	c := b.closed
	b.mu.RUnlock()
	return c
}

// Subscriber represents a WebSocket client subscribed to output
type Subscriber struct {
	ID string
	Ch chan OutputLine
}

// SubHub manages subscribers for an OutputBuf
type SubHub struct {
	subs map[string]*Subscriber
	mu   sync.RWMutex
	buf  *OutputBuf
}

// NewSubHub creates a new subscriber hub for the given buffer
func NewSubHub(buf *OutputBuf) *SubHub {
	return &SubHub{
		subs: make(map[string]*Subscriber),
		buf:  buf,
	}
}

// Subscribe adds a new subscriber, replays existing output, then subscribes to future output
func (h *SubHub) Subscribe(connID string) (*Subscriber, error) {
	sub := &Subscriber{
		ID: connID,
		Ch: make(chan OutputLine, 64),
	}
	h.mu.Lock()
	h.subs[connID] = sub
	h.mu.Unlock()

	// Replay existing output
	for _, line := range h.buf.Replay(0) {
		select {
		case sub.Ch <- line:
		default:
			log.Printf("订阅者 %s 回放缓冲满，丢弃 seq=%d", connID, line.Seq)
		}
	}
	return sub, nil
}

// Broadcast pushes a new output line to all subscribers (non-blocking)
func (h *SubHub) Broadcast(line OutputLine) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, sub := range h.subs {
		select {
		case sub.Ch <- line:
		default:
			log.Printf("订阅者 %s 缓冲满，丢弃 seq=%d", sub.ID, line.Seq)
		}
	}
}

// Unsubscribe removes a subscriber and closes its channel
func (h *SubHub) Unsubscribe(connID string) {
	h.mu.Lock()
	sub, ok := h.subs[connID]
	if ok {
		close(sub.Ch)
		delete(h.subs, connID)
	}
	h.mu.Unlock()
}

// CloseAll closes all subscriber channels and clears the map
func (h *SubHub) CloseAll() {
	h.mu.Lock()
	for _, sub := range h.subs {
		close(sub.Ch)
	}
	h.subs = make(map[string]*Subscriber)
	h.mu.Unlock()
}