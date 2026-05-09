import os
import time
from pathlib import Path

from utils import save_uploaded_file, cleanup_old_files, cleanup_session_files


def test_save_uploaded_file(tmp_path):
    """保存上传文件到 uploads 目录，返回新路径"""
    upload_dir = tmp_path / "uploads"
    src_file = tmp_path / "test.exe"
    src_file.write_text("fake exe content")

    result = save_uploaded_file(str(src_file), str(upload_dir))

    assert os.path.exists(result)
    assert result.endswith(".exe")
    assert Path(result).parent == upload_dir


def test_save_uploaded_file_creates_dir(tmp_path):
    """upload 目录不存在时应自动创建"""
    upload_dir = tmp_path / "new_uploads"
    src_file = tmp_path / "app.dll"
    src_file.write_text("dll data")

    result = save_uploaded_file(str(src_file), str(upload_dir))

    assert os.path.isdir(upload_dir)
    assert os.path.exists(result)


def test_cleanup_old_files(tmp_path):
    """删除超过保留时间的文件"""
    upload_dir = tmp_path / "uploads"
    upload_dir.mkdir()

    old_file = upload_dir / "old.exe"
    old_file.write_text("old")
    new_file = upload_dir / "new.exe"
    new_file.write_text("new")

    # 设置旧文件 mtime 为 48 小时前
    old_mtime = time.time() - 48 * 3600
    os.utime(str(old_file), (old_mtime, old_mtime))

    cleanup_old_files(str(upload_dir), retention_hours=24)

    assert not old_file.exists()
    assert new_file.exists()


def test_cleanup_old_files_empty_dir(tmp_path):
    """空目录清理不报错"""
    upload_dir = tmp_path / "empty_uploads"
    upload_dir.mkdir()
    cleanup_old_files(str(upload_dir), retention_hours=24)


def test_cleanup_session_files(tmp_path):
    """清理会话的文件列表"""
    files = []
    for name in ["a.exe", "b.signed.exe", "c.dll"]:
        p = tmp_path / name
        p.write_text("data")
        files.append(str(p))

    cleanup_session_files(files)

    for f in files:
        assert not os.path.exists(f)


def test_cleanup_session_files_non_existent():
    """清理不存在的文件不报错"""
    cleanup_session_files(["/nonexistent/file.exe", "/another/fake.dll"])
