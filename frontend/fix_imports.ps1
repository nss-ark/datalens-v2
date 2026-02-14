# Fix imports in shared package files
# In shared, internal imports should use relative paths (within the package)
# Files in shared/src/services/api.ts import from '../stores/authStore' - this is fine (relative within shared)
# Files in shared/src/services/auth.ts import from './api' and '../types/auth', '../types/common' - fine
# Files in shared/src/stores/authStore.ts imports from '../types/auth' - fine
# Files in shared/src/hooks/useAuth.ts imports from '../services/auth', '../stores/authStore' - fine

# No changes needed in shared - relative imports within same package work

# ============================================================
# Fix imports in CONTROL-CENTRE package
# Pattern: replace imports from '../services/api' with '@datalens/shared'
# Pattern: replace imports from '../types/common' with '@datalens/shared'
# Pattern: replace imports from '../stores/authStore' with '@datalens/shared'
# Pattern: replace imports from '../stores/toastStore' with '@datalens/shared'
# Pattern: replace imports from '../stores/uiStore' with '@datalens/shared'
# Pattern: replace imports from '../hooks/useAuth' with '@datalens/shared'
# Pattern: replace imports from '@/components/ui/*' with '@datalens/shared'
# Pattern: replace imports from '@/lib/utils' with '@datalens/shared'
# Pattern: replace imports from '../components/common/ErrorBoundary' with '@datalens/shared'
# etc.
# ============================================================

$ccDir = "packages/control-centre/src"
$adminDir = "packages/admin/src"
$portalDir = "packages/portal/src"

function Fix-Imports {
    param([string]$Dir, [string]$Label)

    $tsxFiles = Get-ChildItem -Path $Dir -Recurse -Include *.tsx, *.ts -File
    $count = 0

    foreach ($file in $tsxFiles) {
        $content = Get-Content $file.FullName -Raw
        if (-not $content) { continue }
        $original = $content

        # ---- Replace relative imports to shared modules ----

        # services/api → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*services/api['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*services/api\.ts['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]\./api['""]", "from '@datalens/shared'"

        # services/auth → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*services/auth['""]", "from '@datalens/shared'"

        # types/common → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*types/common['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]\./common['""]", "from '@datalens/shared'"

        # types/auth → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*types/auth['""]", "from '@datalens/shared'"

        # stores/authStore → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*stores/authStore['""]", "from '@datalens/shared'"

        # stores/toastStore → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*stores/toastStore['""]", "from '@datalens/shared'"

        # stores/uiStore → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*stores/uiStore['""]", "from '@datalens/shared'"

        # hooks/useAuth → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*hooks/useAuth['""]", "from '@datalens/shared'"

        # hooks/useMediaQuery → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*hooks/useMediaQuery['""]", "from '@datalens/shared'"

        # lib/utils (cn utility) → @datalens/shared
        $content = $content -replace "from\s+['""]@/lib/utils['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*lib/utils['""]", "from '@datalens/shared'"

        # utils/cn → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*utils/cn['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/utils/cn['""]", "from '@datalens/shared'"

        # shadcn/ui components → @datalens/shared
        $content = $content -replace "from\s+['""]@/components/ui/(\w+)['""]", "from '@datalens/shared'"

        # Common components → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*components/common/ErrorBoundary['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/common/ErrorFallbacks['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/common/Toast['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/common/StatusBadge['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/common/Button['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/common/Modal['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/ErrorBoundary['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/ErrorFallbacks['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/Toast['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/StatusBadge['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/Button['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/common/Modal['""]", "from '@datalens/shared'"

        # DataTable → @datalens/shared
        $content = $content -replace "from\s+['""](\.\./)*components/DataTable/DataTable['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""](\.\./)*components/DataTable/Pagination['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/DataTable/DataTable['""]", "from '@datalens/shared'"
        $content = $content -replace "from\s+['""]@/components/DataTable/Pagination['""]", "from '@datalens/shared'"

        if ($content -ne $original) {
            Set-Content $file.FullName -Value $content -NoNewline
            $count++
        }
    }

    Write-Host "$Label : $count files updated"
}

# Fix CC imports
Fix-Imports -Dir $ccDir -Label "Control Centre"

# Fix Admin imports
Fix-Imports -Dir $adminDir -Label "Admin"

# Fix Portal imports
Fix-Imports -Dir $portalDir -Label "Portal"

Write-Host "`n=== IMPORT REWRITING COMPLETE ==="
