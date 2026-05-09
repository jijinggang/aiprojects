import subprocess
from unittest import mock

import pytest

from cert_manager import CertInfo, scan_certificates, get_cert_password, _parse_certutil_output


CERTUTIL_OUTPUT = """
================ Certificate 0 ================
Serial Number: 0123456789abcdef
Issuer: CN=DigiCert Trusted G4 Code Signing RSA4096 SHA384 2021 CA1
  O="DigiCert, Inc.", C=US
NotBefore: 2024-01-01 00:00
NotAfter: 2025-12-31 23:59
Subject: CN=My Company Inc., O=My Company Inc., L=San Jose
  S=California, C=US
Non-root Certificate
  Key: (2048 Bits)
"""

CERTUTIL_OUTPUT_MINIMAL = """
================ Certificate 0 ================
Issuer: CN=Test CA
NotAfter: 2026-06-15 12:00
Subject: CN=Test Cert
"""


def test_parse_certutil_output():
    """解析 certutil 输出提取证书元数据"""
    info = _parse_certutil_output(CERTUTIL_OUTPUT)
    assert info["subject"].startswith("CN=My Company Inc.")
    assert "O=My Company Inc." in info["subject"]
    assert info["issuer"].startswith("CN=DigiCert Trusted")
    assert info["expiry"] == "2025-12-31 23:59"


def test_parse_certutil_output_minimal():
    """解析最小 certutil 输出"""
    info = _parse_certutil_output(CERTUTIL_OUTPUT_MINIMAL)
    assert info["subject"] == "CN=Test Cert"
    assert info["issuer"] == "CN=Test CA"
    assert info["expiry"] == "2026-06-15 12:00"


def test_parse_certutil_output_missing_fields():
    """缺少字段时抛出错误"""
    with pytest.raises(ValueError):
        _parse_certutil_output("No certificate here")


def test_scan_certificates(tmp_path):
    """扫描证书目录找到 PFX 文件"""
    cert_dir = tmp_path / "certs"
    cert_dir.mkdir()
    (cert_dir / "cert1.pfx").write_text("pfx data 1")
    (cert_dir / "cert2.pfx").write_text("pfx data 2")
    (cert_dir / "readme.txt").write_text("not a cert")

    passwords = {"cert1.pfx": "pass1", "cert2.pfx": "pass2"}

    certutil_out = """
Issuer: CN=Test CA
NotAfter: 2026-01-01 00:00
Subject: CN=Test Cert
"""

    with mock.patch("subprocess.run") as mock_run:
        mock_result = mock.MagicMock()
        mock_result.stdout = certutil_out
        mock_result.returncode = 0
        mock_run.return_value = mock_result

        certs = scan_certificates(str(cert_dir), passwords)

        assert len(certs) == 2
        assert all(isinstance(c, CertInfo) for c in certs)
        assert certs[0].filename == "cert1.pfx"
        assert certs[0].subject == "CN=Test Cert"


def test_scan_certificates_empty_dir(tmp_path):
    """空目录返回空列表"""
    cert_dir = tmp_path / "empty_certs"
    cert_dir.mkdir()

    certs = scan_certificates(str(cert_dir), {})
    assert certs == []


def test_scan_certificates_no_pfx(tmp_path):
    """无 PFX 文件返回空列表"""
    cert_dir = tmp_path / "no_pfx"
    cert_dir.mkdir()
    (cert_dir / "readme.txt").write_text("hello")

    certs = scan_certificates(str(cert_dir), {})
    assert certs == []


def test_scan_certificates_damaged_skipped(tmp_path):
    """损坏的证书被跳过"""
    cert_dir = tmp_path / "certs"
    cert_dir.mkdir()
    (cert_dir / "good.pfx").write_text("good")
    (cert_dir / "bad.pfx").write_text("bad")

    passwords = {"good.pfx": "pass1", "bad.pfx": "pass2"}

    certutil_out = """
Issuer: CN=Good CA
NotAfter: 2026-01-01
Subject: CN=Good Cert
"""

    def mock_run_side_effect(*args, **kwargs):
        cmd = args[0] if args else kwargs.get("args", [])
        cmd_str = " ".join(cmd)
        if "bad.pfx" in cmd_str:
            raise subprocess.CalledProcessError(1, cmd_str, output="", stderr="error")
        result = mock.MagicMock()
        result.stdout = certutil_out
        result.returncode = 0
        return result

    with mock.patch("subprocess.run", side_effect=mock_run_side_effect):
        certs = scan_certificates(str(cert_dir), passwords)
        assert len(certs) == 1
        assert certs[0].filename == "good.pfx"


def test_get_cert_password():
    """在密码映射中查找证书密码"""
    passwords = {"my_cert.pfx": "secret123"}
    assert get_cert_password("my_cert.pfx", passwords) == "secret123"


def test_get_cert_password_missing():
    """密码未配置时抛出错误"""
    with pytest.raises(KeyError, match="my_cert.pfx"):
        get_cert_password("my_cert.pfx", {})


def test_get_cert_password_multiple():
    """多证书密码匹配"""
    passwords = {"a.pfx": "pa", "b.pfx": "pb"}
    assert get_cert_password("b.pfx", passwords) == "pb"
