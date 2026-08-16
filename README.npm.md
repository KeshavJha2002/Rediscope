# Rediscope v1.0.0

> Interactive byte-level RDB binary decompiler and visual state inspector for Redis.

[![npm version](https://img.shields.io/npm/v/rediscope.svg)](https://www.npmjs.com/package/rediscope)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Rediscope** decompiles binary Redis RDB snapshots into an interactive, zero-dependency 3-pane Web UI on localhost. It bridges the gap between raw binary bytes and higher-level logical structures, mapping every single byte in the RDB stream to its corresponding record, key, value, and metadata.

---

## Quick Start (Zero Install)

Run instantly without installing anything using `npx`:

```bash
# Decompile and launch viewer for an RDB file
npx rediscope rdb <path/file_name>.rdb

# Run on a custom port
npx rediscope rdb <path/file_name>.rdb -p 8080

# Inspect all RDB files in a directory
npx rediscope rdb *.rdb
```

---

## Installation

### Global Installation
Install globally to have the `rediscope` command available anywhere:

```bash
npm install -g rediscope
```

### Local Project Installation
```bash
npm install --save-dev rediscope
```

---

## CLI Usage & Options

### Command Syntax

```bash
rediscope rdb <file-or-pattern> [more-patterns ...] [options]
```

### Options & Port Assignment

| Flag | Short | Environment Var | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--port <num>` | `-p <num>` | `PORT=<num>` | `Auto (random free)` | Assign custom localhost web server port |
| `--out <dir>` | — | — | `.rediscope/rdb-viewer` | Output directory for the generated static viewer |
| `--no-open` | — | — | `false` | Do not automatically open default web browser |
| `--no-serve` | — | — | `false` | Headless mode: parse and write `index.html` without running server |
| `--help` | `-h` | — | — | Display help information and usage patterns |
| `--version` | `-v` | — | — | Display version information (`1.0.0`) |

### Usage Examples

```bash
# 1. Inspect single RDB file
rediscope rdb <path/file_name>.rdb

# 2. Assign custom server port
rediscope rdb <path/file_name>.rdb -p 3000
rediscope rdb <path/file_name>.rdb --port 8080
PORT=5000 rediscope rdb <path/file_name>.rdb

# 3. Multiple files and glob patterns
rediscope rdb file1.rdb file2.rdb
rediscope rdb *.rdb
rediscope rdb 'snapshots/redis-*.rdb'

# 4. Regex file patterns
rediscope rdb 'redis-[67]\..*\.rdb'

# 5. Headless export (build static HTML viewer for CI/reports without starting server)
rediscope rdb <path/file_name>.rdb --out ./reports/viewer --no-serve
```

---

## What Rediscope Does

Rediscope parses the raw RDB binary stream and renders a synchronized, hardware-accelerated **3-Pane Web Application**:

1. **Left Sidebar (File Selector)**:
   - Displays all loaded RDB snapshots with metadata badges (keys count, binary size, version).
   - Instant switching between files with active tab synchronization.

2. **Center Pane (Record Tree & Structure)**:
   - **File Metadata**: Signature header, AUX fields (`redis-ver`, `redis-bits`, `ctime`, `used-mem`, `aof-base`), logical DB selectors, resize hints, slot info, and hash templates.
   - **Key-Value Records**: Parsed keys with data types, internal encodings (`listpack`, `quicklist`, `embstr`, `sliced-array`, `stream`, etc.), memory sizes, and decoded values.
   - **Interactive Tabs with Close (`×`) Buttons**: Easily close any active file tab or switch between multiple open snapshots.
   - **Trailer**: EOF opcode and 64-bit CRC checksum verification.

3. **Right Pane (Byte View Grid)**:
   - **10-Column Hex Matrix**: Every individual byte in the binary file is represented as an interactive cell.
   - **Color-Coded Classification**:
     - 🟦 **Key Bytes (Cyan)**: Bytes corresponding to key names and string labels.
     - 🟩 **Value Bytes (Green)**: Bytes corresponding to encoded data values.
     - ⬜ **Other Bytes (Neutral Grey)**: Opcodes, length encodings, headers, and checksums.
   - **4-Field Hover Tooltip**: Real-time inspection showing **Hex offset**, **Hex/Decimal byte value**, **ASCII printable character**, and **String segment descriptor**.
   - **Bidirectional Focus & Auto-Scroll**: Clicking any record in the center pane automatically highlights its byte range with hardware-accelerated dimming and smooth-scrolls the byte grid to the target range.
   - **Continuous Byte Run Selection**: Clicking any byte in the grid highlights its entire logical token run.

---

## Compatibility Matrix

Rediscope supports all major Redis versions and RDB formats from legacy Redis 6.0 up to the latest Redis trunk:

### 1. Redis & RDB Versions

| RDB Version | Redis Versions | Magic Signature | Status |
| :--- | :--- | :--- | :--- |
| **RDB v9** | Redis 6.0, 6.2 LTS | `REDIS0009` | ✅ Fully Supported |
| **RDB v10** | Redis 7.0 LTS | `REDIS0010` | ✅ Fully Supported |
| **RDB v11** | Redis 7.2 LTS | `REDIS0011` | ✅ Fully Supported |
| **RDB v12** | Redis 7.4 LTS, 8.0, 8.2, 8.4 | `REDIS0012` | ✅ Fully Supported |
| **RDB v13** | Redis 8.6 | `REDIS0013` | ✅ Fully Supported |
| **RDB v14** | Redis 8.8 | `REDIS0014` | ✅ Fully Supported |
| **RDB v15** | Redis 8.9+ (trunk) | `REDIS0015` | ✅ Fully Supported |

### 2. Opcodes Supported

| Opcode | Hex | Purpose | Support Status |
| :--- | :--- | :--- | :--- |
| `EOF` | `0xFF` | End of RDB stream delimiter | ✅ Verified |
| `SELECTDB` | `0xFE` | Logical database selector | ✅ Verified |
| `EXPIRETIME` | `0xFD` | Expiry timestamp in seconds | ✅ Verified |
| `EXPIRETIME_MS` | `0xFC` | Expiry timestamp in milliseconds | ✅ Verified |
| `RESIZEDB` | `0xFB` | Hash table size and expire size hints | ✅ Verified |
| `AUX` | `0xFA` | Arbitrary metadata key-value pairs | ✅ Verified |
| `FREQ` | `0xF9` | LFU frequency counter byte | ✅ Verified |
| `IDLE` | `0xF8` | LRU idle time value | ✅ Verified |
| `MODULE_AUX` | `0xF7` | Module auxiliary serialized data | ✅ Verified |
| `FUNCTION2` | `0xF5` | Function library definitions | ✅ Verified |
| `SLOT_INFO` | `0xF4` | Cluster slot assignment metadata | ✅ Verified |
| `KEY_META` | `0xF3` | Extended key metadata descriptors | ✅ Verified |
| `HASH_TEMPLATE` | `0xF2` | Shared schema template for hash records | ✅ Verified |

### 3. Data Types & Encodings Supported

- **Strings**: Raw strings, embstr, integer encodings (`8-bit`, `16-bit`, `32-bit`), LZF compressed strings.
- **Lists**: Linked list, ziplist, quicklist (`LIST_QUICKLIST`), quicklist2 (`LIST_QUICKLIST_2`).
- **Sets**: Plain set, intset (`SET_INTSET`), listpack (`SET_LISTPACK`).
- **Sorted Sets (ZSets)**: String score, binary double (`ZSET_2`), ziplist (`ZSET_ZIPLIST`), listpack (`ZSET_LISTPACK`).
- **Hashes**: Plain hash, zipmap, ziplist, listpack (`HASH_LISTPACK`), hash metadata (`HASH_METADATA`), hash listpack with TTLs (`HASH_LISTPACK_EX`), template hashes (`HASH_TMPL_LP`, `HASH_TMPL_ARRAY`).
- **Streams**: Stream listpacks through `STREAM_LISTPACKS_5` with Consumer Groups, PEL, NACK, and IDMP tracking.
- **Arrays**: Native tagged array data type (`ARRAY`).
- **HyperLogLog & Bitmaps**: Probabilistic structures and raw byte representations.
- **Modules**: Self-describing module serialization format (`MODULE_2`).

---

## Feature Index by Version

### `v1.0.0` (Official Release)
- **Interactive Web UI**: Zero-dependency standalone HTML/CSS/JS viewer generated on-the-fly.
- **Multi-File Support**: Parse and switch between multiple RDB files simultaneously with closable tabs (`×`).
- **Byte-to-Record Visual Mapping**: Bidirectional sync between high-level records and raw binary byte grid.
- **Micro-Hover Inspection**: 4-field instant tooltip (Hex offset, byte value, ASCII char, string segment).
- **Flexible Port Selection**: Assign ports via `-p`, `--port`, or `PORT` environment variable.
- **Pattern Matching**: Load files via exact paths, glob patterns (`*.rdb`), or regex expressions.
- **Precompiled Multi-Platform Binaries**: Ships with prebuilt native binaries for `darwin-arm64`, `darwin-x64`, `linux-arm64`, `linux-x64`, and `win32-x64`.
- **Full RDB Version Coverage**: Complete decoding parity across RDB formats `v9` through `v15`.

---

## License

MIT License
