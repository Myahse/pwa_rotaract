# Push backend secrets from .env to Fly (run from anywhere).
#   cd backend
#   powershell -ExecutionPolicy Bypass -File .\scripts\set-fly-secrets.ps1

$ErrorActionPreference = "Stop"
$backendRoot = Split-Path $PSScriptRoot -Parent
$repoRoot = Split-Path $backendRoot -Parent

$envPath = @(
  (Join-Path $repoRoot ".env"),
  (Join-Path $backendRoot ".env")
) | Where-Object { Test-Path $_ } | Select-Object -First 1

if (-not $envPath) {
  Write-Error "No .env found at repo root or backend/.env"
}

$allowed = @(
  "APP_ENV", "HOST", "CORS_ORIGINS", "LOG_LEVEL",
  "JWT_SECRET", "JWT_ACCESS_TTL", "INVITE_TTL", "PASSWORD_RESET_TTL",
  "DATABASE_URL", "BOOTSTRAP_ADMIN_EMAIL", "BOOTSTRAP_ADMIN_PASSWORD",
  "APP_PUBLIC_URL", "UPLOAD_DIR", "MAX_AVATAR_SIZE",
  "BIRTHDAY_TIMEZONE", "BIRTHDAY_NOTIFY_HOUR", "CRON_SECRET",
  "VAPID_PUBLIC_KEY", "VAPID_PRIVATE_KEY", "VAPID_SUBJECT",
  "BREVO_API_KEY", "BREVO_SENDER_EMAIL", "BREVO_SENDER_NAME",
  "GOOGLE_CLIENT_ID"
)

$secretMap = @{}
Get-Content $envPath | ForEach-Object {
  $line = $_.Trim()
  if ($line -eq "" -or $line.StartsWith("#")) { return }
  $idx = $line.IndexOf("=")
  if ($idx -lt 1) { return }
  $key = $line.Substring(0, $idx).Trim()
  $val = $line.Substring($idx + 1).Trim().Trim('"').Trim("'")
  if ($allowed -notcontains $key) { return }
  if ($val -match "REPLACE_WITH") {
    Write-Warning "Skipping $key (placeholder — set a real value in .env)"
    return
  }
  if ($val -eq "") { return }
  $secretMap[$key] = $val
}

$toml = Join-Path $backendRoot "fly.toml"
$match = Select-String -Path $toml -Pattern "^app\s*=\s*'([^']+)'" | Select-Object -First 1
if ($match) {
  $appName = $match.Matches.Groups[1].Value
  $flyApi = "https://${appName}.fly.dev"
  $secretMap.Remove("API_PUBLIC_URL")
  $secretMap["API_PUBLIC_URL"] = $flyApi
  Write-Host "API_PUBLIC_URL -> $flyApi"
}

if ($secretMap.Count -eq 0) {
  Write-Error "No secrets to set. Fix JWT_SECRET and CRON_SECRET in .env (no REPLACE_WITH placeholders)."
}

Write-Host "Setting $($secretMap.Count) secrets from $envPath (one at a time, safe for Neon URLs with &) ..."
Set-Location $backendRoot
$fly = Join-Path $env:USERPROFILE ".fly\bin\flyctl.exe"
if (-not (Test-Path $fly)) { Write-Error "flyctl not found. Install Fly CLI first." }

# PowerShell treats & in Neon DATABASE_URL as a command separator unless quoted per secret.
foreach ($key in ($secretMap.Keys | Sort-Object)) {
  $assignment = "${key}=$($secretMap[$key])"
  Write-Host "  -> $key"
  & $fly secrets set $assignment
  if ($LASTEXITCODE -ne 0) {
    Write-Error "flyctl secrets set failed for $key"
  }
}
Write-Host "Done. Next: flyctl deploy"
