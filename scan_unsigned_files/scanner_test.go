package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScan_FiltersByExtension(t *testing.T) {
	dir := t.TempDir()

	// 创建测试文件结构
	createFile(t, dir, "app.exe")
	createFile(t, dir, "lib.dll")
	createFile(t, dir, "readme.txt")
	createFile(t, dir, "notes.EXE") // 大写扩展名

	mockCheck := func(path string) (bool, string, error) {
		return false, "无数字签名", nil
	}

	files, err := Scan(dir, mockCheck)
	if err != nil {
		t.Fatalf("Scan 返回错误: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("期望 3 个文件 (app.exe, lib.dll, notes.EXE)，实际 %d", len(files))
	}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Path))
		if ext != ".exe" && ext != ".dll" {
			t.Errorf("不应包含非 PE 文件: %s", f.Path)
		}
	}
}

func TestScan_RecursiveWalk(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755)
	createFile(t, dir, "root.exe")
	createFile(t, filepath.Join(dir, "sub"), "mid.dll")
	createFile(t, filepath.Join(dir, "sub", "deep"), "deep.exe")

	allUnsigned := func(path string) (bool, string, error) {
		return false, "无数字签名", nil
	}

	files, err := Scan(dir, allUnsigned)
	if err != nil {
		t.Fatalf("Scan 返回错误: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("期望 3 个文件，实际 %d", len(files))
	}

	// 验证相对路径
	paths := make(map[string]bool)
	for _, f := range files {
		paths[f.RelPath] = true
	}
	expected := []string{"root.exe", "sub/mid.dll", "sub/deep/deep.exe"}
	for _, p := range expected {
		if !paths[p] {
			t.Errorf("缺少预期路径: %s", p)
		}
	}
}

func TestScan_SkipsSignedFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, dir, "good.exe")
	createFile(t, dir, "bad.dll")

	selectiveCheck := func(path string) (bool, string, error) {
		if filepath.Base(path) == "good.exe" {
			return true, "", nil
		}
		return false, "无数字签名", nil
	}

	files, err := Scan(dir, selectiveCheck)
	if err != nil {
		t.Fatalf("Scan 返回错误: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("期望 1 个未签名文件，实际 %d", len(files))
	}
	if files[0].RelPath != "bad.dll" {
		t.Errorf("期望 bad.dll，实际 %s", files[0].RelPath)
	}
}

func TestScan_FileInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.exe")
	content := []byte("PE test content")
	os.WriteFile(path, content, 0644)

	mockCheck := func(path string) (bool, string, error) {
		return false, "证书已过期", nil
	}

	files, err := Scan(dir, mockCheck)
	if err != nil {
		t.Fatalf("Scan 返回错误: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("期望 1 个文件，实际 %d", len(files))
	}

	f := files[0]
	if f.RelPath != "test.exe" {
		t.Errorf("RelPath 期望 test.exe，实际 %s", f.RelPath)
	}
	if f.Size != int64(len(content)) {
		t.Errorf("Size 期望 %d，实际 %d", len(content), f.Size)
	}
	if f.Reason != "证书已过期" {
		t.Errorf("Reason 期望 '证书已过期'，实际 '%s'", f.Reason)
	}
}

func createFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(name+" content"), 0644); err != nil {
		t.Fatal(err)
	}
}
