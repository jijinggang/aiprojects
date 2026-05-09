import zipfile

import pytest

from zip_handler import (
    extract_zip,
    find_signable_files,
    repack_zip,
    cleanup_temp_dir,
    ZipProcessingError,
)


def _create_test_zip(zip_path: str, files: dict[str, str]) -> None:
    """Helper: create a ZIP with given {arcname: content} mapping."""
    with zipfile.ZipFile(zip_path, "w", zipfile.ZIP_DEFLATED) as zf:
        for name, content in files.items():
            zf.writestr(name, content)


def test_extract_zip(tmp_path):
    """解压 ZIP 到目标目录"""
    zip_path = tmp_path / "test.zip"
    _create_test_zip(str(zip_path), {
        "setup.exe": "fake exe",
        "helper.dll": "fake dll",
        "readme.txt": "install instructions",
    })
    dest = tmp_path / "extracted"

    result = extract_zip(str(zip_path), str(dest))

    assert result == str(dest)
    assert (dest / "setup.exe").read_text() == "fake exe"
    assert (dest / "helper.dll").read_text() == "fake dll"
    assert (dest / "readme.txt").read_text() == "install instructions"


def test_extract_zip_nested_dirs(tmp_path):
    """解压包含子目录的 ZIP，目录结构保留"""
    zip_path = tmp_path / "nested.zip"
    _create_test_zip(str(zip_path), {
        "subdir/app.exe": "app",
        "subdir/lib/helper.dll": "lib",
    })
    dest = tmp_path / "nested_out"

    extract_zip(str(zip_path), str(dest))

    assert (dest / "subdir" / "app.exe").exists()
    assert (dest / "subdir" / "lib" / "helper.dll").exists()


def test_extract_zip_not_a_zip(tmp_path):
    """非 ZIP 文件抛出 ZipProcessingError"""
    bad_path = tmp_path / "notazip.txt"
    bad_path.write_text("hello")

    with pytest.raises(ZipProcessingError, match="Invalid ZIP"):
        extract_zip(str(bad_path), str(tmp_path / "out"))


def test_find_signable_files(tmp_path):
    """找到目录下所有 .exe 和 .dll 文件"""
    root = tmp_path / "files"
    root.mkdir()
    (root / "a.exe").write_text("a")
    (root / "b.dll").write_text("b")
    (root / "c.exe").write_text("c")
    (root / "readme.txt").write_text("txt")
    (root / "config.xml").write_text("xml")

    sub = root / "lib"
    sub.mkdir()
    (sub / "d.dll").write_text("d")

    result = find_signable_files(str(root))

    expected = {
        str(root / "a.exe"),
        str(root / "b.dll"),
        str(root / "c.exe"),
        str(sub / "d.dll"),
    }
    assert set(result) == expected


def test_find_signable_files_none(tmp_path):
    """无签名目标时返回空列表"""
    root = tmp_path / "noexe"
    root.mkdir()
    (root / "readme.txt").write_text("txt")

    result = find_signable_files(str(root))
    assert result == []


def test_find_signable_files_empty_dir(tmp_path):
    """空目录返回空列表"""
    root = tmp_path / "empty"
    root.mkdir()

    result = find_signable_files(str(root))
    assert result == []


def test_repack_zip(tmp_path):
    """将目录重新打包为 ZIP"""
    src = tmp_path / "src"
    src.mkdir()
    (src / "signed.exe").write_text("signed exe content")
    (src / "helper.dll").write_text("signed dll content")
    (src / "readme.txt").write_text("readme")

    zip_out = tmp_path / "output.signed.zip"
    result = repack_zip(str(src), str(zip_out))

    assert result == str(zip_out)
    assert zipfile.is_zipfile(str(zip_out))

    with zipfile.ZipFile(str(zip_out), "r") as zf:
        names = set(zf.namelist())
        assert "signed.exe" in names
        assert "helper.dll" in names
        assert "readme.txt" in names
        assert zf.read("signed.exe") == b"signed exe content"


def test_repack_zip_relative_paths(tmp_path):
    """重新打包后 ZIP 内路径是相对路径"""
    src = tmp_path / "src"
    sub = src / "sub"
    sub.mkdir(parents=True)
    (sub / "x.dll").write_text("x")

    zip_out = tmp_path / "rel.zip"
    repack_zip(str(src), str(zip_out))

    with zipfile.ZipFile(str(zip_out), "r") as zf:
        assert "sub/x.dll" in zf.namelist()


def test_cleanup_temp_dir(tmp_path):
    """清理临时目录"""
    temp = tmp_path / "temp"
    temp.mkdir()
    (temp / "a.exe").write_text("a")
    (temp / "sub").mkdir()
    (temp / "sub" / "b.dll").write_text("b")

    cleanup_temp_dir(str(temp))

    assert not temp.exists()


def test_cleanup_temp_dir_nonexistent(tmp_path):
    """清理不存在的目录不报错"""
    cleanup_temp_dir(str(tmp_path / "ghost"))
