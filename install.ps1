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

$asset = "ccolon-windows-$arch.exe"
$url = "https://github.com/$repo/releases/download/$version/$asset"
$checksumsUrl = "https://github.com/$repo/releases/download/$version/checksums.txt"

Write-Host "Downloading CColon $version..."

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$outPath = Join-Path $installDir $binaryName
Invoke-WebRequest -Uri $url -OutFile $outPath -UseBasicParsing

# Verify SHA256 checksum
Write-Host "Verifying checksum..."
try {
    $checksumsPath = Join-Path $env:TEMP "ccolon_checksums.txt"
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
    $checksumContent = Get-Content $checksumsPath
    $expectedLine = $checksumContent | Where-Object { $_ -match $asset }
    if ($expectedLine) {
        $expectedHash = ($expectedLine -split '\s+')[0]
        $actualHash = (Get-FileHash -Path $outPath -Algorithm SHA256).Hash.ToLower()
        if ($expectedHash -ne $actualHash) {
            Write-Error "Checksum verification FAILED!`n  Expected: $expectedHash`n  Got:      $actualHash"
            Remove-Item $outPath -Force
            Remove-Item $checksumsPath -Force -ErrorAction SilentlyContinue
            exit 1
        }
        Write-Host "Checksum verified OK."
    } else {
        Write-Host "Warning: asset not found in checksums.txt, skipping verification."
    }
    Remove-Item $checksumsPath -Force -ErrorAction SilentlyContinue
} catch {
    Write-Host "Warning: checksums.txt not available for this release, skipping verification."
}

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH. Restart your terminal for changes to take effect."
}

Write-Host "CColon $version installed to $outPath"
Write-Host "Run 'ccolon' to start the interactive shell."
