# rediscope

Redis inspection toolkit for persistence files and, later, live Redis runtime views.

`v1.0.0-beta.0` ships the RDB vertical: parse one or more `.rdb` files and generate a static byte-level HTML viewer.

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
rediscope rdb dump.rdb
rediscope rdb dump-1.rdb dump-2.rdb --out .rediscope/rdb-viewer
```

The command parses each file independently, then writes:

```text
.rediscope/rdb-viewer/index.html
```

Open that file in a browser. No server is required.

## Viewer Flow

The generated page has three panes:

- Left: RDB file list. Multiple files can be chained in one viewer.
- Center: parsed RDB structure, including header, AUX fields, database markers, expiry metadata, keys, values, EOF, and checksum.
- Right: byte view. Key bytes are sky blue, value bytes are light green, and all other bytes are grey.

Clicking a parsed RDB section highlights that exact byte range. Clicking inside the byte view highlights only the contiguous neighboring run with the same byte class.

## Parser Scope

The parser follows Redis trunk RDB loader structure for the phase-1 byte ranges:

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

Module payloads remain semantically opaque without the module that produced them. rediscope can preserve byte ranges for self-describing module payloads, but it does not claim to decode module-private meaning.

## Development

```bash
npm run build
npm test
npm pack --dry-run
```

The npm package includes the Node launcher and the built Go binary. `prepack` runs the Go build so `npm pack` and `npm publish` do not depend on a stale local `dist/` directory.

## Test Data Policy

Keep committed RDB fixtures small and intentional. Good committed fixtures belong under `testdata/rdb/` and should cover parser contracts such as old RDB versions, native types, expiry, stream metadata, checksum handling, and module-opaque payloads.

Keep bulk generated corpora internal or in external artifact storage. Track their Redis version, commit/tag, generation commands, config, expected key count/types, and file hash in a manifest instead of committing every generated binary.
