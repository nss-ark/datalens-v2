# Step 1: Create all package directories
$dirs = @(
    'packages/shared/src/ui',
    'packages/shared/src/components/DataTable',
    'packages/shared/src/types',
    'packages/shared/src/services',
    'packages/shared/src/stores',
    'packages/shared/src/hooks',
    'packages/shared/src/lib',
    'packages/control-centre/src/pages/Consent',
    'packages/control-centre/src/pages/Governance',
    'packages/control-centre/src/pages/Breach',
    'packages/control-centre/src/pages/Compliance',
    'packages/control-centre/src/components/Layout',
    'packages/control-centre/src/components/Dashboard',
    'packages/control-centre/src/components/DSR',
    'packages/control-centre/src/components/DataSources',
    'packages/control-centre/src/components/Consent',
    'packages/control-centre/src/components/Governance/Lineage',
    'packages/control-centre/src/components/Breach',
    'packages/control-centre/src/components/Charts',
    'packages/control-centre/src/components/Forms',
    'packages/control-centre/src/services',
    'packages/control-centre/src/types',
    'packages/control-centre/src/hooks',
    'packages/control-centre/src/stores',
    'packages/control-centre/src/assets',
    'packages/admin/src/pages/Tenants',
    'packages/admin/src/pages/Users',
    'packages/admin/src/pages/Compliance',
    'packages/admin/src/components/Layout',
    'packages/admin/src/components',
    'packages/admin/src/services',
    'packages/admin/src/types',
    'packages/portal/src/pages/Grievance',
    'packages/portal/src/components',
    'packages/portal/src/services',
    'packages/portal/src/types',
    'packages/portal/src/stores'
)
foreach ($d in $dirs) {
    New-Item -ItemType Directory -Path $d -Force | Out-Null
}
Write-Host "All directories created successfully"

