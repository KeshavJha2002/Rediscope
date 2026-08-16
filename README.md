# Rediscope

> Byte-level state observability and binary decompiler toolkit for Redis.

[![CI Matrix](https://github.com/rediscope/rediscope/actions/workflows/ci.yml/badge.svg)](https://github.com/rediscope/rediscope/actions)
[![npm version](https://img.shields.io/npm/v/rediscope.svg)](https://www.npmjs.com/package/rediscope)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Rediscope** provides deep, byte-level observability into Redis persistence formats and runtime memory state. It is designed to make low-level binary state transitions, memory encodings, and opcode layouts human-understandable.

---

## Quickstart

### Running via npx (Zero Install)

```bash
# Decompile and inspect an RDB snapshot
npx rediscope rdb <path/file_name>.rdb

# Run on a custom port
npx rediscope rdb <path/file_name>.rdb -p 3000

# Inspect multiple snapshots
npx rediscope rdb *.rdb
```

### Running via Global npm Install

```bash
npm install -g rediscope
rediscope rdb <path/file_name>.rdb
```

---

## Local Development & Source Setup

### Prerequisites
- **Go**: 1.22+ installed
- **Node.js**: 20+ (Node 24 recommended) installed

### Setup Steps

```bash
# 1. Clone the repository
git clone https://github.com/rediscope/rediscope.git
cd rediscope

# 2. Install dependencies
npm install

# 3. Build the native binary
npm run build
# (Runs: go build -buildvcs=false -o dist/rediscope ./cmd/rediscope)

# 4. Run automated test suite
npm test
# (Runs: go test ./...)

# 5. Run local binary directly
./dist/rediscope rdb test/testdata/native-types.rdb

# 6. Verify npm package bundle locally
npm run pack:local
```

---

## Repository Structure & Architecture

```
rediscope/
├── bin/
│   └── rediscope.js            # Node.js launcher & prebuild platform dispatcher
├── cmd/
│   └── rediscope/
│       └── main.go             # Go CLI entrypoint
├── internal/
│   ├── cli/
│   │   ├── app.go              # CLI command router, port resolver, pattern matching
│   │   └── server.go           # Embedded zero-dependency HTTP static server
│   ├── rdb/
│   │   ├── parser.go           # RDB binary parser & state machine
│   │   ├── reader.go           # Length-encoded integer & string binary reader
│   │   ├── types.go            # Redis type opcode constants (0x00-0x21)
│   │   ├── crc64.go            # 64-bit CRC checksum validation
│   │   └── models.go           # RDB file metadata and record models
│   └── viewer/
│       ├── viewer.go           # HTML generator & payload serializer
│       └── template.go         # Embedded single-file Web UI (HTML/CSS/JS)
├── prebuilds/                  # Native multi-platform cross-compiled binaries
│   ├── darwin-arm64/rediscope  # macOS Apple Silicon
│   ├── darwin-x64/rediscope    # macOS Intel
│   ├── linux-arm64/rediscope   # Linux AArch64
│   ├── linux-x64/rediscope     # Linux x86_64
│   └── win32-x64/rediscope.exe # Windows x64
├── test/
│   ├── testdata/               # Self-contained RDB binary test fixtures
│   └── rdb_parser_test.go      # End-to-end parser & CLI unit test suite
├── docs/                       # Specifications, invariants, and version docs
├── .github/
│   └── workflows/
│       └── ci.yml              # Multi-OS & Multi-Node/Go CI matrix workflow
├── package.json                # npm package configuration & build scripts
├── README.md                   # GitHub repository documentation
└── README.npm.md               # npm registry documentation (swapped on publish)
```

---

## Feature Index by Version

### 📦 `v1.0.0` (Official Stable Release)
- **RDB Byte Decompiler**: High-performance parsing of binary `.rdb` files across all formats from `v9` to `v15` (Redis 6.0 through 8.x).
- **Interactive 3-Pane Web UI**:
  - **File Selector**: Multi-file explorer with live closable tabs (`×`) and active file state management.
  - **Record Tree**: Full structural breakdown of metadata, logical DBs, AUX fields, hash templates, and key-value records.
  - **10-Column Byte Matrix**: Color-coded key bytes (cyan), value bytes (green), and metadata (grey).
- **Deep Hover Inspection**: 4-field instant tooltip (Hex offset, byte value, ASCII char, string segment).
- **Record-to-Byte Synchronization**: Hardware-accelerated byte highlighting and auto-scrolling when clicking any record.
- **Port Flexibility**: Configurable port via `-p <port>`, `--port <port>`, or `PORT` environment variable.
- **Pattern Matching**: Load files via exact paths, glob patterns (`*.rdb`), or regex expressions.
- **Cross-Platform Prebuilds**: Ships ready-to-run on Linux, macOS, and Windows.

### 🔮 `v2.0.0` (Upcoming Roadmap / In-Development)
- **Live Redis Polling Engine**: Direct socket connection to live Redis instances without triggering `SAVE` / `BGSAVE`.
- **Tiered Namespace Navigation**: TUI prefix-tree namespace explorer (`users:` $\rightarrow$ `profile:` $\rightarrow$ leaf keys) with 4-arrow drilldown.
- **Physical Memory Pointer Graph**: Trace `dbDict` hash table buckets and C struct pointers (`dictEntry` $\rightarrow$ `sds` $\rightarrow$ `robj` $\rightarrow$ `listpack`).
- **Live Command Stream**: Real-time `MONITOR` protocol listener with microsecond latency metrics.
- **Snapshot Diff Engine**: Temporal mutation tracking between snapshots (T0 ↔ T1 memory deltas and type shifts).
- **Cross-View Snap**: Synchronized side-by-side comparison of on-disk serialized bytes vs in-memory physical structures.

---

## Cross-Platform & Multi-Environment Testing

Rediscope is tested across operating systems and runtimes via GitHub Actions:
- **Operating Systems**: Ubuntu Linux, macOS, Windows
- **Node.js**: 18.x, 20.x, 22.x
- **Go**: 1.22, 1.23

To run multi-platform test suites locally:

```bash
# Unit & integration tests
npm test

# Dry-run npm packaging
npm pack --dry-run
```

---

## License

MIT License
