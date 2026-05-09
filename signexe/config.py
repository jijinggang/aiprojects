import glob
import os
import re
import shutil
from dataclasses import dataclass, field

import yaml


@dataclass
class AuthUser:
    username: str
    password: str


@dataclass
class CertificatesConfig:
    passwords: dict[str, str] = field(default_factory=dict)


@dataclass
class SigningConfig:
    timestamper_url: str = "http://timestamp.digicert.com"
    additional_flags: str = ""
    max_file_size_mb: int = 500


@dataclass
class ServerConfig:
    host: str = "127.0.0.1"
    port: int = 7860


@dataclass
class UploadsConfig:
    retention_hours: int = 24


@dataclass
class AppConfig:
    auth_users: list[AuthUser] = field(default_factory=list)
    certificates: CertificatesConfig = field(default_factory=CertificatesConfig)
    signing: SigningConfig = field(default_factory=SigningConfig)
    server: ServerConfig = field(default_factory=ServerConfig)
    uploads: UploadsConfig = field(default_factory=UploadsConfig)
    signtool_exe: str = ""


_SDK_GLOB = "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\*"


def _extract_sdk_version(path: str) -> tuple[int, ...]:
    m = re.search(r"(\d+\.\d+\.\d+\.\d+)", path)
    if not m:
        return (0,)
    return tuple(int(x) for x in m.group(1).split("."))


def _find_signtool() -> str:
    candidates = glob.glob(
        os.path.join(_SDK_GLOB, "x64", "signtool.exe")
    )
    if candidates:
        candidates.sort(key=_extract_sdk_version, reverse=True)
        return candidates[0]

    found = shutil.which("signtool.exe")
    if found:
        return found

    raise RuntimeError(
        "Cannot find signtool.exe. "
        "Install Windows SDK or set signtool_exe in config.yaml. "
        f"Searched: {_SDK_GLOB}, PATH"
    )


def load_config(path: str) -> AppConfig:
    if not os.path.exists(path):
        raise FileNotFoundError(f"Config file not found: {path}")

    with open(path, "r", encoding="utf-8") as f:
        raw = yaml.safe_load(f)

    if raw is None:
        raw = {}

    auth_raw = raw.get("auth", {}) or {}
    users = [
        AuthUser(username=u["username"], password=u["password"])
        for u in auth_raw.get("users", [])
    ]

    certs_raw = raw.get("certificates", {}) or {}
    cert_passwords = certs_raw.get("passwords", {}) or {}

    signing_raw = raw.get("signing", {}) or {}
    server_raw = raw.get("server", {}) or {}
    uploads_raw = raw.get("uploads", {}) or {}

    signtool_exe = raw.get("signtool_exe", "")
    if not signtool_exe:
        signtool_exe = _find_signtool()

    return AppConfig(
        auth_users=users,
        certificates=CertificatesConfig(passwords=cert_passwords),
        signing=SigningConfig(
            timestamper_url=signing_raw.get("timestamper_url", "http://timestamp.digicert.com"),
            additional_flags=signing_raw.get("additional_flags", ""),
            max_file_size_mb=signing_raw.get("max_file_size_mb", 500),
        ),
        server=ServerConfig(
            host=server_raw.get("host", "127.0.0.1"),
            port=server_raw.get("port", 7860),
        ),
        uploads=UploadsConfig(
            retention_hours=uploads_raw.get("retention_hours", 24),
        ),
        signtool_exe=signtool_exe,
    )
