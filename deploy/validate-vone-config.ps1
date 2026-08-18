[CmdletBinding()]
param(
    [string]$EnvFile,
    [string]$ModelConfig,
    [switch]$Template
)

$ErrorActionPreference = 'Stop'
$script:Failures = [System.Collections.Generic.List[string]]::new()

if ([string]::IsNullOrWhiteSpace($EnvFile)) {
    $EnvFile = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\.env'))
}
if ([string]::IsNullOrWhiteSpace($ModelConfig)) {
    $ModelConfig = [IO.Path]::GetFullPath(
        (Join-Path $PSScriptRoot '..\config\builtin_models.vone.yaml')
    )
}

function Add-Failure {
    param([string]$Message)
    $script:Failures.Add($Message)
}

function Read-DotEnv {
    param([string]$Path)

    $values = @{}
    $lineNumber = 0
    foreach ($line in Get-Content -LiteralPath $Path) {
        $lineNumber++
        $trimmed = $line.Trim()
        if ($trimmed.Length -eq 0 -or $trimmed.StartsWith('#')) {
            continue
        }

        $separator = $line.IndexOf('=')
        if ($separator -lt 1) {
            Add-Failure "Invalid env syntax at line $lineNumber."
            continue
        }

        $name = $line.Substring(0, $separator).Trim()
        $value = $line.Substring($separator + 1).Trim()
        if ($name -notmatch '^[A-Za-z_][A-Za-z0-9_]*$') {
            Add-Failure "Invalid env variable name at line $lineNumber."
            continue
        }
        if ($values.ContainsKey($name)) {
            Add-Failure "Duplicate env variable: $name."
            continue
        }
        if (($value.StartsWith('"') -and $value.EndsWith('"')) -or
            ($value.StartsWith("'") -and $value.EndsWith("'"))) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        $values[$name] = $value
    }
    return $values
}

function Require-Key {
    param(
        [hashtable]$Values,
        [string]$Name,
        [switch]$AllowEmpty
    )

    if (-not $Values.ContainsKey($Name)) {
        Add-Failure "Missing env variable: $Name."
        return
    }
    if (-not $AllowEmpty -and [string]::IsNullOrWhiteSpace($Values[$Name])) {
        Add-Failure "Required env variable is empty: $Name."
    }
}

function Require-MinLength {
    param(
        [hashtable]$Values,
        [string]$Name,
        [int]$Minimum
    )

    if (-not $Values.ContainsKey($Name) -or
        [string]::IsNullOrWhiteSpace($Values[$Name])) {
        return
    }
    if ($Values[$Name].Length -lt $Minimum) {
        Add-Failure "$Name must contain at least $Minimum characters."
    }
}

function Require-HttpUrl {
    param(
        [hashtable]$Values,
        [string]$Name
    )

    if (-not $Values.ContainsKey($Name) -or [string]::IsNullOrWhiteSpace($Values[$Name])) {
        return
    }

    $uri = $null
    if (-not [Uri]::TryCreate($Values[$Name], [UriKind]::Absolute, [ref]$uri) -or
        $uri.Scheme -notin @('http', 'https')) {
        Add-Failure "$Name must be an absolute HTTP or HTTPS URL."
    }
}

$requiredTemplateKeys = @(
    'COMPOSE_PROJECT_NAME', 'VONE_IMAGE_TAG', 'TZ', 'GIN_MODE',
    'WEKNORA_VONE_KB_ACL_MODE',
    'DB_DRIVER', 'DB_HOST', 'DB_PORT', 'DB_USER', 'DB_PASSWORD', 'DB_NAME',
    'RETRIEVE_DRIVER', 'STREAM_MANAGER_TYPE', 'REDIS_ADDR', 'REDIS_PASSWORD',
    'REDIS_DB', 'REDIS_PREFIX', 'JWT_SECRET', 'SYSTEM_AES_KEY',
    'VONE_LLM_MODEL_NAME', 'VONE_LLM_BASE_URL', 'VONE_LLM_API_KEY',
    'VONE_LLM_PROVIDER', 'VONE_EMBEDDING_MODEL_NAME',
    'VONE_EMBEDDING_BASE_URL', 'VONE_EMBEDDING_API_KEY',
    'VONE_EMBEDDING_PROVIDER', 'VONE_RERANK_MODEL_NAME',
    'VONE_RERANK_BASE_URL', 'VONE_RERANK_API_KEY', 'VONE_RERANK_PROVIDER'
)
$templateSecretKeys = @(
    'DB_PASSWORD', 'REDIS_PASSWORD', 'JWT_SECRET', 'SYSTEM_AES_KEY',
    'VONE_LLM_API_KEY', 'VONE_EMBEDDING_API_KEY', 'VONE_RERANK_API_KEY'
)

if (-not (Test-Path -LiteralPath $EnvFile -PathType Leaf)) {
    Add-Failure "Environment file not found: $EnvFile."
    $envValues = @{}
} else {
    $envValues = Read-DotEnv -Path $EnvFile
    foreach ($key in $requiredTemplateKeys) {
        $allowTemplateSecret = $Template -and $key -in $templateSecretKeys
        Require-Key -Values $envValues -Name $key -AllowEmpty:$allowTemplateSecret
    }
}

