from unittest import mock

import pytest

from config import AppConfig
from signer import build_signtool_command, sign_executable, SigningError


def _make_config(**kwargs):
    """Helper: create AppConfig with defaults overridden."""
    config = AppConfig()
    config.signtool_exe = "C:\\test\\signtool.exe"
    config.signing.timestamper_url = "http://timestamp.example.com"
    for key, val in kwargs.items():
        parts = key.split(".")
        if len(parts) == 1:
            setattr(config, parts[0], val)
        elif len(parts) == 2:
            setattr(getattr(config, parts[0]), parts[1], val)
    return config


def test_build_signtool_command_basic():
    """基本签名命令构建"""
    config = _make_config()
    cmd = build_signtool_command(
        "test.exe", "cert.pfx", "mypassword", config
    )
    assert cmd[0] == config.signtool_exe
    assert cmd[1] == "sign"
    assert "/f" in cmd
    assert "cert.pfx" in cmd
    assert "/p" in cmd
    assert "mypassword" in cmd
    assert "/fd" in cmd and "SHA256" in cmd
    assert "/tr" in cmd
    assert config.signing.timestamper_url in cmd
    assert "/td" in cmd and "SHA256" in cmd
    assert "/v" in cmd
    assert cmd[-1] == "test.exe"


def test_build_signtool_command_additional_flags():
    """additional_flags 应出现在命令中"""
    config = _make_config(**{"signing.additional_flags": "/as"})
    cmd = build_signtool_command(
        "app.dll", "my.pfx", "pwd", config
    )
    assert "/as" in cmd


def test_build_signtool_command_no_additional_flags():
    """additional_flags 为空时不应出现多余参数"""
    config = _make_config(**{"signing.additional_flags": ""})
    cmd = build_signtool_command(
        "app.dll", "my.pfx", "pwd", config
    )
    # No empty string in command
    assert "" not in cmd


def test_build_signtool_command_order():
    """参数顺序: sign [flags] /f /p /fd /tr /td /v <exe>"""
    config = _make_config(**{"signing.additional_flags": "/as /debug"})
    cmd = build_signtool_command(
        "file.exe", "c.pfx", "p", config
    )
    sign_idx = cmd.index("sign")
    f_idx = cmd.index("/f")
    p_idx = cmd.index("/p")
    fd_idx = cmd.index("/fd")
    exe_idx = len(cmd) - 1

    assert sign_idx < f_idx < p_idx < fd_idx < exe_idx


def test_sign_executable_success():
    """签名成功时应 yield 日志行"""
    config = _make_config()
    log_lines = [
        "Done Adding Additional Store",
        "Successfully signed: test.exe",
    ]

    with mock.patch("subprocess.Popen") as mock_popen:
        mock_process = mock.MagicMock()
        mock_process.stdout.readline.side_effect = log_lines + [""]
        mock_process.wait.return_value = 0
        mock_popen.return_value = mock_process

        result = list(sign_executable(
            "in.exe", "c.pfx", "pwd", config
        ))

        assert result == log_lines


def test_sign_executable_failure():
    """签名失败应抛出 SigningError"""
    config = _make_config()
    log_lines = ["Error: Access denied", "Signing failed"]

    with mock.patch("subprocess.Popen") as mock_popen:
        mock_process = mock.MagicMock()
        mock_process.stdout.readline.side_effect = log_lines + [""]
        mock_process.wait.return_value = 1
        mock_popen.return_value = mock_process

        with pytest.raises(SigningError) as exc_info:
            list(sign_executable("in.exe", "c.pfx", "pwd", config))

        assert exc_info.value.exit_code == 1
        assert "Error: Access denied" in exc_info.value.stderr


def test_sign_executable_subprocess_error():
    """子进程启动失败应传播异常"""
    config = _make_config()

    with mock.patch("subprocess.Popen", side_effect=FileNotFoundError("signtool not found")):
        with pytest.raises(FileNotFoundError):
            list(sign_executable("in.exe", "c.pfx", "pwd", config))
