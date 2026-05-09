import os
import shutil
import zipfile


class ZipProcessingError(Exception):
    pass


def extract_zip(zip_path: str, dest_dir: str) -> str:
    os.makedirs(dest_dir, exist_ok=True)
    try:
        with zipfile.ZipFile(zip_path, "r") as zf:
            zf.extractall(dest_dir)
    except (zipfile.BadZipFile, zipfile.LargeZipFile) as e:
        raise ZipProcessingError(f"Invalid ZIP file: {e}") from e
    return dest_dir


def find_signable_files(root_dir: str) -> list[str]:
    signable_exts = {".exe", ".dll"}
    result: list[str] = []
    for dirpath, _dirnames, filenames in os.walk(root_dir):
        for name in filenames:
            ext = os.path.splitext(name)[1].lower()
            if ext in signable_exts:
                result.append(os.path.join(dirpath, name))
    return result


def repack_zip(source_dir: str, output_path: str) -> str:
    with zipfile.ZipFile(output_path, "w", zipfile.ZIP_DEFLATED) as zf:
        for dirpath, _dirnames, filenames in os.walk(source_dir):
            for name in filenames:
                full = os.path.join(dirpath, name)
                arcname = os.path.relpath(full, source_dir)
                zf.write(full, arcname)
    return output_path


def cleanup_temp_dir(dir_path: str) -> None:
    if os.path.isdir(dir_path):
        shutil.rmtree(dir_path, ignore_errors=True)
