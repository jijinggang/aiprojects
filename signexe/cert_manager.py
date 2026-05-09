import logging
import os
import re
import subprocess
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class CertInfo:
    filename: str
    path: str
    subject: str
    issuer: str
    expiry: str


def get_cert_password(filename: str, passwords: dict[str, str]) -> str:
    if filename not in passwords:
        raise KeyError(
            f"Certificate '{filename}' has no password configured. "
            f"Add it to certificates.passwords in config.yaml"
        )
    return passwords[filename]


def _parse_certutil_output(output: str) -> dict[str, str]:
    subject = None
    issuer = None
    expiry = None

    subject_match = re.search(r"^Subject:\s*(.+)$", output, re.MULTILINE)
    if subject_match:
        subject = subject_match.group(1)
        next_line_match = re.search(
            r"^Subject:\s*.+\n\s+(.+)$", output, re.MULTILINE
        )
        if next_line_match:
            subject += " " + next_line_match.group(1)

    issuer_match = re.search(r"^Issuer:\s*(.+)$", output, re.MULTILINE)
    if issuer_match:
        issuer = issuer_match.group(1)
        next_line_match = re.search(
            r"^Issuer:\s*.+\n\s+(.+)$", output, re.MULTILINE
        )
        if next_line_match:
            issuer += " " + next_line_match.group(1)

    expiry_match = re.search(r"^NotAfter:\s*(.+)$", output, re.MULTILINE)
    if expiry_match:
        expiry = expiry_match.group(1)

    if not subject or not issuer or not expiry:
        raise ValueError("Failed to parse certificate metadata from certutil output")

    return {"subject": subject, "issuer": issuer, "expiry": expiry}


def scan_certificates(cert_dir: str, passwords: dict[str, str]) -> list[CertInfo]:
    if not os.path.isdir(cert_dir):
        return []

    certs = []
    for entry in sorted(os.scandir(cert_dir), key=lambda e: e.name):
        if not entry.is_file() or not entry.name.lower().endswith(".pfx"):
            continue

        try:
            password = get_cert_password(entry.name, passwords)
        except KeyError:
            logger.warning(
                "Skipping %s: no password configured", entry.name
            )
            continue

        try:
            result = subprocess.run(
                ["certutil", "-dump", "-p", password, entry.path],
                capture_output=True,
                text=True,
                timeout=30,
            )
            result.check_returncode()
            info = _parse_certutil_output(result.stdout)
            certs.append(
                CertInfo(
                    filename=entry.name,
                    path=entry.path,
                    subject=info["subject"],
                    issuer=info["issuer"],
                    expiry=info["expiry"],
                )
            )
        except (subprocess.CalledProcessError, ValueError) as e:
            logger.warning("Skipping %s: %s", entry.name, e)

    return certs
