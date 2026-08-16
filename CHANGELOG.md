# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-08-16

### 🚀 Official Initial Release — Interactive RDB Byte Explorer

Rediscope v1.0.0 provides deep, byte-level decompilation and state observability for Redis RDB snapshots through an interactive 3-pane Web UI on localhost.

### ✨ Added
- **Interactive 3-Pane Web Application**:
  - **Left Pane (File Selector)**: Multi-file snapshot explorer with version and byte metadata badges.
  - **Center Pane (Record Tree)**: Comprehensive structural hierarchy for metadata headers, AUX fields, logical DB selectors, resize hints, and key-value records.
  - **Interactive Closable Tabs**: Real-time tab switching with hover-styled close (`×`) buttons, keyboard shortcuts (`Delete` / `Backspace`), and clean empty states.
  - **Right Pane (10-Column Byte Grid)**: Granular hex byte matrix mapping every single binary byte to its corresponding record.
- **Color-Coded Byte Classification**:
  - 🟦 **Key Bytes (Cyan)**: Key names and string labels.
  - 🟩 **Value Bytes (Green)**: Encoded data structures and payload values.
  - ⬜ **Other Bytes (Grey)**: Opcodes, length encodings, headers, and CRC checksums.
- **Micro-Hover Inspection**: Instant 4-field floating tooltip displaying hex offset, hex/decimal byte value, ASCII character, and string token context.
- **Bidirectional Focus & Auto-Scrolling**: Hardware-accelerated byte highlighting and smooth scrolling on record click.
- **Full RDB Version Support (`v9` through `v15`)**:
  - Full compatibility across Redis 6.0 through Redis 8.x / unstable trunk.
  - Complete opcode decoding: `EOF`, `SELECTDB`, `EXPIRETIME_MS`, `RESIZEDB`, `AUX`, `FREQ`, `IDLE`, `MODULE_AUX`, `FUNCTION2`, `SLOT_INFO`, `KEY_META`, `HASH_TEMPLATE`.
  - Full data type coverage: Strings, Lists (quicklist2), Sets (listpack), Sorted Sets, Hashes (listpacks and schema templates), Streams (`STREAM_LISTPACKS_5`), Arrays, HyperLogLog, Bitmaps, and Modules.
- **Flexible Port Configuration**:
  - Support for `-p <port>`, `-p=<port>`, `--port <port>`, `--port=<port>`, and `PORT=<port>` environment variable.
- **File Pattern Resolver**:
  - Load snapshots via exact paths, glob patterns (`*.rdb`), or regex expressions (`redis-[67]\..*\.rdb`).
- **Headless Mode**:
  - Generate static single-file HTML reports via `--no-serve` and `--out <dir>`.
- **Precompiled Multi-Platform Binaries**:
  - Ships with prebuilt native binaries for `darwin-arm64`, `darwin-x64`, `linux-arm64`, `linux-x64`, and `win32-x64`.
