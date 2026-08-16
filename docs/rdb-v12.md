# RDB Version 12 — Redis 7.4 LTS / 8.0 / 8.2 / 8.4

**Magic:** `REDIS0012`
**Verified fixtures:**
- `redis-7.4.10-bulk.rdb` (27,691 bytes), `redis-8.0.6-bulk.rdb` (26,969 bytes), `redis-8.2.8-bulk.rdb` (26,969 bytes), `redis-8.4.5-bulk.rdb` (26,969 bytes)
- Plus corresponding complex fixtures (~64 MB each)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

> **Note:** Redis 8.0 through 8.4 all produce RDB v12 — no new type bytes or opcodes were introduced in these Redis releases.

---

## What's New in v12 (vs v11)

### New Type Bytes

| Type | Hex | Name | Description | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 22 | `0x16` | HASH_METADATA_PRE_GA | Hash with per-field TTL (7.4 RC, no minExpire prefix) | [`rdb.h:77`](src/rdb.h#L77) | [`rdb.c:3343`](src/rdb.c#L3343) |
| 23 | `0x17` | HASH_LISTPACK_EX_PRE_GA | Hash LP with per-field TTL (7.4 RC) | [`rdb.h:78`](src/rdb.h#L78) | [`rdb.c:3822`](src/rdb.c#L3822) |
| 24 | `0x18` | HASH_METADATA | Hash with per-field TTL (GA, includes minExpire prefix) | [`rdb.h:79`](src/rdb.h#L79) | [`rdb.c:3356`](src/rdb.c#L3356) |
| 25 | `0x19` | HASH_LISTPACK_EX | Hash LP with per-field TTL (GA, includes minExpire) | [`rdb.h:80`](src/rdb.h#L80) | [`rdb.c:3602`](src/rdb.c#L3602) |
| 26 | `0x1A` | STREAM_LISTPACKS_4 | Streams with IDMP (monotonic ID progress) support | [`rdb.h:81`](src/rdb.h#L81) | [`rdb.c:3872`](src/rdb.c#L3872) |

### New Opcodes

| Opcode | Hex | Description | Source (define) | Source (handler) |
|--------|-----|-------------|-----------------|------------------|
| KEY_META | `0xF3` | Per-key metadata (module metadata classes) | [`rdb.h:102`](src/rdb.h#L102) | [`rdb.c:2348`](src/rdb.c#L2348) |
| SLOT_INFO | `0xF4` | Per-slot sizing hints (cluster mode) | [`rdb.h:103`](src/rdb.h#L103) | [`rdb.c:4767`](src/rdb.c#L4767) |

## Value Encoding — New Types

### HASH_METADATA_PRE_GA (type 22) — RC format

Source: [`src/rdb.c:3343`](src/rdb.c#L3343) — `rdbtype == RDB_TYPE_HASH_METADATA || rdbtype == RDB_TYPE_HASH_METADATA_PRE_GA`

```
[count: length-encoded]
  for each field:
    [field: string-encoded]
    [ttl_delta: uint64 (8 bytes)]
    [value: string-encoded]
```

**No** `minExpire` base timestamp prefix. Each TTL is absolute.

### HASH_METADATA (type 24) — GA format

Source: [`src/rdb.c:3356`](src/rdb.c#L3356) — `if (rdbtype == RDB_TYPE_HASH_METADATA)`

```
[minExpire: uint64 (8 bytes)]            ← base timestamp
[count: length-encoded]
  for each field:
    [field: string-encoded]
    [ttl_delta: relative to minExpire]
    [value: string-encoded]
```

The `minExpire` prefix is the minimum expiration time across all fields.

### HASH_LISTPACK_EX (type 25) — GA format

Source: [`src/rdb.c:3602`](src/rdb.c#L3602) — `if (rdbtype == RDB_TYPE_HASH_LISTPACK_EX)`

```
[minExpire: uint64 (8 bytes)]
[blob: string-encoded]   ← listpack with relative TTL fields interleaved
```

### STREAM_LISTPACKS_4 (type 26)

Source: [`src/rdb.c:3872`](src/rdb.c#L3872), IDMP entries at [`src/rdb.c:4393`](src/rdb.c#L4393)

Extends STREAM_LISTPACKS_3 with IDMP (ID Monotonic Progress) tracking.

## Rediscope v1:beta Handling

| Feature | Status | Rediscope Source |
|---------|--------|------------------|
| Types 22–26 registered | ✅ | [`types.go:59–63`](../internal/rdb/types.go#L59) |
| Hash metadata skipping | ✅ | [`reader.go:216`](../internal/rdb/reader.go#L216) `skipValue()` |
| STREAM_LISTPACKS_4 skipping | ✅ | [`reader.go:340`](../internal/rdb/reader.go#L340) `skipStream()` |
| KEY_META opcode | ✅ | [`parser.go:360–382`](../internal/rdb/parser.go#L360) |
| SLOT_INFO opcode | ✅ | [`parser.go:253–275`](../internal/rdb/parser.go#L253) |

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-7.4.10-bulk.rdb` | 468 | 27 KB | ✅ |
| `redis-8.0.6-bulk.rdb` | 468 | 27 KB | ✅ |
| `redis-8.2.8-bulk.rdb` | 468 | 27 KB | ✅ |
| `redis-8.4.5-bulk.rdb` | 468 | 27 KB | ✅ |
| Complex fixtures (×4) | ~54K each | 64 MB | ❌ |
