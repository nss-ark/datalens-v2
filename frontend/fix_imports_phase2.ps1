# Fix Phase 2: Remaining import issues identified by tsc

# ============================================================
# 1. Fix SHARED package internal imports
# ============================================================

$sharedDir = "packages/shared/src"
$sharedFiles = Get-ChildItem -Path $sharedDir -Recurse -Include *.tsx, *.ts -File

foreach ($file in $sharedFiles) {
    $content = Get-Content $file.FullName -Raw
    if (-not $content) { continue }
    $original = $content

    # @/lib/utils -> relative path to lib/utils within shared
    # Calculate relative path from file to shared/src/lib/utils
    $relativePath = [System.IO.Path]::GetDirectoryName($file.FullName)
    $sharedSrcPath = (Resolve-Path "$sharedDir").Path

    # For files in shared/src/ui/ -> ../lib/utils
    # For files in shared/src/components/ -> ../lib/utils
    # For files in shared/src/components/DataTable/ -> ../../lib/utils
    $depth = ($relativePath.Replace($sharedSrcPath, '').Split([System.IO.Path]::DirectorySeparatorChar) | Where-Object { $_ -ne '' }).Count

    $prefix = ''
    for ($i = 0; $i -lt $depth; $i++) { $prefix += '../' }
    if ($prefix -eq '') { $prefix = './' }

    # @/lib/utils
    $content = $content -replace "from\s+['""]@/lib/utils['""]", "from '${prefix}lib/utils'"

    # @/components/ui/badge (form.tsx imports label)
    $content = $content -replace "from\s+['""]@/components/ui/(\w+)['""]", "from '${prefix}ui/`$1'"

    # ../../utils/cn -> ../lib/utils (wrong relative path from old structure)
    $content = $content -replace "from\s+['""](\.\./)*utils/cn['""]", "from '${prefix}lib/utils'"

    # ../../stores/toastStore -> ../stores/toastStore
    $content = $content -replace "from\s+['""](\.\./)*stores/toastStore['""]", "from '${prefix}stores/toastStore'"
    $content = $content -replace "from\s+['""](\.\./)*stores/authStore['""]", "from '${prefix}stores/authStore'"
    $content = $content -replace "from\s+['""](\.\./)*types/auth['""]", "from '${prefix}types/auth'"
    $content = $content -replace "from\s+['""](\.\./)*types/common['""]", "from '${prefix}types/common'"
    $content = $content -replace "from\s+['""](\.\./)*services/api['""]", "from '${prefix}services/api'"

    if ($content -ne $original) {
        Set-Content $file.FullName -Value $content -NoNewline
        Write-Host "Fixed shared: $($file.Name)"
    }
}

# ============================================================
# 2. Fix CC remaining relative UI/common imports (deeper paths)
# ============================================================

$ccDir = "packages/control-centre/src"
$ccFiles = Get-ChildItem -Path $ccDir -Recurse -Include *.tsx, *.ts -File

foreach ($file in $ccFiles) {
    $content = Get-Content $file.FullName -Raw
    if (-not $content) { continue }
    $original = $content

    # ../../components/ui/xxx -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*components/ui/(\w+)['""]", "from '@datalens/shared'"

    # ../../components/common/xxx -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*components/common/(\w+)['""]", "from '@datalens/shared'"

    # ../../components/DataTable/xxx -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*components/DataTable/(\w+)['""]", "from '@datalens/shared'"

    if ($content -ne $original) {
        Set-Content $file.FullName -Value $content -NoNewline
        Write-Host "Fixed CC: $($file.Name)"
    }
}

# ============================================================
# 3. Fix Admin remaining imports
# ============================================================

$adminDir = "packages/admin/src"
$adminFiles = Get-ChildItem -Path $adminDir -Recurse -Include *.tsx, *.ts -File

foreach ($file in $adminFiles) {
    $content = Get-Content $file.FullName -Raw
    if (-not $content) { continue }
    $original = $content

    $content = $content -replace "from\s+['""](\.\./)*components/ui/(\w+)['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*components/common/(\w+)['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*components/DataTable/(\w+)['""]", "from '@datalens/shared'"

    if ($content -ne $original) {
        Set-Content $file.FullName -Value $content -NoNewline
        Write-Host "Fixed Admin: $($file.Name)"
    }
}

# ============================================================
# 4. Fix Portal remaining imports  
# ============================================================

$portalDir = "packages/portal/src"
$portalFiles = Get-ChildItem -Path $portalDir -Recurse -Include *.tsx, *.ts -File

foreach ($file in $portalFiles) {
    $content = Get-Content $file.FullName -Raw
    if (-not $content) { continue }
    $original = $content

    $content = $content -replace "from\s+['""](\.\./)*components/ui/(\w+)['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*components/common/(\w+)['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*components/DataTable/(\w+)['""]", "from '@datalens/shared'"

    if ($content -ne $original) {
        Set-Content $file.FullName -Value $content -NoNewline
        Write-Host "Fixed Portal: $($file.Name)"
    }
}

Write-Host "`n=== PHASE 2 IMPORT FIXES COMPLETE ==="
