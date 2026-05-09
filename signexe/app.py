import os
import shutil
from pathlib import Path

import gradio as gr

from config import load_config
from cert_manager import scan_certificates
from signer import sign_executable, SigningError
from utils import save_uploaded_file
from zip_handler import (
    extract_zip,
    find_signable_files,
    repack_zip,
    cleanup_temp_dir,
    ZipProcessingError,
)

CONFIG = load_config("config.yaml")
CERT_DIR = "certificates"
UPLOAD_DIR = "uploads"


def _build_auth() -> list[tuple[str, str]] | None:
    users = [(u.username, u.password) for u in CONFIG.auth_users]
    return users if users else None


def on_file_upload(file_path: str | None) -> tuple[str, str | None]:
    if file_path is None:
        return "No file selected", None
    saved = save_uploaded_file(file_path, UPLOAD_DIR)
    name = Path(file_path).name
    size = os.path.getsize(saved)
    is_zip = name.lower().endswith(".zip")
    label = "ZIP archive" if is_zip else "File"
    return f"**{label}**: `{name}` — {size:,} bytes", saved


def on_refresh_certs() -> tuple[gr.Dropdown, str]:
    certs = scan_certificates(CERT_DIR, CONFIG.certificates.passwords)
    if not certs:
        return (
            gr.Dropdown(choices=[], value=None, interactive=False),
            "*No certificates found in certificates/ directory.*",
        )
    choices = []
    for c in certs:
        label = f"{c.filename}  |  {c.subject}  |  Expires: {c.expiry}"
        choices.append(label)
    return (
        gr.Dropdown(choices=choices, value=choices[0], interactive=True),
        f"{len(certs)} certificate(s) loaded.",
    )


def _sign_zip(
    zip_path: str, cert_path: str, password: str
) -> tuple[str, str | None]:
    """Extract ZIP, sign all EXE/DLL, repack. Returns (all_logs, signed_zip_or_none)."""
    in_path = Path(zip_path)
    temp_dir = in_path.parent / f"_zip_{in_path.stem}"
    signed_zip = in_path.parent / f"{in_path.stem}.signed.zip"

    all_logs = f"Processing ZIP: {in_path.name}\n"

    try:
        all_logs += "Extracting...\n"
        extract_zip(str(zip_path), str(temp_dir))

        files = find_signable_files(str(temp_dir))
        if not files:
            all_logs += "No .exe or .dll files found in ZIP.\nRepacking original files...\n"
            repack_zip(str(temp_dir), str(signed_zip))
            all_logs += f"✓ Repacked to: {signed_zip.name}\n"
            return all_logs, str(signed_zip)

        all_logs += f"Found {len(files)} signable file(s):\n"
        for f in files:
            rel = os.path.relpath(f, str(temp_dir))
            all_logs += f"  - {rel}\n"
        all_logs += f"\nCertificate: {Path(cert_path).name}\n\n"

        for f in files:
            rel = os.path.relpath(f, str(temp_dir))
            all_logs += f"--- Signing {rel} ---\n"
            for line in sign_executable(f, cert_path, password, CONFIG):
                all_logs += line + "\n"
            all_logs += "\n"

        all_logs += "Repacking signed files...\n"
        repack_zip(str(temp_dir), str(signed_zip))
        all_logs += f"✓ ZIP signed successfully → {signed_zip.name}\n"
        return all_logs, str(signed_zip)

    except ZipProcessingError as e:
        all_logs += f"\n✗ ZIP error: {e}\n"
        return all_logs, None
    except SigningError as e:
        all_logs += f"\n✗ Signing failed inside ZIP (exit code {e.exit_code})\n"
        return all_logs, None
    except FileNotFoundError:
        all_logs += f"\n✗ signtool.exe not found at: {CONFIG.signtool_exe}\n"
        return all_logs, None
    finally:
        cleanup_temp_dir(str(temp_dir))


