package main

import (
	"testing"
	"time"
)

func drainChannel(ch chan OutputLine, n int) []OutputLine {
	var result []OutputLine
	if n == 0 {
		// Drain all currently available without waiting
		for {
			select {
			case line := <-ch:
				result = append(result, line)
			default:
				return result
			}
		}
	}
	timeout := time.After(2 * time.Second)
	for len(result) < n {
		select {
		case line := <-ch:
			result = append(result, line)
		case <-timeout:
			return result
		}
	}
	return result
}

func tryRecv(ch chan OutputLine) (OutputLine, bool) {
	select {
	case line, ok := <-ch:
		return line, ok
	default:
		return OutputLine{}, true // channel still open but empty
	}
}

func TestOutputBuf_WriteAndReplay(t *testing.T) {
	buf := NewOutputBuf(10000)

	seq1 := buf.Write("line 1\n", "stdout")
	seq2 := buf.Write("error output\n", "stderr")
	seq3 := buf.Write("line 3\n", "stdout")

	if seq1 != 0 || seq2 != 1 || seq3 != 2 {
		t.Fatalf("expected seq 0,1,2, got %d,%d,%d", seq1, seq2, seq3)
	}

	lines := buf.Replay(0)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0].Text != "line 1\n" || lines[0].Stream != "stdout" {
		t.Errorf("line 0 mismatch: got text=%q stream=%q", lines[0].Text, lines[0].Stream)
	}
	if lines[1].Text != "error output\n" || lines[1].Stream != "stderr" {
		t.Errorf("line 1 mismatch: got text=%q stream=%q", lines[1].Text, lines[1].Stream)
	}
	if lines[2].Text != "line 3\n" || lines[2].Stream != "stdout" {
		t.Errorf("line 2 mismatch: got text=%q stream=%q", lines[2].Text, lines[2].Stream)
	}

	lines2 := buf.Replay(2)
	if len(lines2) != 1 {
		t.Fatalf("expected 1 line from seq 2, got %d", len(lines2))
	}
	if lines2[0].Seq != 2 {
		t.Errorf("expected seq 2, got %d", lines2[0].Seq)
	}
}

func TestOutputBuf_MaxSizeEviction(t *testing.T) {
	maxSize := 5
	buf := NewOutputBuf(maxSize)

	for i := 0; i < 10; i++ {
		buf.Write("line "+string(rune('0'+i))+"\n", "stdout")
	}

	lines := buf.Replay(0)
	if len(lines) != maxSize {
		t.Fatalf("expected %d lines after eviction, got %d", maxSize, len(lines))
	}
	if lines[0].Seq != 5 {
		t.Errorf("expected first line seq=5, got %d", lines[0].Seq)
	}
	if lines[maxSize-1].Seq != 9 {
		t.Errorf("expected last line seq=9, got %d", lines[maxSize-1].Seq)
	}
}

func TestSubHub_SubscribeReplay(t *testing.T) {
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)

	buf.Write("existing line 1\n", "stdout")
	buf.Write("existing line 2\n", "stderr")

	sub, err := hub.Subscribe("conn1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	replayed := drainChannel(sub.Ch, 2)
	if len(replayed) != 2 {
		t.Fatalf("expected 2 replayed lines, got %d", len(replayed))
	}
	if replayed[0].Text != "existing line 1\n" {
		t.Errorf("first replayed line mismatch: %q", replayed[0].Text)
	}
	if replayed[1].Text != "existing line 2\n" {
		t.Errorf("second replayed line mismatch: %q", replayed[1].Text)
	}
}

func TestSubHub_Broadcast(t *testing.T) {
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)

	sub1, _ := hub.Subscribe("conn1")
	sub2, _ := hub.Subscribe("conn2")

	drainChannel(sub1.Ch, 0)
	drainChannel(sub2.Ch, 0)

	hub.Broadcast(OutputLine{Seq: 0, Text: "new line\n", Stream: "stdout"})

	line1 := drainChannel(sub1.Ch, 1)
	line2 := drainChannel(sub2.Ch, 1)

	if len(line1) != 1 || line1[0].Text != "new line\n" {
		t.Errorf("sub1 didn't receive broadcast: %v", line1)
	}
	if len(line2) != 1 || line2[0].Text != "new line\n" {
		t.Errorf("sub2 didn't receive broadcast: %v", line2)
	}
}

func TestSubHub_BroadcastSlowClient(t *testing.T) {
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)

	sub, _ := hub.Subscribe("conn1")
	drainChannel(sub.Ch, 0)

	// Fill subscriber's buffered channel (64 capacity)
	for i := 0; i < 64; i++ {
		hub.Broadcast(OutputLine{Seq: i, Text: "fill\n", Stream: "stdout"})
	}
	drainChannel(sub.Ch, 64)

	// Broadcast should not block even if client is slow
	done := make(chan bool)
	go func() {
		hub.Broadcast(OutputLine{Seq: 100, Text: "overflow\n", Stream: "stdout"})
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Broadcast blocked on slow client")
	}
}

func TestSubHub_CloseAll(t *testing.T) {
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)

	sub1, _ := hub.Subscribe("conn1")
	sub2, _ := hub.Subscribe("conn2")

	hub.CloseAll()

	// Closed channels should return (zero, false)
	_, ok1 := <-sub1.Ch
	_, ok2 := <-sub2.Ch

	if ok1 {
		t.Error("sub1.Ch should be closed")
	}
	if ok2 {
		t.Error("sub2.Ch should be closed")
	}
}

func TestSubHub_Unsubscribe(t *testing.T) {
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)

	sub, _ := hub.Subscribe("conn1")
	drainChannel(sub.Ch, 0)

	hub.Unsubscribe("conn1")

	_, ok := <-sub.Ch
	if ok {
		t.Error("channel should be closed after unsubscribe")
	}
}