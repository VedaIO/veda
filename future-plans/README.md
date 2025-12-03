# ProcGuard Future Plans

This directory contains detailed implementation plans for advanced features that will benefit from CGO + Zig and native Windows API integration.

## Priority Features (Critical)

### 1. [Self-Protection System](./01-self-protection.md)
Prevent unauthorized termination of ProcGuard. Only stoppable via the "Dừng ProcGuard" button.
- **Key APIs:** `RtlSetProcessIsCritical`, `NtSetInformationProcess` (PPL)
- **Complexity:** High (requires EV code signing for PPL)
- **Timeline:** 15-20 days

### 2. [Anti-Cheat Defense](./02-anti-cheat-defense.md)
Detect executable renaming, hash-based tracking, symbolic link resolution, and tamper reporting.
- **Key APIs:** `WinVerifyTrust`, `GetFinalPathNameByHandleW`, PE parsing
- **Complexity:** Medium
- **Timeline:** 10-15 days

### 3. [Uninstall Protection](./03-uninstall-protection.md)
Prevent unauthorized uninstallation. Only removable via password-protected built-in menu.
- **Key APIs:** `SetNamedSecurityInfo` (ACLs), `RegSetKeySecurity`
- **Complexity:** Medium-High
- **Timeline:** 10-12 days

### 4. [Accurate Screen Time](./04-accurate-screen-time.md)
Track active window time, idle detection, multi-monitor support, and browser tab tracking.
- **Key APIs:** `SetWinEventHook`, `GetForegroundWindow`, UI Automation
- **Complexity:** High (event-driven architecture)
- **Timeline:** 15-20 days

### 5. [Multi-User Session Handling](./05-multi-user-sessions.md)
Support multiple Windows users with independent time limits and Fast User Switching.
- **Key APIs:** `WTSEnumerateSessions`, `CreateProcessAsUserW`
- **Complexity:** High (complex IPC and session management)
- **Timeline:** 15-18 days

## Secondary Features (Nice-to-Have)

### 6. [Network Filtering](./06-network-filtering.md)
Per-application network access control using Windows Filtering Platform (WFP).
- **Key APIs:** `FwpmEngineOpen0`, `FwpmFilterAdd0`, callout driver (optional)
- **Complexity:** Very High (kernel driver for data usage tracking)
- **Timeline:** 15-30 days (depending on scope)

### 7. [Credential Management](./07-credential-management.md)
Store passwords in Windows Credential Manager, Windows Hello support, DPAPI encryption.
- **Key APIs:** `CredWrite`, `CredRead`, `CryptProtectData`, WinRT Hello
- **Complexity:** Medium (COM + WinRT required)
- **Timeline:** 7-10 days

### 8. [Windows Toast Notifications](./08-toast-notifications.md)
Native Windows 10/11 notifications with action buttons.
- **Key APIs:** `Windows.UI.Notifications.ToastNotificationManager` (WinRT)
- **Complexity:** Medium-High (COM + WinRT + background activation)
- **Timeline:** 10-12 days

### 9. [Advanced Task Scheduling](./09-task-scheduling.md)
Time-based rules ("Allow Minecraft 7-8 PM weekends") via Task Scheduler COM API.
- **Key APIs:** `ITaskService`, `ITaskDefinition`, `ITrigger`
- **Complexity:** Medium (COM interface)
- **Timeline:** 10-12 days

### 10. [Advanced Logging & Crash Reporting](./10-crash-reporting.md)
Minidumps, ETW circular buffers, Windows Error Reporting integration, forensic analysis.
- **Key APIs:** `MiniDumpWriteDump`, `EventWrite` (ETW), WER
- **Complexity:** High (kernel-mode ETW provider)
- **Timeline:** 12-15 days

## Why CGO + Zig?

All features above require **deep Windows API integration** that is either:
1. **Impossible in pure Go** (e.g., COM interfaces, WinRT, ETW callouts)
2. **Very painful in pure Go** (e.g., complex structs, callback functions)
3. **Missing from `x/sys/windows`** (e.g., undocumented APIs like `NtSetInformationProcess`)

### Zig Benefits:
- **Clean FFI:** Easier to work with Windows SDK headers
- **Static Linking:** Avoids DLL dependency hell
- **Cross-Compilation:** Build Windows binaries from Linux/Mac
- **Struct Packing:** Better control over memory layout for WinAPI structures

## Implementation Strategy

### Recommended Order:
1. **Start:** Self-Protection (foundational for all other features)
2. **Next:** Anti-Cheat Defense (addresses immediate bypass attempts)
3. **Then:** Accurate Screen Time (improves core functionality)
4. **After:** Multi-User Sessions (expands market to schools/offices)
5. **Finally:** Secondary features based on user demand

### Hybrid Architecture:
```
Pure Go Core (Current)
├─ Wails UI
├─ SQLite Database
├─ HTTP Server
└─ Basic process monitoring

CGO + Zig Modules (New)
├─ Self-Protection DLL
├─ Anti-Cheat Scanner
├─ Screen Time Monitor (event-driven)
├─ Session Manager
└─ Network Filter Driver (optional)
```

## Testing & Quality

Each feature document includes:
- ✅ **Implementation Phases** (broken into testable chunks)
- ✅ **Test Cases** (specific scenarios to validate)
- ✅ **Success Metrics** (measurable goals)
- ✅ **CGO Benefits** (justification for native code)

## Questions?

- **Why not pure Go?** Many of these APIs are Windows-specific and have no Go equivalent
- **Why Zig over C++?** Cleaner FFI, better cross-compilation, modern tooling
- **Security concerns?** All features designed with security in mind, include tamper detection

## Contributing

When implementing a feature:
1. Read the corresponding plan document thoroughly
2. Start with Phase 1 (usually simplest)
3. Test incrementally after each phase
4. Update this README with actual implementation time
5. Document any deviations from the plan

---

**Total Development Effort:** 100-150 days (solo developer)  
**Estimated with Team of 3:** 40-60 days  
**Realistic Timeline with Priorities 1-5 only:** 70-85 days
