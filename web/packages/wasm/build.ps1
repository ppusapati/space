# Samavaya WASM Build Script for Windows
# Builds all Rust crates to WebAssembly using wasm-pack

param(
    [ValidateSet("web", "nodejs", "bundler")]
    [string]$Target = "web",
    [switch]$Dev,
    [string]$Crate,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PkgDir = Join-Path $ScriptDir "pkg"
$CratesDir = Join-Path $ScriptDir "crates"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

if ($Help) {
    Write-Host "Samavaya WASM Build Script"
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Target <target>  Build target: web, nodejs, bundler (default: web)"
    Write-Host "  -Dev              Build in development mode (faster, larger)"
    Write-Host "  -Crate <name>     Build only specific crate"
    Write-Host "  -Help             Show this help message"
    exit 0
}

Write-Host ""
Write-ColorOutput "==================================================================" "Cyan"
Write-ColorOutput "  Samavaya WASM Build System" "Cyan"
Write-ColorOutput "==================================================================" "Cyan"
Write-Host ""

# Check for wasm-pack
try {
    $null = Get-Command wasm-pack -ErrorAction Stop
} catch {
    Write-ColorOutput "Error: wasm-pack is not installed" "Red"
    Write-ColorOutput "Install it with: cargo install wasm-pack" "Yellow"
    exit 1
}

# Check for Rust
try {
    $null = Get-Command cargo -ErrorAction Stop
} catch {
    Write-ColorOutput "Error: Rust/Cargo is not installed" "Red"
    Write-ColorOutput "Install it from: https://rustup.rs" "Yellow"
    exit 1
}

# Create output directory
if (-not (Test-Path $PkgDir)) {
    New-Item -ItemType Directory -Path $PkgDir -Force | Out-Null
}

# Build mode
$BuildFlag = if ($Dev) { "--dev" } else { "--release" }
$BuildMode = if ($Dev) { "DEVELOPMENT" } else { "RELEASE" }

Write-ColorOutput "Building in $BuildMode mode" $(if ($Dev) { "Yellow" } else { "Green" })
Write-ColorOutput "Target: $Target" "Cyan"
Write-Host ""

# Get all crates
$Crates = Get-ChildItem -Path $CratesDir -Directory | Where-Object {
    Test-Path (Join-Path $_.FullName "Cargo.toml")
}

$TotalCrates = $Crates.Count
$BuiltCrates = 0
$FailedCrates = 0
$FailedNames = @()

foreach ($CrateDir in $Crates) {
    $CrateName = $CrateDir.Name

    # Skip if filter is set and doesn't match
    if ($Crate -and $CrateName -ne $Crate) {
        continue
    }

    Write-ColorOutput "----------------------------------------------------------------" "Cyan"
    Write-ColorOutput "Building: $CrateName" "Cyan"

    $OutDir = Join-Path $PkgDir $CrateName

    # Convert crate name with hyphens to underscores for wasm file naming
    $wasmName = $CrateName -replace '-', '_'
    $pkgWasm = Join-Path $OutDir "samavaya_$($wasmName)_bg.wasm"

    # Remove old output to ensure clean build detection
    if (Test-Path $OutDir) {
        Remove-Item -Path $OutDir -Recurse -Force
    }

    # Run wasm-pack directly - stderr output is normal (INFO messages)
    $env:RUST_BACKTRACE = "1"

    # Run wasm-pack and let it output to console, redirect stderr to stdout
    $wasmPackArgs = @(
        "build",
        $CrateDir.FullName,
        "--target", $Target,
        "--out-dir", $OutDir,
        $BuildFlag,
        "--out-name", "samavaya_$wasmName"
    )

    # Execute wasm-pack, suppress stderr by redirecting to $null
    # Temporarily set ErrorActionPreference to ignore stderr as errors
    $oldErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "SilentlyContinue"
    & wasm-pack @wasmPackArgs 2>$null
    $ErrorActionPreference = $oldErrorActionPreference

    # Check for the wasm file to verify success (most reliable check)
    if (Test-Path $pkgWasm) {
        Write-ColorOutput "[OK] Built: $CrateName" "Green"
        $BuiltCrates++

        # Clean up unnecessary files
        $gitignore = Join-Path $OutDir ".gitignore"
        $packageLock = Join-Path $OutDir "package-lock.json"
        if (Test-Path $gitignore) { Remove-Item $gitignore -Force }
        if (Test-Path $packageLock) { Remove-Item $packageLock -Force }
    } else {
        Write-ColorOutput "[FAIL] Failed: $CrateName" "Red"
        # Re-run to show error output
        & wasm-pack @wasmPackArgs
        $FailedCrates++
        $FailedNames += $CrateName
    }
}

Write-Host ""
Write-ColorOutput "==================================================================" "Cyan"
Write-ColorOutput "  Build Summary" "Cyan"
Write-ColorOutput "==================================================================" "Cyan"
Write-ColorOutput "Successful: $BuiltCrates" "Green"
Write-ColorOutput "Failed: $FailedCrates" "Red"
Write-ColorOutput "Total: $TotalCrates" "Cyan"

if ($FailedNames.Count -gt 0) {
    Write-Host ""
    Write-ColorOutput "Failed crates:" "Red"
    foreach ($name in $FailedNames) {
        Write-ColorOutput "  - $name" "Red"
    }
}

# Generate index file
$IndexFile = Join-Path $PkgDir "index.ts"
$timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
$IndexContent = "// Auto-generated WASM module index`n// Generated at: $timestamp`n`n"

foreach ($CrateDir in $Crates) {
    $CrateName = $CrateDir.Name
    $CratePkgDir = Join-Path $PkgDir $CrateName

    if (Test-Path $CratePkgDir) {
        # Convert kebab-case to camelCase
        $parts = $CrateName -split '-'
        $exportName = ""
        for ($i = 0; $i -lt $parts.Count; $i++) {
            if ($i -eq 0) {
                $exportName += $parts[$i].ToLower()
            } else {
                $exportName += $parts[$i].Substring(0,1).ToUpper() + $parts[$i].Substring(1).ToLower()
            }
        }

        $wasmName = $CrateName -replace '-', '_'
        $IndexContent += "export * as $exportName from './$CrateName/samavaya_$wasmName';`n"
    }
}

Set-Content -Path $IndexFile -Value $IndexContent -Encoding UTF8

Write-Host ""
Write-ColorOutput "Generated: $IndexFile" "Green"

if ($FailedCrates -eq 0) {
    Write-Host ""
    Write-ColorOutput "Build completed successfully!" "Green"
    exit 0
} else {
    Write-Host ""
    Write-ColorOutput "Build completed with $FailedCrates failure(s)" "Yellow"
    exit 1
}