# Step 2: Move shared files
# shadcn/ui components
Copy-Item src/components/ui/*.tsx packages/shared/src/ui/ -Force
Write-Host "Moved ui components"

# Common components with CSS modules
$commonFiles = @(
    @('src/components/common/ErrorBoundary.tsx', 'packages/shared/src/components/ErrorBoundary.tsx'),
    @('src/components/common/ErrorFallbacks.tsx', 'packages/shared/src/components/ErrorFallbacks.tsx'),
    @('src/components/common/Toast.tsx', 'packages/shared/src/components/Toast.tsx'),
    @('src/components/common/Toast.module.css', 'packages/shared/src/components/Toast.module.css'),
    @('src/components/common/StatusBadge.tsx', 'packages/shared/src/components/StatusBadge.tsx'),
    @('src/components/common/StatusBadge.module.css', 'packages/shared/src/components/StatusBadge.module.css'),
    @('src/components/common/Button.tsx', 'packages/shared/src/components/Button.tsx'),
    @('src/components/common/Button.module.css', 'packages/shared/src/components/Button.module.css'),
    @('src/components/common/Modal.tsx', 'packages/shared/src/components/Modal.tsx'),
    @('src/components/common/Modal.module.css', 'packages/shared/src/components/Modal.module.css')
)
foreach ($pair in $commonFiles) {
    Copy-Item $pair[0] $pair[1] -Force
}
Write-Host "Moved common components"

# DataTable
Copy-Item src/components/DataTable/*.tsx packages/shared/src/components/DataTable/ -Force
Copy-Item src/components/DataTable/*.css packages/shared/src/components/DataTable/ -Force
Write-Host "Moved DataTable"

# Types, services, stores, hooks, lib
Copy-Item src/types/common.ts packages/shared/src/types/common.ts -Force
Copy-Item src/types/auth.ts packages/shared/src/types/auth.ts -Force
Copy-Item src/services/api.ts packages/shared/src/services/api.ts -Force
Copy-Item src/services/auth.ts packages/shared/src/services/auth.ts -Force
Copy-Item src/stores/authStore.ts packages/shared/src/stores/authStore.ts -Force
Copy-Item src/stores/toastStore.ts packages/shared/src/stores/toastStore.ts -Force
Copy-Item src/stores/uiStore.ts packages/shared/src/stores/uiStore.ts -Force
Copy-Item src/hooks/useAuth.ts packages/shared/src/hooks/useAuth.ts -Force
Copy-Item src/hooks/useMediaQuery.ts packages/shared/src/hooks/useMediaQuery.ts -Force
Copy-Item src/lib/utils.ts packages/shared/src/lib/utils.ts -Force
Write-Host "Moved shared types, services, stores, hooks, lib"

# Step 3: Move control-centre files
# Entry files
Copy-Item src/index.css packages/control-centre/src/index.css -Force
Copy-Item src/App.css packages/control-centre/src/App.css -Force
Copy-Item src/main.tsx packages/control-centre/src/main.tsx -Force

# CC pages - top level
$ccPages = @('Login.tsx','Register.tsx','Dashboard.tsx','DataSources.tsx','DataSourceDetail.tsx','DataSourceConfig.tsx','PIIDiscovery.tsx','PIIDiscovery.module.css','DSRList.tsx','DSRDetail.tsx','ConsentWidgets.tsx','WidgetDetail.tsx')
foreach ($p in $ccPages) {
    if (Test-Path "src/pages/$p") { Copy-Item "src/pages/$p" "packages/control-centre/src/pages/$p" -Force }
}
Write-Host "Moved CC top-level pages"

# CC pages - subdirs
Copy-Item src/pages/Consent/*.tsx packages/control-centre/src/pages/Consent/ -Force
Copy-Item src/pages/Governance/*.tsx packages/control-centre/src/pages/Governance/ -Force
Copy-Item src/pages/Breach/*.tsx packages/control-centre/src/pages/Breach/ -Force
Copy-Item src/pages/Compliance/*.tsx packages/control-centre/src/pages/Compliance/ -Force
Write-Host "Moved CC subdir pages"

# CC layouts
Copy-Item src/components/Layout/AppLayout.tsx packages/control-centre/src/components/Layout/ -Force
Copy-Item src/components/Layout/Sidebar.tsx packages/control-centre/src/components/Layout/ -Force
Copy-Item src/components/Layout/Sidebar.module.css packages/control-centre/src/components/Layout/ -Force
Copy-Item src/components/Layout/Header.tsx packages/control-centre/src/components/Layout/ -Force
Copy-Item src/components/Layout/Header.module.css packages/control-centre/src/components/Layout/ -Force
Write-Host "Moved CC layouts"

# CC components
Copy-Item src/components/Dashboard/*.tsx packages/control-centre/src/components/Dashboard/ -Force
Copy-Item src/components/DSR/*.tsx packages/control-centre/src/components/DSR/ -Force
Copy-Item src/components/DataSources/*.tsx packages/control-centre/src/components/DataSources/ -Force
Copy-Item src/components/Consent/*.tsx packages/control-centre/src/components/Consent/ -Force
Copy-Item src/components/Governance/PolicyForm.tsx packages/control-centre/src/components/Governance/ -Force
Copy-Item src/components/Governance/SuggestionCard.tsx packages/control-centre/src/components/Governance/ -Force
Copy-Item src/components/Governance/Lineage/*.tsx packages/control-centre/src/components/Governance/Lineage/ -Force
Copy-Item src/components/Breach/*.tsx packages/control-centre/src/components/Breach/ -Force
# Charts and Forms directories may be empty, copy if they have files
if (Test-Path src/components/Charts/*.tsx) { Copy-Item src/components/Charts/*.tsx packages/control-centre/src/components/Charts/ -Force }
if (Test-Path src/components/Forms/*.tsx) { Copy-Item src/components/Forms/*.tsx packages/control-centre/src/components/Forms/ -Force }
Write-Host "Moved CC components"

# CC ProtectedRoute
Copy-Item src/components/common/ProtectedRoute.tsx packages/control-centre/src/components/ProtectedRoute.tsx -Force

# CC services
$ccServices = @('dsr.ts','consent.ts','datasource.ts','datasources.ts','discovery.ts','governance.ts','breach.ts','analytics.ts','darkPatternService.ts','dashboard.ts','identity.ts','lineage.ts','grievanceService.ts','notificationService.ts','translationService.ts')
foreach ($s in $ccServices) {
    if (Test-Path "src/services/$s") { Copy-Item "src/services/$s" "packages/control-centre/src/services/$s" -Force }
}
Write-Host "Moved CC services"

# CC types
$ccTypes = @('dsr.ts','consent.ts','governance.ts','breach.ts','datasource.ts','discovery.ts','analytics.ts','dashboard.ts','darkPattern.ts','identity.ts','lineage.ts','grievance.ts','notification.ts','translation.ts')
foreach ($t in $ccTypes) {
    if (Test-Path "src/types/$t") { Copy-Item "src/types/$t" "packages/control-centre/src/types/$t" -Force }
}
Write-Host "Moved CC types"

# CC hooks
$ccHooks = @('useBreach.ts','useConsent.ts','useDSR.ts','useDataSources.ts','useDiscovery.ts')
foreach ($h in $ccHooks) {
    if (Test-Path "src/hooks/$h") { Copy-Item "src/hooks/$h" "packages/control-centre/src/hooks/$h" -Force }
}
Write-Host "Moved CC hooks"

# CC assets
Copy-Item src/assets/* packages/control-centre/src/assets/ -Force
Write-Host "Moved CC assets"

# Step 4: Move Admin files
Copy-Item src/index.css packages/admin/src/index.css -Force
Copy-Item src/main.tsx packages/admin/src/main.tsx -Force

# Admin pages
Copy-Item src/pages/Admin/Dashboard.tsx packages/admin/src/pages/ -Force
Copy-Item src/pages/Admin/Tenants/TenantList.tsx packages/admin/src/pages/Tenants/ -Force
Copy-Item src/pages/Admin/Tenants/TenantForm.tsx packages/admin/src/pages/Tenants/ -Force
Copy-Item src/pages/Admin/Users/UserList.tsx packages/admin/src/pages/Users/ -Force
Copy-Item src/pages/Admin/Users/RoleAssignModal.tsx packages/admin/src/pages/Users/ -Force
Copy-Item src/pages/Admin/Compliance/DSRList.tsx packages/admin/src/pages/Compliance/ -Force
Copy-Item src/pages/Admin/Compliance/DSRDetail.tsx packages/admin/src/pages/Compliance/ -Force
Write-Host "Moved Admin pages"

# Admin layouts + components
Copy-Item src/components/Layout/AdminLayout.tsx packages/admin/src/components/Layout/ -Force
Copy-Item src/components/Layout/AdminSidebar.tsx packages/admin/src/components/Layout/ -Force
Copy-Item src/components/common/AdminRoute.tsx packages/admin/src/components/AdminRoute.tsx -Force
Write-Host "Moved Admin layouts/components"

# Admin services + types
Copy-Item src/services/adminService.ts packages/admin/src/services/ -Force
Copy-Item src/types/admin.ts packages/admin/src/types/ -Force
Write-Host "Moved Admin services/types"

# Copy Login page to Admin (shared auth flow)
Copy-Item src/pages/Login.tsx packages/admin/src/pages/Login.tsx -Force
Write-Host "Copied Login page to Admin"

# Step 5: Move Portal files
Copy-Item src/index.css packages/portal/src/index.css -Force
Copy-Item src/main.tsx packages/portal/src/main.tsx -Force

# Portal pages
Copy-Item src/pages/Portal/Login.tsx packages/portal/src/pages/Login.tsx -Force
Copy-Item src/pages/Portal/Dashboard.tsx packages/portal/src/pages/Dashboard.tsx -Force
Copy-Item src/pages/Portal/History.tsx packages/portal/src/pages/History.tsx -Force
Copy-Item src/pages/Portal/Requests.tsx packages/portal/src/pages/Requests.tsx -Force
Copy-Item src/pages/Portal/RequestNew.tsx packages/portal/src/pages/RequestNew.tsx -Force
Copy-Item src/pages/Portal/Profile.tsx packages/portal/src/pages/Profile.tsx -Force
Copy-Item src/pages/Portal/ConsentManage.tsx packages/portal/src/pages/ConsentManage.tsx -Force
Copy-Item src/pages/Portal/Grievance/SubmitGrievance.tsx packages/portal/src/pages/Grievance/ -Force
Copy-Item src/pages/Portal/Grievance/MyGrievances.tsx packages/portal/src/pages/Grievance/ -Force
Write-Host "Moved Portal pages"

# Portal layout + components
Copy-Item src/components/Layout/PortalLayout.tsx packages/portal/src/components/PortalLayout.tsx -Force
Copy-Item src/components/Portal/PortalProtectedRoute.tsx packages/portal/src/components/PortalProtectedRoute.tsx -Force
Copy-Item src/components/Portal/DPRRequestModal.tsx packages/portal/src/components/DPRRequestModal.tsx -Force
Copy-Item src/components/Portal/GuardianVerifyModal.tsx packages/portal/src/components/GuardianVerifyModal.tsx -Force
Copy-Item src/components/Portal/IdentityCard.tsx packages/portal/src/components/IdentityCard.tsx -Force
Copy-Item src/components/common/OTPInput.tsx packages/portal/src/components/OTPInput.tsx -Force
Write-Host "Moved Portal components"

# Portal services + types + store
Copy-Item src/services/portalApi.ts packages/portal/src/services/ -Force
Copy-Item src/services/portalService.ts packages/portal/src/services/ -Force
Copy-Item src/types/portal.ts packages/portal/src/types/ -Force
Copy-Item src/stores/portalAuthStore.ts packages/portal/src/stores/ -Force
Write-Host "Moved Portal services/types/store"

Write-Host "`n=== ALL FILE MOVES COMPLETE ==="