def on_sign(
    file_path: str | None,
    cert_display: str | None,
) -> tuple[str, gr.DownloadButton]:
    if file_path is None:
        yield "⚠ No file uploaded. Please upload a file first.", gr.DownloadButton(interactive=False)
        return
    if not cert_display:
        yield "⚠ No certificate selected. Please click Refresh to scan certificates.", gr.DownloadButton(interactive=False)
        return

    filename = cert_display.split("  |  ")[0].strip()
    try:
        password = CONFIG.certificates.passwords[filename]
    except KeyError:
        yield f"⚠ No password configured for '{filename}'.", gr.DownloadButton(interactive=False)
        return
    cert_path = os.path.join(CERT_DIR, filename)

    is_zip = file_path.lower().endswith(".zip")

    if is_zip:
        all_logs, result = _sign_zip(file_path, cert_path, password)
        yield all_logs, (
            gr.DownloadButton(value=result, interactive=True)
            if result
            else gr.DownloadButton(interactive=False)
        )
    else:
        all_logs = ""
        in_path = Path(file_path)
        signed_path = in_path.parent / f"{in_path.stem}.signed{in_path.suffix}"
        shutil.copy2(file_path, str(signed_path))

        all_logs = f"Signing: {in_path.name}\nCertificate: {filename}\n\n"
        try:
            for line in sign_executable(str(signed_path), cert_path, password, CONFIG):
                all_logs += line + "\n"
                yield all_logs, gr.DownloadButton(interactive=False)
            all_logs += "\n✓ Signing completed successfully.\n"
            yield all_logs, gr.DownloadButton(value=str(signed_path), interactive=True)
        except SigningError as e:
            all_logs += f"\n✗ Signing failed (exit code {e.exit_code})\n"
            yield all_logs, gr.DownloadButton(interactive=False)
        except FileNotFoundError:
            all_logs += f"\n✗ signtool.exe not found at: {CONFIG.signtool_exe}\n"
            yield all_logs, gr.DownloadButton(interactive=False)


def create_demo() -> gr.Blocks:
    with gr.Blocks(title="EXE / DLL / ZIP Signing Tool") as demo:
        gr.Markdown("# EXE / DLL / ZIP Signing Tool")
        gr.Markdown(
            "Upload an EXE, DLL, or ZIP file, select a certificate, and sign with signtool. "
            "ZIP files are extracted, signed internally, and repacked."
        )

        file_state = gr.State(None)

        upload = gr.File(
            label="Upload File",
            file_types=[".exe", ".dll", ".msi", ".zip", ".sys", ".cat", ".cab", ".ocx", ".drv"],
            type="filepath",
        )
        file_info = gr.Markdown("No file selected")

        gr.Markdown("---")

        with gr.Row():
            cert_dropdown = gr.Dropdown(
                label="Signing Certificate",
                choices=[],
                interactive=True,
                scale=3,
            )
            refresh_btn = gr.Button("Refresh Certificates", scale=1)

        cert_status = gr.Markdown("")

        gr.Markdown("---")

        sign_btn = gr.Button("Sign File", variant="primary", size="lg")

        log_output = gr.Textbox(
            label="Signing Log",
            lines=14,
            max_lines=40,
            interactive=False,
            autoscroll=True,
        )

        download_btn = gr.DownloadButton(
            label="Download Signed File",
            variant="secondary",
            interactive=False,
            visible=True,
        )

        upload.upload(on_file_upload, upload, [file_info, file_state])
        refresh_btn.click(on_refresh_certs, None, [cert_dropdown, cert_status])
        sign_btn.click(on_sign, [file_state, cert_dropdown], [log_output, download_btn])

    return demo


demo = create_demo()

if __name__ == "__main__":
    demo.launch(
        auth=_build_auth(),
        server_name=CONFIG.server.host,
        server_port=CONFIG.server.port,
        max_file_size=f"{CONFIG.signing.max_file_size_mb}mb",
    )
