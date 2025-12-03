#include "selfprotect.h"
#include <windows.h>
#include <aclapi.h>
#include <sddl.h>

int IsUserAdmin() {
    BOOL isAdmin = FALSE;
    PSID adminGroup = NULL;
    SID_IDENTIFIER_AUTHORITY ntAuthority = SECURITY_NT_AUTHORITY;
    
    // Create SID for administrators group
    if (AllocateAndInitializeSid(
        &ntAuthority,
        2,
        SECURITY_BUILTIN_DOMAIN_RID,
        DOMAIN_ALIAS_RID_ADMINS,
        0, 0, 0, 0, 0, 0,
        &adminGroup
    )) {
        // Check if current user is in administrators group
        CheckTokenMembership(NULL, adminGroup, &isAdmin);
        FreeSid(adminGroup);
    }
    
    return isAdmin ? 1 : 0;
}

int ProtectProcess() {
    HANDLE hProcess = GetCurrentProcess();
    PACL pOldDACL = NULL, pNewDACL = NULL;
    PSECURITY_DESCRIPTOR pSD = NULL;
    EXPLICIT_ACCESS ea;
    PSID pEveryoneSID = NULL;
    SID_IDENTIFIER_AUTHORITY SIDAuthWorld = SECURITY_WORLD_SID_AUTHORITY;
    DWORD dwRes;
    
    // Get current process DACL
    dwRes = GetSecurityInfo(
        hProcess,
        SE_KERNEL_OBJECT,
        DACL_SECURITY_INFORMATION,
        NULL,
        NULL,
        &pOldDACL,
        NULL,
        &pSD
    );
    
    if (dwRes != ERROR_SUCCESS) {
        return -1;
    }
    
    // Create SID for "Everyone"
    if (!AllocateAndInitializeSid(
        &SIDAuthWorld,
        1,
        SECURITY_WORLD_RID,
        0, 0, 0, 0, 0, 0, 0,
        &pEveryoneSID
    )) {
        if (pSD) LocalFree(pSD);
        return -1;
    }
    
    // Setup EXPLICIT_ACCESS to deny PROCESS_TERMINATE for Everyone
    ZeroMemory(&ea, sizeof(EXPLICIT_ACCESS));
    ea.grfAccessPermissions = PROCESS_TERMINATE;
    ea.grfAccessMode = DENY_ACCESS;
    ea.grfInheritance = NO_INHERITANCE;
    ea.Trustee.TrusteeForm = TRUSTEE_IS_SID;
    ea.Trustee.TrusteeType = TRUSTEE_IS_WELL_KNOWN_GROUP;
    ea.Trustee.ptstrName = (LPTSTR)pEveryoneSID;
    
    // Create new DACL
    dwRes = SetEntriesInAcl(1, &ea, pOldDACL, &pNewDACL);
    
    if (dwRes != ERROR_SUCCESS) {
        FreeSid(pEveryoneSID);
        if (pSD) LocalFree(pSD);
        return -1;
    }
    
    // Apply new DACL to process
    dwRes = SetSecurityInfo(
        hProcess,
        SE_KERNEL_OBJECT,
        DACL_SECURITY_INFORMATION,
        NULL,
        NULL,
        pNewDACL,
        NULL
    );
    
    // Cleanup
    if (pNewDACL) LocalFree(pNewDACL);
    if (pEveryoneSID) FreeSid(pEveryoneSID);
    if (pSD) LocalFree(pSD);
    
    return (dwRes == ERROR_SUCCESS) ? 0 : -1;
}
