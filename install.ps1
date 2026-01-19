$ErrorActionPreference = "Stop"

Write-Host "Installing Guise..."

# Check for Go
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "Go detected. Building from source..."
    go build -o guise.exe main.go
    
    # Create a bin directory if it doesn't exist
    $InstallDir = "$env:USERPROFILE\bin"
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    }
    
    Move-Item -Force guise.exe "$InstallDir\guise.exe"
    
    # Add to path if not present
    if ($env:Path -notlike "*$InstallDir*") {
        Write-Host "Adding $InstallDir to PATH..."
        [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$InstallDir", [EnvironmentVariableTarget]::User)
        $env:Path += ";$InstallDir"
    }
    
    Write-Host "Success! Run 'guise' to start."
} else {
    Write-Host "Error: Pre-built binaries not yet hosted. Please install Go to build from source."
    Exit 1
}
