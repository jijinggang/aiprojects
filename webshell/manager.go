package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ExecStatus represents the status of a command execution
type ExecStatus string

const (
	StatusRunning ExecStatus = "running"
	StatusDone    ExecStatus = "done"
	StatusFailed  ExecStatus = "failed"
)

// CmdFile represents a discovered .cmd file
type CmdFile struct {
	Name     string
	FullName string
	ModTime  time.Time
}

// ExecRecord represents a single execution of a .cmd file
type ExecRecord struct {
	ID        int64      `json:"ID"`
	CmdName   string     `json:"cmdName"`
	Status    ExecStatus `json:"status"`
	StartTime time.Time  `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	ExitCode  int        `json:"exitCode"`
	Output    *OutputBuf `json:"-"`
	Hub       *SubHub    `json:"-"`
}

// CmdLock provides per-.cmd file mutual exclusion
type CmdLock struct {
	mu      sync.Mutex
	running *ExecRecord
}

// CmdManager manages all .cmd files, execution, and history
type CmdManager struct {
	dir        string
	cmdFiles   map[string]*CmdFile
	locks      map[string]*CmdLock
	history    map[string][]*ExecRecord
	maxHistory int
	idCounter  atomic.Int64
	mu         sync.RWMutex
}

// NewCmdManager creates a new CmdManager
func NewCmdManager(dir string, maxHistory int) *CmdManager {
	return &CmdManager{
		dir:        dir,
		cmdFiles:   make(map[string]*CmdFile),
		locks:      make(map[string]*CmdLock),
		history:    make(map[string][]*ExecRecord),
		maxHistory: maxHistory,
	}
}

// Scan discovers .cmd files in the configured directory
func (m *CmdManager) Scan() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".cmd" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		name := entry.Name()
		fullName := filepath.Join(m.dir, name)
		m.cmdFiles[name] = &CmdFile{
			Name:     name,
			FullName: fullName,
			ModTime:  info.ModTime(),
		}
		m.locks[name] = &CmdLock{}
		m.history[name] = nil
	}
	return nil
}

// GetCmdList returns info about all commands for WebSocket response
func (m *CmdManager) GetCmdList() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(m.cmdFiles))
	for name, cf := range m.cmdFiles {
		lock := m.locks[name]
		lock.mu.Lock()
		running := lock.running != nil
		var currentExecId int64
		var watchers int
		if running {
			currentExecId = lock.running.ID
			watchers = len(lock.running.Hub.subs)
		}
		lock.mu.Unlock()

		entry := map[string]interface{}{
			"name":         cf.Name,
			"fullName":     cf.FullName,
			"modTime":      cf.ModTime.Format(time.RFC3339),
			"running":      running,
			"currentExecId": currentExecId,
			"historyCount": len(m.history[name]),
		}
		if watchers > 0 {
			entry["watchers"] = watchers
		}
		result = append(result, entry)
	}
	return result
}

// Start begins execution of a .cmd file, with mutual exclusion
func (m *CmdManager) Start(cmdName string) (*ExecRecord, error) {
	m.mu.RLock()
	cf, ok := m.cmdFiles[cmdName]
	lock := m.locks[cmdName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("CMD_NOT_FOUND: %s 不存在", cmdName)
	}

	lock.mu.Lock()
	if lock.running != nil {
		lock.mu.Unlock()
		return nil, fmt.Errorf("CMD_LOCKED: %s 正在执行中", cmdName)
	}

	id := m.idCounter.Add(1)
	buf := NewOutputBuf(10000)
	hub := NewSubHub(buf)
	record := &ExecRecord{
		ID:        id,
		CmdName:   cmdName,
		Status:    StatusRunning,
		StartTime: time.Now(),
		ExitCode:  -1,
		Output:    buf,
		Hub:       hub,
	}
	lock.running = record
	lock.mu.Unlock()

	// Add to history
	m.mu.Lock()
	m.history[cmdName] = append(m.history[cmdName], record)
	m.mu.Unlock()

	// Start process in goroutine
	go m.runProcess(cmdName, cf.FullName, record, lock)

	return record, nil
}

// runProcess executes cmd.exe /c <file> and manages the lifecycle
func (m *CmdManager) runProcess(cmdName, fullName string, record *ExecRecord, lock *CmdLock) {
	cmd := exec.Command("cmd.exe", "/c", fullName)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	cmd.Start()

	// Read stdout and stderr in separate goroutines
	go m.readPipe(stdoutPipe, "stdout", record)
	go m.readPipe(stderrPipe, "stderr", record)

	cmd.Wait()

	exitCode := cmd.ProcessState.ExitCode()
	now := time.Now()

	// Update record
	record.ExitCode = exitCode
	record.EndTime = &now
	if exitCode == 0 {
		record.Status = StatusDone
	} else {
		record.Status = StatusFailed
	}

	// Release lock
	lock.mu.Lock()
	lock.running = nil
	lock.mu.Unlock()

	// Close output hub
	record.Output.Close()
	record.Hub.CloseAll()

	// Trim history
	m.mu.Lock()
	m.trimHistory(cmdName)
	m.mu.Unlock()
}

func (m *CmdManager) readPipe(pipe io.Reader, stream string, record *ExecRecord) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		seq := record.Output.Write(line, stream)
		record.Hub.Broadcast(OutputLine{Seq: seq, Text: line, Stream: stream})
	}
}

// trimHistory removes oldest completed records if exceeding maxHistory
func (m *CmdManager) trimHistory(cmdName string) {
	records := m.history[cmdName]
	// Count completed records, keep running ones
	completed := 0
	for i := 0; i < len(records); i++ {
		if records[i].Status != StatusRunning {
			completed++
		}
	}
	// Remove oldest completed records beyond maxHistory
	excess := completed - m.maxHistory
	if excess <= 0 {
		return
	}
	newRecords := make([]*ExecRecord, 0, len(records))
	for _, r := range records {
		if r.Status == StatusRunning {
			newRecords = append(newRecords, r)
			continue
		}
		if excess > 0 {
			excess--
			continue // skip (evict)
		}
		newRecords = append(newRecords, r)
	}
	m.history[cmdName] = newRecords
}

// GetHistory returns execution history for a command
func (m *CmdManager) GetHistory(cmdName string) ([]*ExecRecord, error) {
	m.mu.RLock()
	cf, ok := m.cmdFiles[cmdName]
	if !ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("CMD_NOT_FOUND: %s 不存在", cmdName)
	}
	_ = cf
	records := m.history[cmdName]
	m.mu.RUnlock()

	result := make([]*ExecRecord, len(records))
	for i, r := range records {
		result[i] = r
	}
	return result, nil
}

// FindRecord locates a specific execution record by ID
func (m *CmdManager) FindRecord(cmdName string, execId int64) (*ExecRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, r := range m.history[cmdName] {
		if r.ID == execId {
			return r, nil
		}
	}
	return nil, fmt.Errorf("REPLAY_NOT_FOUND: 执行记录 %d 不存在或已被淘汰", execId)
}

// Subscribe subscribes a WebSocket connection to a running command's output
func (m *CmdManager) Subscribe(cmdName, connID string) (*Subscriber, error) {
	m.mu.RLock()
	lock := m.locks[cmdName]
	m.mu.RUnlock()

	lock.mu.Lock()
	if lock.running == nil {
		lock.mu.Unlock()
		return nil, fmt.Errorf("CMD_NOT_RUNNING: %s 当前未在执行", cmdName)
	}
	sub, err := lock.running.Hub.Subscribe(connID)
	lock.mu.Unlock()
	return sub, err
}

// Unsubscribe removes a WebSocket connection from a command's output
func (m *CmdManager) Unsubscribe(cmdName, connID string) {
	m.mu.RLock()
	lock := m.locks[cmdName]
	m.mu.RUnlock()

	lock.mu.Lock()
	if lock.running != nil {
		lock.running.Hub.Unsubscribe(connID)
	}
	lock.mu.Unlock()
}