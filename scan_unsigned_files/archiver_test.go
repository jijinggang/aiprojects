package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateZip_PreservesStructure(t *testing.T) {
	dir := t.TempDir()

	// 创建真实的源文件
	os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755)
	writeTestFile(t, dir, "root.exe", "root content")
	writeTestFile(t, filepath.Join(dir, "sub"), "mid.dll", "mid content")
	writeTestFile(t, filepath.Join(dir, "sub", "deep"), "deep.exe", "deep content")

	files := []FileInfo{
		{Path: filepath.Join(dir, "root.exe"), RelPath: "root.exe", Size: 12, Reason: "无数字签名"},
		{Path: filepath.Join(dir, "sub", "mid.dll"), RelPath: "sub/mid.dll", Size: 11, Reason: "证书已过期"},
		{Path: filepath.Join(dir, "sub", "deep", "deep.exe"), RelPath: "sub/deep/deep.exe", Size: 12, Reason: "证书已吊销"},
	}

	zipPath := filepath.Join(dir, "output.zip")
	err := CreateZip(files, zipPath)
	if err != nil {
		t.Fatalf("CreateZip 返回错误: %v", err)
	}

	// 验证 ZIP 内容
	verifyZip(t, zipPath, map[string]string{
		"root.exe":           "root content",
		"sub/mid.dll":        "mid content",
		"sub/deep/deep.exe":  "deep content",
	})
}

func TestCreateZip_EmptyFiles(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "empty.zip")

	err := CreateZip(nil, zipPath)
	if err != nil {
		t.Fatalf("空文件列表不应报错，但 CreateZip 返回: %v", err)
	}

	// 空列表不应创建 ZIP 文件
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Error("空文件列表不应创建 ZIP 文件")
	}
}

func verifyZip(t *testing.T, zipPath string, expected map[string]string) {
	t.Helper()

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("无法打开 ZIP: %v", err)
	}
	defer r.Close()

	found := make(map[string]bool)
	for _, f := range r.File {
		content, ok := expected[f.Name]
		if !ok {
			t.Errorf("ZIP 中包含未预期的文件: %s", f.Name)
			continue
		}
		found[f.Name] = true

		rc, err := f.Open()
		if err != nil {
			t.Errorf("无法读取 ZIP 条目 %s: %v", f.Name, err)
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Errorf("读取 ZIP 条目 %s 失败: %v", f.Name, err)
			continue
		}
		if string(data) != content {
			t.Errorf("文件 %s 内容期望 %q，实际 %q", f.Name, content, string(data))
		}
	}

	for name := range expected {
		if !found[name] {
			t.Errorf("ZIP 中缺少预期文件: %s", name)
		}
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
