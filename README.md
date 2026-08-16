# rediscope

Redis inspection toolkit for persistence files and live Redis runtime views.

`v1.0.0-beta.0` ships the RDB vertical: parse one or more `.rdb` files (or patterns) and serve an interactive byte-level HTML viewer on localhost.

## Install

For local development:

```bash
npm install
npm run build
node bin/rediscope.js rdb ../lab_artifacts/redis_persistence/native-types.rdb
```

After publishing:

```bash
npx rediscope@beta rdb dump.rdb
```

## Commands

### RDB viewer

```bash
# Single file
rediscope rdb dump.rdb

# Multiple files or glob patterns
rediscope rdb test/rdb/*bulk.rdb

# Regex file patterns
rediscope rdb 'test/rdb/redis-[67]\..*bulk\.rdb'

# Custom output directory and port
rediscope rdb *rdb --port 8080 --out .rediscope/rdb-viewer

# Headless / static build (no server)
rediscope rdb dump.rdb --no-serve
```

The command parses each matched file independently, writes `.rediscope/rdb-viewer/index.html`, and spins up a live localhost web server (opening your default browser automatically).

## Viewer Flow

The generated viewer has three panes matching the lab environment:

- **Left (Files)**: RDB file list. Switch between open files or inspect multi-file runs with active tabs.
- **Center (Records & Structure)**: Grouped RDB structure:
  - **File metadata**: Signature header, AUX fields (`redis-ver`, `redis-bits`, `ctime`, `used-mem`, `aof-base`), logical DB selectors, resize hints, slot info, idle/frequency metadata, and templates.
  - **Key value pairs**: Parsed keys with data types, encodings (`listpack`, `quicklist`, `embstr`, `sliced-array`, `stream`, etc.), sizes, and formatted values.
  - **Trailer**: EOF opcode and CRC64 checksum.
- **Right (Byte view)**: 10-column interactive byte grid with color classification (cyan for key bytes, green for value bytes, neutral grey for headers/types/trailer), interactive byte hover tooltips (hex, decimal, ASCII, string segment), record-to-byte range focus with dimming, and contiguous byte run selection.

## Parser Scope

The parser follows Redis trunk RDB loader structure for phase-1 byte ranges:

- RDB header and version
- AUX fields
- SELECTDB, RESIZEDB, SLOT_INFO
- EXPIRETIME and EXPIRETIME_MS
- IDLE and FREQ object metadata
- Redis native value encodings, including listpack/quicklist variants
- stream listpacks through `RDB_TYPE_STREAM_LISTPACKS_5`
- stream consumer groups, PEL metadata, NACK zone, and IDMP metadata
- array values
- hash metadata and hash template records
- module AUX and module values when they use Redis' self-describing module opcodes
- EOF and checksum

## Development

```bash
npm run build
npm test
npm pack --dry-run
```
