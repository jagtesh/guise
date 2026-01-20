$ErrorActionPreference = "Stop"

$Repo = "jagtesh/guise"
$Binary = "guise.exe"
$Arch = "x86_64" # Defaulting to x64 for Windows
$File = "guise_Windows_$Arch.zip"
$Url = "https://github.com/$Repo/releases/latest/download/$File"
$InstallDir = "$env:USERPROFILE\bin"

Write-Host "Downloading Guise (Windows $Arch)..."
Write-Host "URL: $Url"

# Create Install Dir
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Download to Temp
$TmpFile = "$env:TEMP\$File"
Invoke-RestMethod -Uri $Url -OutFile $TmpFile

# Extract
Expand-Archive -Path $TmpFile -DestinationPath $env:TEMP -Force
Move-Item -Force "$env:TEMP\guise.exe" "$InstallDir\guise.exe"

# Add to PATH
if ($env:Path -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$InstallDir", [EnvironmentVariableTarget]::User)
    $env:Path += ";$InstallDir"
}

# Cleanup
Remove-Item $TmpFile
Write-Host "Success! Run 'guise' to start."