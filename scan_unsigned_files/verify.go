package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	wintrust           = windows.NewLazySystemDLL("wintrust.dll")
	procWinVerifyTrust = wintrust.NewProc("WinVerifyTrust")
)

// GUID for WINTRUST_ACTION_GENERIC_VERIFY_V2
// {00AAC56B-CD44-11d0-8CC2-00C04FC295EE}
var wintrustActionGenericVerifyV2 = windows.GUID{
	Data1: 0x00AAC56B,
	Data2: 0xCD44,
	Data3: 0x11d0,
	Data4: [8]byte{0x8C, 0xC2, 0x00, 0xC0, 0x4F, 0xC2, 0x95, 0xEE},
}

// WinTrust constants
const (
	WTD_UI_NONE          = 2
	WTD_CHOICE_FILE      = 1
	WTD_STATEACTION_CLOSE = 2
	WTD_REVOKE_NONE      = 0
	WTD_SAFER_FLAG       = 0x100
)

// Windows trust error codes
const (
	TRUST_E_NOSIGNATURE          = 0x800B0100
	TRUST_E_SUBJECT_FORM_UNKNOWN = 0x800B0003
	TRUST_E_PROVIDER_UNKNOWN     = 0x800B0001
	CERT_E_EXPIRED               = 0x800B0101
	CERT_E_REVOKED               = 0x800B010C
	CERT_E_UNTRUSTEDROOT         = 0x800B0109
	CERT_E_CHAINING              = 0x800B010A
	TRUST_E_FAIL                 = 0x800B010B
	CERT_E_WRONG_USAGE           = 0x800B0110
	CRYPT_E_SECURITY_SETTINGS    = 0x80092026
)

// WinTrust file info structure
type wintrustFileInfo struct {
	Size     uint32
	FilePath *uint16
	File     windows.Handle
	KnownSubject *windows.GUID
}

// WinTrust data structure (matches WINTRUST_DATA from wintrust.h)
type wintrustData struct {
	Size              uint32
	PolicyCallback    uintptr
	SIPClientData     uintptr
	UIChoice          uint32
	RevocationChecks  uint32
	UnionChoice       uint32
	FileInfo          *wintrustFileInfo
	StateAction       uint32
	StateData         windows.Handle
	ProvFlags         uint32
	UIContext         uint32
	SignatureSettings *wintrustSignatureSettings
}

type wintrustSignatureSettings struct {
	Size             uint32
	Index            uint32
	Flags            uint32
	SecondarySigs    uint32
	VerifiedSigIndex uint32
	CryptoPolicy     uintptr
}

// IsSigned checks if a PE file has a valid digital signature.
// Returns (true, "", nil) if the file is validly signed.
// Returns (false, reason, nil) if not signed or signature is invalid/expired/revoked.
func IsSigned(filePath string) (bool, string, error) {
	pathUTF16, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return false, "", fmt.Errorf("UTF16PtrFromString: %w", err)
	}

	fileInfo := &wintrustFileInfo{
		Size:     uint32(unsafe.Sizeof(wintrustFileInfo{})),
		FilePath: pathUTF16,
	}

	data := &wintrustData{
		Size:     uint32(unsafe.Sizeof(wintrustData{})),
		UIChoice: WTD_UI_NONE,
		UnionChoice: WTD_CHOICE_FILE,
		FileInfo:  fileInfo,
		RevocationChecks: WTD_REVOKE_NONE,
	}

	// First call: verify
	data.StateAction = 1 // WTD_STATEACTION_VERIFY
	ret, _, _ := procWinVerifyTrust.Call(
		0, // HWND = NULL (no UI)
		uintptr(unsafe.Pointer(&wintrustActionGenericVerifyV2)),
		uintptr(unsafe.Pointer(data)),
	)

	// Close the state
	data.StateAction = WTD_STATEACTION_CLOSE
	procWinVerifyTrust.Call(
		0, // HWND = NULL (no UI)
		uintptr(unsafe.Pointer(&wintrustActionGenericVerifyV2)),
		uintptr(unsafe.Pointer(data)),
	)

	if ret == 0 {
		return true, "", nil
	}

	reason := winTrustErrorReason(uint32(ret))
	return false, reason, nil
}

func winTrustErrorReason(code uint32) string {
	switch code {
	case TRUST_E_NOSIGNATURE:
		return "无数字签名"
	case TRUST_E_SUBJECT_FORM_UNKNOWN, TRUST_E_PROVIDER_UNKNOWN:
		return "文件格式不支持或无法验证"
	case CERT_E_EXPIRED:
		return "证书已过期"
	case CERT_E_REVOKED:
		return "证书已吊销"
	case CERT_E_UNTRUSTEDROOT:
		return "证书根不受信任"
	case CERT_E_CHAINING:
		return "证书链验证失败"
	case TRUST_E_FAIL:
		return "签名验证失败"
	case CERT_E_WRONG_USAGE:
		return "证书用途不匹配"
	case CRYPT_E_SECURITY_SETTINGS:
		return "安全设置阻止验证"
	default:
		return fmt.Sprintf("签名验证失败 (0x%08X)", code)
	}
}

func init() {
	// Verify wintrust.dll is available
	if err := wintrust.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 无法加载 wintrust.dll: %v\n", err)
	}
}
