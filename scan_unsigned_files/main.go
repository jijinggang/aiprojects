package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("scan_unsigned_files", flag.ExitOnError)
	output := fs.String("o", "", "输出 ZIP 文件路径")
	listMode := fs.Bool("list", false, "仅列出未签名文件，不生成 ZIP")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s [flags] <目录>\n\n", fs.Name())
		fmt.Fprintln(os.Stderr, "扫描目录中所有未签名的 .exe/.dll 文件，打包为 ZIP 或列出清单。")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("请指定要扫描的目录")
	}

	dir := fs.Arg(0)

	files, err := Scan(dir, IsSigned)
	if err != nil {
		return err
	}

	if *listMode {
		printFileList(files)
		return nil
	}

	if *output == "" {
		*output = filepath.Base(dir) + "_unsigned.zip"
	}

	if err := CreateZip(files, *output); err != nil {
		return err
	}

	if len(files) > 0 {
		fmt.Printf("已创建 %s (%d 个未签名文件)\n", *output, len(files))
	}
	return nil
}

func printFileList(files []FileInfo) {
	if len(files) == 0 {
		fmt.Println("未发现未签名文件")
		return
	}

	fmt.Printf("\n未签名文件列表 (共 %d 个)\n", len(files))
	fmt.Println("════════════════════════════════════════════════════")

	for _, f := range files {
		kb := float64(f.Size) / 1024.0
		fmt.Printf("%-50s %8.1f KB    %s\n", f.RelPath, kb, f.Reason)
	}
	fmt.Println("════════════════════════════════════════════════════")
}
