import os
import shutil
import time
from pathlib import Path
from uuid import uuid4


def save_uploaded_file(temp_path: str, upload_dir: str) -> str:
    os.makedirs(upload_dir, exist_ok=True)

    src = Path(temp_path)
    ext = src.suffix
    dest = Path(upload_dir) / f"{uuid4().hex}{ext}"
    shutil.copy2(temp_path, str(dest))
    return str(dest)


def cleanup_old_files(upload_dir: str, retention_hours: int) -> None:
    if not os.path.isdir(upload_dir):
        return

    cutoff = time.time() - retention_hours * 3600
    for entry in os.scandir(upload_dir):
        if entry.is_file():
            try:
                stat = entry.stat()
                if stat.st_mtime < cutoff:
                    os.unlink(entry.path)
            except OSError:
                pass


def cleanup_session_files(file_list: list[str]) -> None:
    for path in file_list:
        try:
            os.unlink(path)
        except OSError:
            pass
