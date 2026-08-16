# RDB Version 10 — Redis 7.0 LTS

**Magic:** `REDIS0010`
**Verified fixture:** `redis-7.0.15-bulk.rdb` (27,691 bytes), `redis-7.0.15-redis-tests-complex.rdb` (64,667,999 bytes)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

---

## What's New in v10 (vs v9)

### New Type Bytes

| Type | Hex | Name | Description | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 18 | `0x12` | LIST_QUICKLIST_2 | Quicklist with per-node container format flag | [`rdb.h:73`](src/rdb.h#L73) | [`rdb.c:3523`](src/rdb.c#L3523) |
| 19 | `0x13` | STREAM_LISTPACKS_2 | Streams with entries_added, first/max deleted entry IDs | [`rdb.h:74`](src/rdb.h#L74) | [`rdb.c:3870`](src/rdb.c#L3870) |
| 20 | `0x14` | SET_LISTPACK | Sets encoded as listpack (for small string sets) | [`rdb.h:75`](src/rdb.h#L75) | [`rdb.c:3727`](src/rdb.c#L3727) |

### New Opcodes

| Opcode | Hex | Description | Source (define) | Source (handler) |
|--------|-----|-------------|-----------------|------------------|
| FUNCTION2 | `0xF5` | Function library data (GA format) | [`rdb.h:104`](src/rdb.h#L104) | [`rdb.c:4903`](src/rdb.c#L4903) |
| FUNCTION_PRE_GA | `0xF6` | Function library data (RC1/RC2 only) | [`rdb.h:105`](src/rdb.h#L105) | [`rdb.c:4900`](src/rdb.c#L4900) |

## Value Encoding — New Types

### LIST_QUICKLIST_2 (type 18)

Source: [`src/rdb.c:3513–3586`](src/rdb.c#L3513)

```
[node_count: length-encoded]
  for each node:
    [container_format: length-encoded]   ← 1=ziplist packed, 2=listpack
    [blob: string-encoded]
```

Differs from v9's `LIST_QUICKLIST` (type 14) which had **no** container format field — all nodes were implicitly ziplist. Source for the container check: [`src/rdb.c:3523`](src/rdb.c#L3523) `if (rdbtype == RDB_TYPE_LIST_QUICKLIST_2)`.

### STREAM_LISTPACKS_2 (type 19)

Source: [`src/rdb.c:3869–3873`](src/rdb.c#L3869) (type dispatch)

Extends v9 STREAM_LISTPACKS (type 15) with additional metadata after the radix tree:

```
[...same radix tree prefix as type 15...]
[stream_length: length-encoded]
[entries_added: length-encoded]           ← NEW in v2
[first_entry_id: 16 bytes (ms + seq)]    ← NEW in v2
[max_deleted_entry_id: 16 bytes]          ← NEW in v2
[consumer_groups: same as v1 format]
```

### SET_LISTPACK (type 20)

Source: [`src/rdb.c:3727`](src/rdb.c#L3727), falls into the blob-load branch at [`src/rdb.c:3588–3597`](src/rdb.c#L3588)

```
[blob: string-encoded]   ← single listpack binary
```

## Rediscope v1:beta Handling

| Feature | Status | Rediscope Source |
|---------|--------|------------------|
| Types 18, 19, 20 registered | ✅ | [`types.go:55–57`](../internal/rdb/types.go#L55) |
| QUICKLIST_2 skipping | ✅ | [`reader.go:216`](../internal/rdb/reader.go#L216) `skipValue()` |
| STREAM_LISTPACKS_2 skipping | ✅ | [`reader.go:340`](../internal/rdb/reader.go#L340) `skipStream()` |
| SET_LISTPACK skipping | ✅ | Handled as single string blob skip |
| FUNCTION2 opcode | ✅ | [`parser.go:300`](../internal/rdb/parser.go#L300) |
| FUNCTION_PRE_GA opcode | 🔴 ABORT | [`parser.go:318–325`](../internal/rdb/parser.go#L318) — returns error |

### ⚠️ FUNCTION_PRE_GA Issue

Opcode `0xF6` causes the parser to return: `"pre-GA function opcode 0xF6 is not supported by Redis trunk"` ([`parser.go:323`](../internal/rdb/parser.go#L323)). GA releases of Redis 7.0 do NOT emit this opcode — only RC1/RC2.

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-7.0.15-bulk.rdb` | 468 | 27 KB | ✅ Parsed in CI |
| `redis-7.0.15-redis-tests-complex.rdb` | 53,844 | 64 MB | ❌ Not in automated tests |
