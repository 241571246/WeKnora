[CmdletBinding()]
param(
    [string]$AppContainer = 'WeKnora-app',
    [string]$FrontendContainer = 'WeKnora-frontend',
    [string]$PostgresContainer = 'WeKnora-postgres',
    [string]$BackendImage = 'vone/weknora-app:0.7.2.1-candidate.3693b799',
    [string]$FrontendImage = 'vone/weknora-ui:0.7.2.1-candidate.3693b799',
    [int]$RequiredIdentityCount = 9
)

$ErrorActionPreference = 'Stop'
$releaseDir = $PSScriptRoot
$workspace = (Resolve-Path -LiteralPath (Join-Path $releaseDir '..\..\..\..')).Path
$blockers = [System.Collections.Generic.List[string]]::new()
$errors = [System.Collections.Generic.List[string]]::new()

$expectedBackendHash = '4D3C00EFFD5E423BBEDB11ADD94A6EDB6CB88467661902DE8F42423CE18951BB'
$expectedFrontendHash = '189E3FEC361611A21C8F3E365F34AD81FD317D09CE6A9A9A009CF6E92B83AC4A'
$backendCandidate = Join-Path $workspace 'docs\vone\runtime\TASK-20260817-QA-ROCKL-01\backend\weknora-server'
$frontendCandidate = Join-Path $workspace 'docs\vone\runtime\TASK-20260817-QA-ROCKL-01\frontend\weknora-frontend-VONE-0.7.2.1-candidate1.zip'

function Add-Blocker([string]$Code, [string]$Detail) {
    $blockers.Add("$Code|$Detail")
}

function Add-CheckError([string]$Code, [string]$Detail) {
    $errors.Add("$Code|$Detail")
}

function Get-ContainerState([string]$Name) {
    $value = & docker inspect $Name --format '{{.State.Status}}' 2>$null
    if ($LASTEXITCODE -ne 0) {
        Add-CheckError 'CONTAINER_INSPECT_FAILED' $Name
        return $null
    }
    return "$value".Trim()
}

function Get-ContainerImage([string]$Name) {
    $value = & docker inspect $Name --format '{{.Config.Image}}' 2>$null
    if ($LASTEXITCODE -ne 0) {
        Add-CheckError 'CONTAINER_IMAGE_READ_FAILED' $Name
        return $null
    }
    return "$value".Trim()
}

function Invoke-DatabaseScalar([string]$Sql) {
    $command = 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At'
    $value = $Sql | & docker exec -i $PostgresContainer sh -lc $command 2>$null
    if ($LASTEXITCODE -ne 0) {
        Add-CheckError 'DATABASE_QUERY_FAILED' $Sql
        return $null
    }
    return "$value".Trim()
}

foreach ($candidate in @(
    @{ Path = $backendCandidate; Expected = $expectedBackendHash; Code = 'BACKEND_CANDIDATE_HASH_MISMATCH' },
    @{ Path = $frontendCandidate; Expected = $expectedFrontendHash; Code = 'FRONTEND_CANDIDATE_HASH_MISMATCH' }
)) {
    if (-not (Test-Path -LiteralPath $candidate.Path -PathType Leaf)) {
        Add-CheckError 'CANDIDATE_MISSING' $candidate.Path
        continue
    }
    $actualHash = (Get-FileHash -LiteralPath $candidate.Path -Algorithm SHA256).Hash
    if ($actualHash -ne $candidate.Expected) {
        Add-CheckError $candidate.Code "expected=$($candidate.Expected),actual=$actualHash"
    }
}

foreach ($container in @($AppContainer, $FrontendContainer, $PostgresContainer)) {
    $state = Get-ContainerState $container
    Write-Output "CONTAINER_STATE=$container|$state"
    if ($state -and $state -ne 'running') {
        Add-Blocker 'CONTAINER_NOT_RUNNING' "$container=$state"
    }
}

$actualBackendImage = Get-ContainerImage $AppContainer
$actualFrontendImage = Get-ContainerImage $FrontendContainer
Write-Output "BACKEND_IMAGE=$actualBackendImage"
Write-Output "FRONTEND_IMAGE=$actualFrontendImage"
if ($actualBackendImage -and $actualBackendImage -ne $BackendImage) {
    Add-Blocker 'BACKEND_IMAGE_MISMATCH' "expected=$BackendImage,actual=$actualBackendImage"
}
if ($actualFrontendImage -and $actualFrontendImage -ne $FrontendImage) {
    Add-Blocker 'FRONTEND_IMAGE_MISMATCH' "expected=$FrontendImage,actual=$actualFrontendImage"
}

