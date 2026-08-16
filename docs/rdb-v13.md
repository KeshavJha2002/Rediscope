# RDB Version 13 — Redis 8.6

**Magic:** `REDIS0013`
**Verified fixture:** `redis-8.6.5-bulk.rdb` (26,976 bytes), `redis-8.6.5-redis-tests-complex.rdb` (64,696,888 bytes)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

---

## What's New in v13 (vs v12)

### New Type Bytes

| Type | Hex | Name | Description | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 27 | `0x1B` | STREAM_LISTPACKS_5 | Streams with XNACK support (NACKed entry tracking) | [`rdb.h:82`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L82) | [`rdb.c:3873`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3873) |
| 28 | `0x1C` | ARRAY | New first-class array data type with tagged elements | [`rdb.h:83`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L83) | [`rdb.c:4393`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4393) |

### New Opcodes

None. Same opcode set as v12 (`0xF3`–`0xFF`).

## Value Encoding — New Types

### STREAM_LISTPACKS_5 (type 27)

Source: [`src/rdb.c:3873`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3873) (type dispatch, shared stream loader)

Extends STREAM_LISTPACKS_4 (type 26) with XNACK support — consumer groups can now track NACKed (negatively acknowledged) entries.

### ARRAY (type 28)

Source: [`src/rdb.c:4393`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4393) — `rdbtype == RDB_TYPE_ARRAY`

```
[element_count: length-encoded]
  for each element:
    [tag: 1 byte]
    [payload: varies by tag]
```

Tag definitions from [`src/rdb.h`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h):

| Tag | Value | Payload | Size |
|-----|-------|---------|------|
| `0` | `arrayTagSDS` | String-encoded SDS value | variable |
| `1` | `arrayTagInt` | int64, little-endian | 8 bytes, FIXED |
| `2` | `arrayTagFloat` | IEEE 754 double | 8 bytes, FIXED |
| `3` | `arrayTagSmallStr` | `[1-byte length] [inline bytes]` | variable |

## All Types Available in v13

Types 0–7, 9–28.

## Rediscope v1:beta Handling

| Feature | Status | Rediscope Source |
|---------|--------|------------------|
| Type 27 registered | ✅ | [`types.go:64`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L64) |
| Type 28 registered | ✅ | [`types.go:65`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L65) |
| STREAM_LISTPACKS_5 skipping | ✅ | [`reader.go:340`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/reader.go#L340) `skipStream()` |
| ARRAY skipping | ✅ | [`reader.go:492`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/reader.go#L492) `skipArray()` |
| GeneralType mapping | ✅ | [`types.go:143`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L143) — type 28 → `"array"` |
| TypeColor mapping | ✅ | [`types.go:172`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L172) — type 28 → `var(--array)` |

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-8.6.5-bulk.rdb` | 468 | 27 KB | ✅ |
| `redis-8.6.5-redis-tests-complex.rdb` | 53,939 | 64 MB | ❌ |
