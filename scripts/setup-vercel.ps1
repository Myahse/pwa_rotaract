# Creates / links four Vercel projects for this monorepo (admin, frontend, website, social).
# Prerequisite: npx vercel login
#
# Usage (from repo root):
#   .\scripts\setup-vercel.ps1
# Optional:
#   .\scripts\setup-vercel.ps1 -Deploy

param(
  [switch]$Deploy,
  [string]$Scope = "myahses-projects"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

$projects = @(
  @{
    Dir = "admin"
    Name = "rotaract-admin"
    Domain = "admin.rotaractiugb.com"
    EnvFile = ".env"
  },
  @{
    Dir = "frontend"
    Name = "rotaract-app"
    Domain = "app.rotaractiugb.com"
    EnvFile = ".env"
  },
  @{
    Dir = "website"
    Name = "rotaract-website"
    Domain = "www.rotaractiugb.com"
    EnvFile = ".env"
  },
  @{
    Dir = "social"
    Name = "rotaract-social"
    Domain = "social.rotaractiugb.com"
    EnvFile = ".env"
  }
)

Write-Host "Checking Vercel login..." -ForegroundColor Cyan
npx vercel whoami | Out-Null
if ($LASTEXITCODE -ne 0) {
  Write-Host "Not logged in. Run: npx vercel login" -ForegroundColor Yellow
  exit 1
}

foreach ($project in $projects) {
  $path = Join-Path $Root $project.Dir
  Write-Host ""
  Write-Host "=== $($project.Name) ($($project.Dir)) ===" -ForegroundColor Green

  Push-Location $path
  try {
    if (-not (Test-Path "node_modules")) {
      Write-Host "Installing dependencies..."
      npm install
    }

    Write-Host "Configuring Vercel project root directory..."
    npx vercel project update $project.Name --root-directory $project.Dir --framework vite --output-directory dist --build-command "npm run build" --yes --scope $Scope | Out-Null

    Write-Host "Linking Vercel project (scope: $Scope)..."
    npx vercel link --yes --project $project.Name --scope $Scope
    if ($LASTEXITCODE -ne 0) {
      Write-Host "Creating project on first deploy..."
      npx vercel --yes --scope $Scope
    }

    if ($Deploy) {
      Write-Host "Deploying to production..."
      npx vercel deploy --prod --yes --scope $Scope
    } else {
      Write-Host "Skipped deploy (pass -Deploy to publish)."
    }

    Write-Host "Suggested domain: $($project.Domain)" -ForegroundColor DarkGray
    Write-Host "Set env vars in Vercel dashboard from $($project.Dir)/$($project.EnvFile)" -ForegroundColor DarkGray
  }
  finally {
    Pop-Location
  }
}

Write-Host ""
Write-Host "Done. Root Directory is set automatically for each project." -ForegroundColor Cyan
Write-Host "API rewrite target: https://pwa-rotaract.onrender.com (not api.rotaractiugb.com — that is Tombola)." -ForegroundColor DarkGray