$runningBackendHash = & docker exec $AppContainer sh -lc 'sha256sum /app/WeKnora | cut -d" " -f1' 2>$null
if ($LASTEXITCODE -ne 0) {
    Add-CheckError 'RUNNING_BACKEND_HASH_READ_FAILED' $AppContainer
} else {
    $runningBackendHash = "$runningBackendHash".Trim().ToUpperInvariant()
    Write-Output "RUNNING_BACKEND_SHA256=$runningBackendHash"
    if ($runningBackendHash -ne $expectedBackendHash) {
        Add-Blocker 'RUNNING_BACKEND_HASH_MISMATCH' "expected=$expectedBackendHash,actual=$runningBackendHash"
    }
}

$aclMode = & docker inspect $AppContainer --format '{{range .Config.Env}}{{println .}}{{end}}' 2>$null |
    Where-Object { $_ -like 'WEKNORA_VONE_KB_ACL_MODE=*' } |
    Select-Object -First 1
if ($LASTEXITCODE -ne 0) {
    Add-CheckError 'ACL_MODE_READ_FAILED' $AppContainer
} elseif (-not $aclMode) {
    Write-Output 'ACL_MODE=absent'
    Add-Blocker 'ACL_SHADOW_MODE_NOT_CONFIGURED' 'WEKNORA_VONE_KB_ACL_MODE is absent'
} else {
    $aclModeValue = ($aclMode -split '=', 2)[1]
    Write-Output "ACL_MODE=$aclModeValue"
    if ($aclModeValue -ne 'shadow') {
        Add-Blocker 'ACL_SHADOW_MODE_NOT_CONFIGURED' "actual=$aclModeValue"
    }
}

$healthCode = & curl.exe --noproxy '*' -sS -o NUL -w '%{http_code}' 'http://127.0.0.1:8080/health' 2>$null
if ($LASTEXITCODE -ne 0) {
    Add-CheckError 'HEALTH_PROBE_FAILED' 'http://127.0.0.1:8080/health'
} else {
    $healthCode = "$healthCode".Trim()
    Write-Output "HEALTH_HTTP_STATUS=$healthCode"
    if ($healthCode -ne '200') {
        Add-Blocker 'BACKEND_NOT_HEALTHY' "status=$healthCode"
    }
}

$voneTableCount = Invoke-DatabaseScalar @'
SELECT count(*)
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'vone_kb_memberships',
    'vone_kb_member_capabilities',
    'vone_kb_collections',
    'vone_kb_collection_bindings',
    'vone_schema_migrations'
  );
'@
$userCount = Invoke-DatabaseScalar 'SELECT count(*) FROM users;'
$tenantCount = Invoke-DatabaseScalar 'SELECT count(*) FROM tenants;'
$kbCount = Invoke-DatabaseScalar 'SELECT count(*) FROM knowledge_bases;'
$shareCount = Invoke-DatabaseScalar 'SELECT count(*) FROM kb_shares;'
$apiKeyCount = Invoke-DatabaseScalar 'SELECT count(*) FROM tenant_api_keys;'

Write-Output "VONE_TABLE_COUNT=$voneTableCount"
Write-Output "USER_COUNT=$userCount"
Write-Output "TENANT_COUNT=$tenantCount"
Write-Output "KNOWLEDGE_BASE_COUNT=$kbCount"
Write-Output "KB_SHARE_COUNT=$shareCount"
Write-Output "TENANT_API_KEY_COUNT=$apiKeyCount"

if ($null -ne $voneTableCount -and [int]$voneTableCount -ne 5) {
    Add-Blocker 'VONE_SCHEMA_NOT_READY' "expected=5,actual=$voneTableCount"
}
if ($null -ne $userCount -and [int]$userCount -lt $RequiredIdentityCount) {
    Add-Blocker 'B48_IDENTITIES_INSUFFICIENT' "required=$RequiredIdentityCount,actual=$userCount"
}
if ($null -ne $apiKeyCount -and [int]$apiKeyCount -lt 1) {
    Add-Blocker 'B48_API_KEY_FIXTURE_MISSING' "actual=$apiKeyCount"
}

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Output "ERROR=$_" }
    $blockers | ForEach-Object { Write-Output "BLOCKER=$_" }
    Write-Output 'B48_STAGING_PREFLIGHT=ERROR'
    exit 1
}

if ($blockers.Count -gt 0) {
    $blockers | ForEach-Object { Write-Output "BLOCKER=$_" }
    Write-Output 'B48_STAGING_PREFLIGHT=BLOCKED'
    exit 2
}

Write-Output 'B48_STAGING_PREFLIGHT=PASS'
