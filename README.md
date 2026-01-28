# Fast-Hasher (fhash)

[![Release](https://github.com/Virace/fast-hasher/actions/workflows/release.yml/badge.svg)](https://github.com/Virace/fast-hasher/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/Virace/fast-hasher)](https://github.com/Virace/fast-hasher/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Virace/fast-hasher)](https://go.dev/)
[![License](https://img.shields.io/github/license/Virace/fast-hasher)](LICENSE)

高性能批量文件校验 CLI 工具，支持多种哈希算法、并发处理和灵活的筛选器。

## 特性

- **多算法支持**: MD5, SHA1, SHA256, SHA512, CRC32, Blake3, XXH3, XXH128, QuickXor
- **高性能并发**: 自动利用多核 CPU 并行处理
- **灵活筛选**: 按文件大小、扩展名、glob 模式过滤
- **多种输出**: 文本格式（兼容 md5sum）、JSON Lines（便于程序解析）
- **易于集成**: 专为 Python 等语言调用设计的机器可读模式

## 安装

### 下载预编译版本

从 [Releases](https://github.com/Virace/fast-hasher/releases) 下载适合您平台的二进制文件：

- `fhash-windows-amd64.exe` - Windows x64
- `fhash-linux-amd64` - Linux x64

### 从源码编译

**要求**: Go 1.21+

```bash
# 直接安装
go install github.com/Virace/fast-hasher/cmd/fhash@latest

# 或克隆后编译
git clone https://github.com/Virace/fast-hasher.git
cd fast-hasher
go build -o fhash ./cmd/fhash
```

**其他平台编译** (macOS, ARM64 等):

```bash
# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o fhash-darwin-amd64 ./cmd/fhash

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o fhash-darwin-arm64 ./cmd/fhash

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o fhash-linux-arm64 ./cmd/fhash

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -o fhash-windows-arm64.exe ./cmd/fhash
```

**使用构建脚本** (Windows PowerShell):

```powershell
# 编译当前平台
.\scripts\build.ps1

# 编译所有发布平台 (Windows/Linux amd64)
.\scripts\build.ps1 -All

# 启用 UPX 压缩 (需要安装 UPX)
.\scripts\build.ps1 -Upx
```

编译产物输出到 `dist/` 目录。

## 使用方法

### 基础用法

```bash
# 查看支持的算法
fhash --list

# 单文件哈希
fhash -a sha256 file.txt

# 多算法
fhash -a md5,sha256 file.txt

# 目录递归扫描
fhash -a sha256 ./dist
```

### 输出格式

**文本格式**（默认，兼容 md5sum/sha256sum）:
```
3ac02015b07182e438dce6ae126270ed  README.md
```

**多算法文本格式**:
```
md5:3ac02015b07182e438dce6ae126270ed  README.md
sha256:c44e50aae1f756f9a288f3b93d00f9f1...  README.md
```

**JSON Lines 格式** (`-j` 或 `--json`):
```json
{"path":"README.md","size":13,"sha256":"c44e50aae..."}
```

### 程序集成模式

使用 `-m` (machine) 模式可禁用进度输出，配合 `-j` (JSON) 便于其他程序解析：

```bash
fhash -a sha256 -m -j ./dist
```

**Python 调用示例**:
```python
import subprocess
import json

result = subprocess.run(
    ['fhash', '-a', 'sha256', '-m', '-j', './dist'],
    capture_output=True, text=True
)

for line in result.stdout.splitlines():
    data = json.loads(line)
    print(f"{data['path']}: {data['sha256']}")
```

### 筛选器

```bash
# 跳过大于 100MB 的文件
fhash -a sha256 --max-size 100MB ./

# 只处理特定扩展名
fhash -a sha256 -I .exe,.dll ./

# 排除特定扩展名
fhash -a sha256 -E .log,.tmp ./

# 使用 glob 模式
fhash -a sha256 -i "*.txt" -e "test_*" ./

# 组合使用
fhash -a sha256 --max-size 50MB -E .log,.tmp -e "node_modules/*" ./project
```

### 从文件列表读取

```bash
# 从文件读取路径列表
fhash -a sha256 -f filelist.txt

# 从 stdin 读取
cat files.txt | fhash -a sha256 --from-stdin -m -j
```

## 命令行参数

| 参数 | 短 | 说明 | 默认值 |
|------|-----|------|--------|
| `--algo` | `-a` | 哈希算法，逗号分隔（**必需**） | - |
| `--recursive` | `-r` | 递归扫描目录 | `true` |
| `--machine` | `-m` | 机器可读模式（无进度） | `false` |
| `--json` | `-j` | JSON Lines 输出 | `false` |
| `--absolute` | | 强制输出绝对路径 | `false` |
| `--on-error` | | 错误处理：`skip` 或 `fail` | `skip` |
| `--from-file` | `-f` | 从文件读取路径列表 | - |
| `--from-stdin` | | 从 stdin 读取路径列表 | `false` |
| `--max-size` | | 跳过超过此大小的文件 | - |
| `--min-size` | | 跳过小于此大小的文件 | - |
| `--include-ext` | `-I` | 只处理这些扩展名 | - |
| `--exclude-ext` | `-E` | 排除这些扩展名 | - |
| `--include` | `-i` | 包含 glob 模式 | - |
| `--exclude` | `-e` | 排除 glob 模式 | - |
| `--workers` | `-w` | 并发数 | CPU 核心数 |
| `--list` | `-l` | 列出支持的算法 | - |
| `--version` | `-v` | 显示版本 | - |

## 支持的算法

| 算法 | 说明 | 输出长度 |
|------|------|----------|
| `md5` | MD5 | 32 hex |
| `sha1` | SHA-1 | 40 hex |
| `sha256` | SHA-256 | 64 hex |
| `sha512` | SHA-512 | 128 hex |
| `crc32` | CRC32 (IEEE) | 8 hex |
| `blake3` | BLAKE3 | 64 hex |
| `xxh3` | XXHash3 64-bit | 16 hex |
| `xxh128` | XXHash3 128-bit | 32 hex |
| `quickxor` | QuickXorHash (OneDrive) | Base64 |

## 错误处理

- `--on-error skip`（默认）：跳过无法读取的文件，在 stderr 或 JSON 输出中记录错误
- `--on-error fail`：遇到第一个错误立即退出

**错误 JSON 格式**:
```json
{"path":"unreadable.bin","error":"permission denied"}
```

## 许可证

MIT License