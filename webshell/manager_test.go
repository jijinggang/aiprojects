package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestDir creates a temporary directory with .cmd files for testing
func createTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create simple .cmd files
	files := map[string]string{
		"echo_hello.cmd": "@echo hello\n",
		"fail.cmd":       "@echo failing\nexit /b 1\n",
		"slow.cmd":       "@echo starting\ntimeout /t 2 /nobreak >nul\n@echo done\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Create a non-.cmd file that should be ignored
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore me"), 0644)

	return dir
}

func TestCmdManager_Scan(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 5)

	err := mgr.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 3 .cmd files
	if len(mgr.cmdFiles) != 3 {
		t.Errorf("expected 3 cmd files, got %d", len(mgr.cmdFiles))
	}

	// Check specific files exist
	names := []string{"echo_hello.cmd", "fail.cmd", "slow.cmd"}
	for _, name := range names {
		if _, ok := mgr.cmdFiles[name]; !ok {
			t.Errorf("expected file %s not found", name)
		}
	}

	// Non-.cmd files should not be included
	if _, ok := mgr.cmdFiles["readme.txt"]; ok {
		t.Error("readme.txt should not be in cmd files")
	}
}

func TestCmdManager_ScanNonexistentDir(t *testing.T) {
	mgr := NewCmdManager("E:\\nonexistent_dir_xyz", 5)

	err := mgr.Scan()
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestCmdManager_StartSuccess(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 5)
	mgr.Scan()

	record, err := mgr.Start("echo_hello.cmd")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if record.Status != StatusRunning {
		t.Errorf("expected status running, got %s", record.Status)
	}
	if record.CmdName != "echo_hello.cmd" {
		t.Errorf("expected cmdName echo_hello.cmd, got %s", record.CmdName)
	}

	// Wait for process to complete
	time.Sleep(2 * time.Second)

	lock := mgr.locks["echo_hello.cmd"]
	lock.mu.Lock()
	running := lock.running
	lock.mu.Unlock()

	if running != nil {
		t.Error("expected running=nil after completion")
	}
}

func TestCmdManager_StartLocked(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 5)
	mgr.Scan()

	// Start slow.cmd first (takes 2 seconds)
	_, err := mgr.Start("slow.cmd")
	if err != nil {
		t.Fatalf("first start should succeed: %v", err)
	}

	// Try starting same command again - should be locked
	_, err = mgr.Start("slow.cmd")
	if err == nil {
		t.Fatal("expected CMD_LOCKED error")
	}
}

func TestCmdManager_StartParallel(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 5)
	mgr.Scan()

	// Different commands should be able to run in parallel
	r1, err1 := mgr.Start("echo_hello.cmd")
	r2, err2 := mgr.Start("fail.cmd")

	if err1 != nil || err2 != nil {
		t.Fatalf("both starts should succeed: err1=%v err2=%v", err1, err2)
	}
	if r1.ID == r2.ID {
		t.Error("different executions should have different IDs")
	}

	// Wait for both to finish
	time.Sleep(2 * time.Second)
}

func TestCmdManager_StartCmdNotFound(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 5)
	mgr.Scan()

	_, err := mgr.Start("nonexistent.cmd")
	if err == nil {
		t.Fatal("expected CMD_NOT_FOUND error")
	}
}

func TestCmdManager_HistoryManagement(t *testing.T) {
	dir := createTestDir(t)
	mgr := NewCmdManager(dir, 3) // maxHistory=3
	mgr.Scan()

	// Execute echo_hello.cmd multiple times
	for i := 0; i < 5; i++ {
		_, err := mgr.Start("echo_hello.cmd")
		if err != nil {
			t.Fatalf("start %d failed: %v", i+1, err)
		}
		time.Sleep(1 * time.Second) // wait for completion
	}

	// Should only have 3 history records (oldest 2 evicted)
	mgr.mu.RLock()
	records := mgr.history["echo_hello.cmd"]
	mgr.mu.RUnlock()

	if len(records) > 3 {
		t.Errorf("expected at most 3 history records, got %d", len(records))
	}
}