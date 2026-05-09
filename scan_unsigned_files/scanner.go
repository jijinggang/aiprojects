package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SigChecker verifies a PE file's digital signature.
// Returns (true, "", nil) if validly signed.
type SigChecker func(path string) (bool, string, error)

// FileInfo holds information about an unsigned PE file.
type FileInfo struct {
	Path    string // 绝对路径
	RelPath string // 相对于扫描根目录的路径
	Size    int64  // 文件大小（字节）
	Reason  string // 未签名原因
}

// Scan recursively scans a directory for unsigned .exe/.dll files.
func Scan(root string, check SigChecker) ([]FileInfo, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径: %w", err)
	}

	rootInfo, statRootErr := os.Stat(root)
	if statRootErr != nil {
		return nil, fmt.Errorf("无法访问目录 %s: %w", root, statRootErr)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("%s 不是一个目录", root)
	}

	var files []FileInfo

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法访问 %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".exe" && ext != ".dll" {
			return nil
		}

		signed, reason, sigErr := check(path)
		if sigErr != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法验证 %s: %v\n", path, sigErr)
			return nil
		}
		if signed {
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法获取文件信息 %s: %v\n", path, statErr)
			return nil
		}

		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			relPath = path
		}

		files = append(files, FileInfo{
			Path:    path,
			RelPath: filepath.ToSlash(relPath),
			Size:    info.Size(),
			Reason:  reason,
		})
		return nil
	})

	return files, err
}
