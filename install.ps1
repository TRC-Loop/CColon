$ErrorActionPreference = "Stop"

$repo = "TRC-Loop/CColon"
$binaryName = "ccolon.exe"
$installDir = "$env:LOCALAPPDATA\CColon"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else {
    Write-Error "CColon requires a 64-bit system."
    exit 1
}

if ($args.Count -gt 0) {
    $version = $args[0]
} else {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{ "User-Agent" = "CColon-Installer" }
    $version = $release.tag_name
}

$url = "https://github.com/$repo/releases/download/$version/ccolon-windows-$arch.exe"

Write-Host "Downloading CColon $version..."

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$outPath = Join-Path $installDir $binaryName
Invoke-WebRequest -Uri $url -OutFile $outPath -UseBasicParsing

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH. Restart your terminal for changes to take effect."
}

Write-Host "CColon $version installed to $outPath"
Write-Host "Run 'ccolon' to start the interactive shell."
