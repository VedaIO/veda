#ifndef SELFPROTECT_H
#define SELFPROTECT_H

#ifdef __cplusplus
extern "C" {
#endif

// Protect the current process from being terminated
// Returns 0 on success, -1 on error
int ProtectProcess();

// Check if current user has admin privileges
// Returns 1 if admin, 0 if not
int IsUserAdmin();

#ifdef __cplusplus
}
#endif

#endif
