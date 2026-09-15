//go:build windows

package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func writeRestrictedAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create replay artifact directory: %w", err)
	}

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("identify replay artifact owner: %w", err)
	}
	sddl := "D:P(A;;FA;;;" + user.User.Sid.String() + ")"
	securityDescriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return fmt.Errorf("create replay artifact ACL: %w", err)
	}
	securityAttributes := windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: securityDescriptor,
	}

	var tempPath string
	var temp *os.File
	for attempt := 0; attempt < 10; attempt++ {
		random := make([]byte, 12)
		if _, err := rand.Read(random); err != nil {
			return fmt.Errorf("generate replay artifact temporary name: %w", err)
		}
		tempPath = filepath.Join(dir, "."+filepath.Base(path)+".tmp-"+hex.EncodeToString(random))
		name, err := windows.UTF16PtrFromString(tempPath)
		if err != nil {
			return fmt.Errorf("encode replay artifact path: %w", err)
		}
		handle, err := windows.CreateFile(
			name,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0,
			&securityAttributes,
			windows.CREATE_NEW,
			windows.FILE_ATTRIBUTE_NORMAL,
			0,
		)
		if errors.Is(err, windows.ERROR_FILE_EXISTS) || errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			continue
		}
		if err != nil {
			return fmt.Errorf("create replay artifact: %w", err)
		}
		temp = os.NewFile(uintptr(handle), tempPath)
		break
	}
	if temp == nil {
		return fmt.Errorf("create replay artifact: temporary name collisions exhausted")
	}
	defer func() { _ = os.Remove(tempPath) }()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write replay artifact: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync replay artifact: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close replay artifact: %w", err)
	}
	from, err := windows.UTF16PtrFromString(tempPath)
	if err != nil {
		return fmt.Errorf("encode replay artifact temporary path: %w", err)
	}
	to, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode replay artifact destination: %w", err)
	}
	if err := windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return fmt.Errorf("publish replay artifact: %w", err)
	}
	return nil
}
