//go:build windows

// Package screentime provides foreground window tracking functionality.
package screentime

/*
#include <windows.h>
#include <stdint.h>

typedef struct {
    uint32_t pid;
    wchar_t title[256];
} ActiveWindowInfo;

int GetActiveWindowInfo(ActiveWindowInfo* info) {
    if (info == NULL) {
        return -1;
    }

    info->pid = 0;
    info->title[0] = L'\0';

    HWND hwnd = GetForegroundWindow();
    if (hwnd == NULL) {
        return -1;
    }

    DWORD pid = 0;
    GetWindowThreadProcessId(hwnd, &pid);
    if (pid == 0) {
        return -1;
    }
    info->pid = (uint32_t)pid;

    GetWindowTextW(hwnd, info->title, 256);

    return 0;
}
*/
import "C"

// WindowInfo contains information about the currently focused window
type WindowInfo struct {
	PID   uint32
	Title string
}

// GetActiveWindowInfo retrieves the PID and title of the foreground window.
// Returns nil if no foreground window is found.
func GetActiveWindowInfo() *WindowInfo {
	var cInfo C.ActiveWindowInfo
	result := C.GetActiveWindowInfo(&cInfo)
	if result != 0 {
		return nil
	}

	title := wcharToString(cInfo.title[:])

	return &WindowInfo{
		PID:   uint32(cInfo.pid),
		Title: title,
	}
}

func wcharToString(wchars []C.wchar_t) string {
	length := 0
	for i, c := range wchars {
		if c == 0 {
			length = i
			break
		}
	}
	if length == 0 {
		return ""
	}

	utf16 := make([]uint16, length)
	for i := 0; i < length; i++ {
		utf16[i] = uint16(wchars[i])
	}

	runes := make([]rune, len(utf16))
	for i, v := range utf16 {
		runes[i] = rune(v)
	}
	return string(runes)
}
