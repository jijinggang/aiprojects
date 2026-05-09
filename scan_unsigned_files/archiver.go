package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateZip creates a ZIP file containing the given files, preserving their
// relative directory structure as specified by RelPath.
func CreateZip(files []FileInfo, outputPath string) error {
	if len(files) == 0 {
		fmt.Println("未发现未签名文件")
		return nil
	}

	// 确保输出目录存在
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	for _, f := range files {
		header := &zip.FileHeader{
			Name:   f.RelPath,
			Method: zip.Deflate,
		}

		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("创建 ZIP 条目 %s 失败: %w", f.RelPath, err)
		}

		src, err := os.Open(f.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 跳过 %s: %v\n", f.RelPath, err)
			continue
		}

		_, err = io.Copy(writer, src)
		src.Close()
		if err != nil {
			return fmt.Errorf("写入 ZIP 条目 %s 失败: %w", f.RelPath, err)
		}
	}

	return nil
}
