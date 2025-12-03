package selfprotect

/*
#cgo windows CFLAGS: -I.
#cgo windows LDFLAGS: -ladvapi32

#include "selfprotect.h"
*/
import "C"
import "fmt"

// ProtectProcess prevents the current process from being terminated by Task Manager
// or other non-admin tools. Returns error if protection fails.
func ProtectProcess() error {
	result := C.ProtectProcess()
	if result != 0 {
		return fmt.Errorf("failed to protect process")
	}
	return nil
}

// IsUserAdmin checks if the current user has administrator privileges
func IsUserAdmin() bool {
	result := C.IsUserAdmin()
	return result == 1
}
