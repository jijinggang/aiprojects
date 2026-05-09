package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSigned_KnownSignedFile(t *testing.T) {
	// Go 自己的二进制文件通常有嵌入式数字签名
	path := filepath.Join(os.Getenv("SystemRoot"), "explorer.exe")
	if _, err := os.Stat(path); err != nil {
		t.Skip("explorer.exe 不可用，跳过测试")
	}

	signed, reason, err := IsSigned(path)
	if err != nil {
		t.Fatalf("IsSigned 返回错误: %v", err)
	}
	if !signed {
		t.Errorf("explorer.exe 应为已签名文件，但 IsSigned 返回 false (原因: %s)", reason)
	}
}

func TestIsSigned_UnsignedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unsigned.exe")
	if err := os.WriteFile(path, []byte("not a real PE"), 0644); err != nil {
		t.Fatal(err)
	}

	signed, reason, err := IsSigned(path)
	if err != nil {
		t.Fatalf("IsSigned 返回错误: %v", err)
	}
	if signed {
		t.Errorf("非 PE 文件应为未签名，但 IsSigned 返回 true (原因: %s)", reason)
	}
	if reason == "" {
		t.Error("未签名文件应返回原因描述")
	}
}
