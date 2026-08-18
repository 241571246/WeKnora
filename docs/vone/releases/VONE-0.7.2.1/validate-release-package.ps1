[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$releaseDir = $PSScriptRoot
$workspace = (Resolve-Path -LiteralPath (Join-Path $releaseDir '..\..\..\..')).Path
$failures = [System.Collections.Generic.List[string]]::new()

function Require-File([string]$RelativePath) {
    $path = Join-Path $workspace $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        $failures.Add("missing file: $RelativePath")
    }
    return $path
}

$manifest = Require-File 'docs\vone\releases\VONE-0.7.2.1\release-manifest.md'
$runbook = Require-File 'docs\vone\releases\VONE-0.7.2.1\deployment-rollback-runbook.md'
$initialDeploymentGuide = Require-File 'docs\vone\releases\VONE-0.7.2.1\initial-deployment-guide.md'
$reconciliation = Require-File 'docs\vone\releases\VONE-0.7.2.1\reconciliation-postgresql.sql'
$b48Preflight = Require-File 'docs\vone\releases\VONE-0.7.2.1\b48-staging-preflight.md'
$b48Validator = Require-File 'docs\vone\releases\VONE-0.7.2.1\validate-b48-staging.ps1'
$upMigration = Require-File 'migrations\vone\versioned\000001_kb_acl_and_collections.up.sql'
$downMigration = Require-File 'migrations\vone\versioned\000001_kb_acl_and_collections.down.sql'
$sqliteMigration = Require-File 'migrations\vone\sqlite\000001_kb_acl_and_collections.up.sql'
$sqliteDownMigration = Require-File 'migrations\vone\sqlite\000001_kb_acl_and_collections.down.sql'
$capabilitySource = Require-File 'internal\types\kb_acl.go'
$systemInfoSource = Require-File 'internal\handler\system.go'

if ($failures.Count -eq 0) {
    $manifestText = Get-Content -LiteralPath $manifest -Raw
    foreach ($section in @('Metadata','Scope','Artifacts','Environment Changes','QA Independence','Gate Evidence','Deployment Plan','Rollback Plan','Production Verification','Approvals','Open Blockers')) {
        if ($manifestText -notmatch "(?m)^## $([regex]::Escape($section))\s*$") {
            $failures.Add("release manifest missing section: $section")
        }
    }

    $capabilityText = Get-Content -LiteralPath $capabilitySource -Raw
    $capabilities = [regex]::Matches($capabilityText, 'KBCapability[A-Za-z]+\s+KBCapability\s*=\s*"([^"]+)"') |
        ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique
    if ($capabilities.Count -ne 17) {
        $failures.Add("capability contract has $($capabilities.Count) values; expected 17")
    }
    $upText = Get-Content -LiteralPath $upMigration -Raw
    foreach ($capability in $capabilities) {
        if ($upText -notmatch [regex]::Escape("'$capability'")) {
            $failures.Add("PostgreSQL migration omits capability: $capability")
        }
    }
    if ($upText -match '(?im)^\s*(INSERT|UPDATE|DELETE|ALTER|DROP|TRUNCATE)\b[^;]*\bkb_shares\b') {
        $failures.Add('VONE migration mutates retained kb_shares')
    }

    $expectedDownTables = @(
        'vone_kb_collection_bindings',
        'vone_kb_collections',
        'vone_kb_member_capabilities',
        'vone_kb_memberships'
    )
    $downText = Get-Content -LiteralPath $downMigration -Raw
    $actualDownTables = [regex]::Matches($downText, '(?im)^DROP TABLE IF EXISTS\s+([a-z0-9_]+)\s*;') |
        ForEach-Object { $_.Groups[1].Value }
    if ((Compare-Object $expectedDownTables $actualDownTables).Count -ne 0) {
        $failures.Add('down migration drops an unexpected set of tables')
    }

    $diagnostics = Get-Content -LiteralPath $systemInfoSource -Raw
    foreach ($field in @('vone_db_version','vone_db_error','kb_acl_mode','kb_acl_metrics')) {
        if ($diagnostics -notmatch [regex]::Escape($field)) {
            $failures.Add("system diagnostics omit: $field")
        }
    }

    try {
        $null = [scriptblock]::Create((Get-Content -LiteralPath $b48Validator -Raw))
    } catch {
        $failures.Add("B48 staging validator has invalid PowerShell syntax: $($_.Exception.Message)")
    }

    Push-Location $workspace
    try {
        $upstreamMigrationChanges = @(git diff --name-only -- migrations ':!migrations/vone')
        if ($LASTEXITCODE -ne 0) { $failures.Add('git could not inspect upstream migration changes') }
        if ($upstreamMigrationChanges.Count -gt 0) {
            $failures.Add("upstream migration files changed: $($upstreamMigrationChanges -join ', ')")
        }
        git diff --check | Out-Null
        if ($LASTEXITCODE -ne 0) { $failures.Add('git diff --check failed') }
    } finally {
        Pop-Location
    }
}

if ($failures.Count -gt 0) {
    $failures | ForEach-Object { Write-Error $_ }
    Write-Output 'RELEASE_PACKAGE_VALIDATION=FAIL'
    exit 1
}

Write-Output 'RELEASE_PACKAGE_VALIDATION=PASS'
foreach ($path in @($manifest,$runbook,$initialDeploymentGuide,$reconciliation,$b48Preflight,$b48Validator,$upMigration,$downMigration,$sqliteMigration,$sqliteDownMigration)) {
    $hash = Get-FileHash -LiteralPath $path -Algorithm SHA256
    Write-Output ("SHA256 {0} {1}" -f $hash.Hash, (Resolve-Path -Relative $path))
}
