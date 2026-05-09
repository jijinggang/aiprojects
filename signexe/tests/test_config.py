import os
import tempfile
from unittest import mock

import pytest
import yaml

from config import AppConfig, load_config, _find_signtool


def test_appconfig_defaults():
    """AppConfig 应该从默认值创建"""
    config = AppConfig()
    assert config.auth_users == []
    assert config.certificates.passwords == {}
    assert config.signing.timestamper_url == "http://timestamp.digicert.com"
    assert config.signing.additional_flags == ""
    assert config.signing.max_file_size_mb == 500
    assert config.server.host == "127.0.0.1"
    assert config.server.port == 7860
    assert config.uploads.retention_hours == 24
    assert config.signtool_exe == ""


def test_load_config_basic():
    """应该从 YAML 文件加载配置"""
    config_data = {
        "auth": {
            "users": [
                {"username": "admin", "password": "secret123"},
                {"username": "operator", "password": "op456"},
            ]
        },
        "certificates": {
            "passwords": {
                "my_cert.pfx": "certpass1",
                "another.pfx": "certpass2",
            }
        },
        "signing": {
            "timestamper_url": "http://timestamp.digicert.com",
            "additional_flags": "/as",
            "max_file_size_mb": 200,
        },
        "server": {"host": "0.0.0.0", "port": 8080},
        "uploads": {"retention_hours": 48},
    }
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
        yaml.dump(config_data, f)
        tmp_path = f.name

    try:
        config = load_config(tmp_path)
        assert len(config.auth_users) == 2
        assert config.auth_users[0].username == "admin"
        assert config.auth_users[0].password == "secret123"
        assert config.auth_users[1].username == "operator"
        assert config.certificates.passwords == {
            "my_cert.pfx": "certpass1",
            "another.pfx": "certpass2",
        }
        assert config.signing.timestamper_url == "http://timestamp.digicert.com"
        assert config.signing.additional_flags == "/as"
        assert config.signing.max_file_size_mb == 200
        assert config.server.host == "0.0.0.0"
        assert config.server.port == 8080
        assert config.uploads.retention_hours == 48
    finally:
        os.unlink(tmp_path)


def test_load_config_file_not_found():
    """不存在的配置文件应抛出 FileNotFoundError"""
    with pytest.raises(FileNotFoundError):
        load_config("nonexistent_config.yaml")


def test_load_config_explicit_signtool():
    """显式指定的 signtool 路径应被保留"""
    config_data = {
        "auth": {"users": []},
        "certificates": {"passwords": {}},
        "signing": {"timestamper_url": "", "additional_flags": "", "max_file_size_mb": 500},
        "server": {"host": "127.0.0.1", "port": 7860},
        "uploads": {"retention_hours": 24},
        "signtool_exe": "C:\\tools\\signtool.exe",
    }
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
        yaml.dump(config_data, f)
        tmp_path = f.name

    try:
        config = load_config(tmp_path)
        assert config.signtool_exe == "C:\\tools\\signtool.exe"
    finally:
        os.unlink(tmp_path)


def test_find_signtool_in_sdk_path():
    """signtool 应在 SDK 路径中被找到，选最高版本"""
    mock_glob_result = [
        "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.19041.0\\x64\\signtool.exe",
        "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\signtool.exe",
        "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22000.0\\x64\\signtool.exe",
    ]

    with mock.patch("glob.glob", return_value=mock_glob_result):
        with mock.patch("os.path.exists", return_value=True):
            result = _find_signtool()
            assert "10.0.22621.0" in result


def test_find_signtool_in_path():
    """SDK 无结果时应在 PATH 中搜索"""
    with mock.patch("glob.glob", return_value=[]):
        with mock.patch("os.path.exists", return_value=False):
            with mock.patch("shutil.which", return_value="C:\\tools\\signtool.exe"):
                result = _find_signtool()
                assert result == "C:\\tools\\signtool.exe"


def test_find_signtool_not_found():
    """找不到 signtool 时应抛出 RuntimeError"""
    with mock.patch("glob.glob", return_value=[]):
        with mock.patch("os.path.exists", return_value=False):
            with mock.patch("shutil.which", return_value=None):
                with pytest.raises(RuntimeError, match="signtool"):
                    _find_signtool()


def test_load_config_auto_find_signtool():
    """未显式配置 signtool 路径时应自动探测"""
    config_data = {
        "auth": {"users": []},
        "certificates": {"passwords": {}},
        "signing": {"timestamper_url": "", "additional_flags": "", "max_file_size_mb": 500},
        "server": {"host": "127.0.0.1", "port": 7860},
        "uploads": {"retention_hours": 24},
    }
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
        yaml.dump(config_data, f)
        tmp_path = f.name

    try:
        mock_path = "C:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\signtool.exe"
        with mock.patch("glob.glob", return_value=[mock_path]):
            with mock.patch("os.path.exists", return_value=True):
                config = load_config(tmp_path)
                assert config.signtool_exe == mock_path
    finally:
        os.unlink(tmp_path)
