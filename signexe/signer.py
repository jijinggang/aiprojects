import subprocess
from typing import Generator

from config import AppConfig


class SigningError(Exception):
    def __init__(self, exit_code: int, stderr: str):
        self.exit_code = exit_code
        self.stderr = stderr
        super().__init__(f"signtool exited with code {exit_code}")


def build_signtool_command(
    exe_path: str,
    cert_path: str,
    password: str,
    config: AppConfig,
) -> list[str]:
    cmd = [
        config.signtool_exe, "sign",
        "/f", cert_path,
        "/p", password,
        "/fd", "SHA256",
        "/tr", config.signing.timestamper_url,
        "/td", "SHA256",
        "/v",
    ]
    flags = config.signing.additional_flags.strip()
    if flags:
        cmd.insert(2, flags)
    cmd.append(exe_path)
    return cmd


def sign_executable(
    exe_path: str,
    cert_path: str,
    password: str,
    config: AppConfig,
) -> Generator[str, None, None]:
    cmd = build_signtool_command(exe_path, cert_path, password, config)

    process = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )

    assert process.stdout is not None
    collected: list[str] = []
    for line in iter(process.stdout.readline, ""):
        line = line.rstrip("\n\r")
        if line:
            collected.append(line)
        yield line

    returncode = process.wait()
    if returncode != 0:
        raise SigningError(returncode, "\n".join(collected))
