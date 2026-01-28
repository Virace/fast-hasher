# 构建脚本 - 编译 fhash 到 dist 目录
# 用法: .\scripts\build.ps1 [-All] [-Upx]

param(
    [switch]$All,          # 编译所有平台
    [switch]$Upx           # 使用 UPX 压缩 (仅 Windows)
)

$ErrorActionPreference = "Stop"

# 获取脚本所在目录，然后定位到项目根目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

# 切换到项目根目录
Push-Location $ProjectRoot

try {
    # 获取 Git commit hash
    $Commit = git rev-parse --short HEAD 2>$null
    if (-not $Commit) { $Commit = "unknown" }

    # 获取编译时间 (UTC)
    $BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss UTC")

    # 版本号: dev
    $Version = "dev"

    # 构建参数
    $LdFlags = "-s -w -X main.Version=$Version -X main.Commit=$Commit -X `"main.BuildTime=$BuildTime`""
    $OutputDir = Join-Path $ProjectRoot "dist"

    # 创建输出目录
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }

    function Build-Platform {
        param(
            [string]$GOOS,
            [string]$GOARCH,
            [string]$Suffix = ""
        )
        
        $OutputName = "fhash-$GOOS-$GOARCH$Suffix"
        $OutputPath = Join-Path $OutputDir $OutputName
        
        Write-Host "Building $OutputName..." -ForegroundColor Cyan
        
        $env:GOOS = $GOOS
        $env:GOARCH = $GOARCH
        $env:CGO_ENABLED = "0"
        
        go build -ldflags $LdFlags -o $OutputPath ./cmd/fhash
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  -> $OutputPath" -ForegroundColor Green
            
            # UPX 压缩 (仅 Windows 二进制)
            if ($Upx -and $GOOS -eq "windows") {
                $upxPath = Get-Command upx -ErrorAction SilentlyContinue
                if ($upxPath) {
                    Write-Host "  Compressing with UPX..." -ForegroundColor Yellow
                    & upx --best --lzma $OutputPath 2>&1 | Out-Null
                    if ($LASTEXITCODE -eq 0) {
                        $size = (Get-Item $OutputPath).Length
                        Write-Host "  -> Compressed: $([math]::Round($size/1KB)) KB" -ForegroundColor Green
                    }
                } else {
                    Write-Host "  UPX not found, skipping compression" -ForegroundColor Yellow
                }
            }
        } else {
            Write-Host "  -> Failed!" -ForegroundColor Red
            exit 1
        }
    }

    Write-Host "Build Info:" -ForegroundColor Yellow
    Write-Host "  Version: $Version" 
    Write-Host "  Commit: $Commit"
    Write-Host "  BuildTime: $BuildTime"
    Write-Host ""

    if ($All) {
        # Windows
        Build-Platform -GOOS "windows" -GOARCH "amd64" -Suffix ".exe"
        
        # Linux
        Build-Platform -GOOS "linux" -GOARCH "amd64"
        
        Write-Host ""
        Write-Host "All builds completed!" -ForegroundColor Green
    } else {
        # 检测当前平台
        $CurrentOS = if ($IsWindows -or $env:OS -eq "Windows_NT") { "windows" } elseif ($IsMacOS) { "darwin" } else { "linux" }
        $CurrentArch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
        $Suffix = if ($CurrentOS -eq "windows") { ".exe" } else { "" }
        
        Build-Platform -GOOS $CurrentOS -GOARCH $CurrentArch -Suffix $Suffix
        
        Write-Host ""
        Write-Host "Build completed!" -ForegroundColor Green
    }

    Write-Host ""
    Write-Host "Output directory: $OutputDir" -ForegroundColor Cyan
    Get-ChildItem $OutputDir | Format-Table Name, Length, LastWriteTime
}
finally {
    Pop-Location
}