if ($envValues.Count -gt 0 -and -not $Template) {
    if ($envValues['WEKNORA_VONE_KB_ACL_MODE'] -notin @('shadow', 'enforce', 'off')) {
        Add-Failure 'WEKNORA_VONE_KB_ACL_MODE must be shadow, enforce, or off.'
    }
    if ($envValues['DB_DRIVER'] -ne 'postgres') {
        Add-Failure 'DB_DRIVER must be postgres for the Vone-weknora Docker deployment.'
    }
    if ($envValues['RETRIEVE_DRIVER'] -notmatch '(^|,)postgres(,|$)') {
        Add-Failure 'RETRIEVE_DRIVER must include postgres.'
    }
    if ($envValues['STREAM_MANAGER_TYPE'] -ne 'redis') {
        Add-Failure 'STREAM_MANAGER_TYPE must be redis.'
    }

    Require-MinLength -Values $envValues -Name 'DB_PASSWORD' -Minimum 16
    Require-MinLength -Values $envValues -Name 'REDIS_PASSWORD' -Minimum 16
    Require-MinLength -Values $envValues -Name 'JWT_SECRET' -Minimum 32
    Require-MinLength -Values $envValues -Name 'VONE_LLM_API_KEY' -Minimum 8
    Require-MinLength -Values $envValues -Name 'VONE_EMBEDDING_API_KEY' -Minimum 8

    if ($envValues.ContainsKey('SYSTEM_AES_KEY') -and
        -not [string]::IsNullOrWhiteSpace($envValues['SYSTEM_AES_KEY'])) {
        $aesByteCount = [Text.Encoding]::UTF8.GetByteCount($envValues['SYSTEM_AES_KEY'])
        if ($aesByteCount -ne 32) {
            Add-Failure 'SYSTEM_AES_KEY must be exactly 32 UTF-8 bytes.'
        }
    }

    Require-HttpUrl -Values $envValues -Name 'VONE_LLM_BASE_URL'
    Require-HttpUrl -Values $envValues -Name 'VONE_EMBEDDING_BASE_URL'
    Require-Key -Values $envValues -Name 'VONE_LLM_MODEL_NAME'
    Require-Key -Values $envValues -Name 'VONE_LLM_PROVIDER'
    Require-Key -Values $envValues -Name 'VONE_EMBEDDING_MODEL_NAME'
    Require-Key -Values $envValues -Name 'VONE_EMBEDDING_PROVIDER'

}

if (-not (Test-Path -LiteralPath $ModelConfig -PathType Leaf)) {
    Add-Failure "Model configuration not found: $ModelConfig."
} else {
    $activeModelLines = Get-Content -LiteralPath $ModelConfig |
        Where-Object { -not $_.TrimStart().StartsWith('#') }
    $activeModelText = $activeModelLines -join "`n"

    if ($activeModelText -match '(?m)^\s*builtin_models:\s*\[\]\s*$') {
        Add-Failure 'builtin_models.vone.yaml is still inactive (builtin_models: []).'
    }
    foreach ($reference in @(
        '${VONE_LLM_MODEL_NAME}', '${VONE_LLM_API_KEY}',
        '${VONE_EMBEDDING_MODEL_NAME}', '${VONE_EMBEDDING_API_KEY}',
        '${VONE_RERANK_MODEL_NAME}', '${VONE_RERANK_API_KEY}'
    )) {
        if (-not $activeModelText.Contains($reference)) {
            Add-Failure "Active model YAML is missing reference $reference."
        }
    }
    $dimensionMatch = [regex]::Match(
        $activeModelText,
        '(?m)^\s*dimension:\s*(\d+)\s*(?:#.*)?$'
    )
    if (-not $dimensionMatch.Success -or [int]$dimensionMatch.Groups[1].Value -lt 1) {
        Add-Failure 'Active embedding model must contain a positive integer dimension.'
    }

    if (-not $Template -and $envValues.Count -gt 0) {
        Require-MinLength -Values $envValues -Name 'VONE_RERANK_API_KEY' -Minimum 8
        Require-HttpUrl -Values $envValues -Name 'VONE_RERANK_BASE_URL'
    }
}

$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$baseCompose = Join-Path $repoRoot 'docker-compose.yml'
$voneCompose = Join-Path $repoRoot 'deploy\docker-compose.vone.yml'
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Add-Failure 'Docker CLI is not available.'
} elseif ((Test-Path -LiteralPath $EnvFile) -and
          (Test-Path -LiteralPath $baseCompose) -and
          (Test-Path -LiteralPath $voneCompose)) {
    & docker compose --env-file $EnvFile `
        -f $baseCompose -f $voneCompose `
        config --no-env-resolution --quiet *> $null
    if ($LASTEXITCODE -ne 0) {
        Add-Failure 'Merged Docker Compose configuration is invalid.'
    }
}

if ($script:Failures.Count -gt 0) {
    Write-Host "[FAIL] Vone-weknora preflight found $($script:Failures.Count) issue(s)."
    foreach ($failure in $script:Failures) {
        Write-Host "  - $failure"
    }
    Write-Host 'No secret values were displayed. No Docker resources were changed.'
    exit 1
}

if ($Template) {
    Write-Host '[PASS] Vone-weknora configuration template is structurally valid.'
} else {
    Write-Host '[PASS] Vone-weknora deployment configuration passed preflight.'
}
Write-Host 'Secret values were not displayed. No Docker resources were changed.'
