package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_DefaultCreatesZip(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "unsigned.exe")

	outputName := filepath.Base(dir) + "_unsigned.zip"

	err := run([]string{"scan_unsigned_files", dir})
	if err != nil {
		t.Fatalf("run 返回错误: %v", err)
	}

	if _, err := os.Stat(outputName); os.IsNotExist(err) {
		t.Errorf("预期生成 ZIP 文件 %s，但未找到", outputName)
	}
	os.Remove(outputName)
}

func TestRun_OutputFlag(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "unsigned.exe")

	zipPath := filepath.Join(dir, "custom.zip")
	err := run([]string{"scan_unsigned_files", "-o", zipPath, dir})
	if err != nil {
		t.Fatalf("run 返回错误: %v", err)
	}

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Errorf("预期生成 ZIP 文件 %s，但未找到", zipPath)
	}
}

func TestRun_ListMode(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "unsigned.exe")

	err := run([]string{"scan_unsigned_files", "--list", dir})
	if err != nil {
		t.Fatalf("run 返回错误: %v", err)
	}

	files, _ := filepath.Glob("*.zip")
	if len(files) > 0 {
		t.Errorf("--list 模式不应生成 ZIP 文件，但发现: %v", files)
	}
}

func TestRun_NoUnsignedFiles(t *testing.T) {
	dir := t.TempDir()

	err := run([]string{"scan_unsigned_files", dir})
	if err != nil {
		t.Fatalf("run 返回错误: %v", err)
	}

	files, _ := filepath.Glob("*.zip")
	if len(files) > 0 {
		t.Errorf("没有未签名文件时不应生成 ZIP，但发现: %v", files)
	}
}

func TestRun_MissingDirectory(t *testing.T) {
	err := run([]string{"scan_unsigned_files", "/nonexistent/path"})
	if err == nil {
		t.Error("不存在的目录应返回错误")
	}
	if !strings.Contains(err.Error(), "无法访问") {
		t.Errorf("错误信息应包含'无法访问'，实际: %v", err)
	}
}
