# Fix Phase 3: Component-to-component relative imports within CC components/
# These are relative imports like '../common/Button', '../ui/button', '../DataTable/DataTable'
# from files inside packages/control-centre/src/components/*/

$ccComponentsDir = "packages/control-centre/src/components"
$files = Get-ChildItem -Path $ccComponentsDir -Recurse -Include *.tsx, *.ts -File

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    if (-not $content) { continue }
    $original = $content

    # ../common/Button, ../../common/Button -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*common/Button['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/Modal['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/StatusBadge['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/Toast['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/ErrorBoundary['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/ErrorFallbacks['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*common/OTPInput['""]", "from '@datalens/shared'"

    # ../ui/button, ../ui/card etc -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*ui/button['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/card['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/input['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/label['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/select['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/textarea['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/dialog['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/form['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/badge['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*ui/table['""]", "from '@datalens/shared'"

    # ../DataTable/DataTable, ../DataTable/Pagination -> @datalens/shared
    $content = $content -replace "from\s+['""](\.\./)*DataTable/DataTable['""]", "from '@datalens/shared'"
    $content = $content -replace "from\s+['""](\.\./)*DataTable/Pagination['""]", "from '@datalens/shared'"

    if ($content -ne $original) {
        Set-Content $file.FullName -Value $content -NoNewline
        Write-Host "Fixed: $($file.Name)"
    }
}

Write-Host "`n=== PHASE 3 COMPLETE ==="
