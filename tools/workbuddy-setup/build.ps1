$ErrorActionPreference = 'Stop'
$previousGOOS = $env:GOOS
$previousGOARCH = $env:GOARCH
$previousCGO = $env:CGO_ENABLED
Push-Location $PSScriptRoot
try {
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'WorkBuddy helper tests failed' }
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    $target = Join-Path $PSScriptRoot '../../frontend/public/downloads/sub2api-workbuddy-setup-windows-amd64.exe'
    New-Item -ItemType Directory -Force -Path (Split-Path $target) | Out-Null
    go build -trimpath -ldflags='-s -w' -o $target .
    if ($LASTEXITCODE -ne 0) { throw 'WorkBuddy helper build failed' }
    Get-FileHash -LiteralPath $target -Algorithm SHA256
} finally {
    $env:GOOS = $previousGOOS
    $env:GOARCH = $previousGOARCH
    $env:CGO_ENABLED = $previousCGO
    Pop-Location
}
