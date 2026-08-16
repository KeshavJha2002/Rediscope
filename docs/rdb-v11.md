# RDB Version 11 — Redis 7.2 LTS

**Magic:** `REDIS0011`
**Verified fixture:** `redis-7.2.15-bulk.rdb` (27,691 bytes), `redis-7.2.15-redis-tests-complex.rdb` (64,659,135 bytes)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

---

## What's New in v11 (vs v10)

### New Type Bytes

| Type | Hex | Name | Description | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 21 | `0x15` | STREAM_LISTPACKS_3 | Streams with per-consumer active_time and per-group entries_read | [`rdb.h:76`](src/rdb.h#L76) | [`rdb.c:3871`](src/rdb.c#L3871) |

### New Opcodes

None. Same opcode set as v10 (`0xF5`–`0xFF`).

## Value Encoding — STREAM_LISTPACKS_3 (type 21)

Source: [`src/rdb.c:3869–3873`](src/rdb.c#L3869) (type dispatch, falls into shared stream loader)

Extends `STREAM_LISTPACKS_2` (type 19) with per-consumer tracking in consumer groups:

```
[...same prefix as type 19...]
[consumer_groups_count: length-encoded]
  for each group:
    [name: string-encoded]
    [last_id: 16 bytes (ms + seq)]
    [entries_read: length-encoded]          ← NEW in v3
    [pel_count: length-encoded]
    [pending entries...]
    [consumers_count: length-encoded]
      for each consumer:
        [name: string-encoded]
        [seen_time: 8 bytes]
        [active_time: 8 bytes]              ← NEW in v3
        [consumer_pel_count: length-encoded]
        [consumer pending entries...]
```

The key additions are `active_time` per consumer (8-byte int64 millisecond timestamp) and `entries_read` per consumer group.

## All Types Available in v11

Types 0–7, 9–21 (everything from v10 plus type 21).

## Rediscope v1:beta Handling

| Feature | Status | Rediscope Source |
|---------|--------|------------------|
| Type 21 registered | ✅ | [`types.go:58`](../internal/rdb/types.go#L58) |
| STREAM_LISTPACKS_3 skipping | ✅ | [`reader.go:340`](../internal/rdb/reader.go#L340) `skipStream()` — [`reader.go:394`](../internal/rdb/reader.go#L394) `skipStreamConsumerGroup()` |
| GeneralType mapping | ✅ | [`types.go:141`](../internal/rdb/types.go#L141) — type 21 → `"stream"` |

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-7.2.15-bulk.rdb` | 468 | 27 KB | ✅ Parsed in CI |
| `redis-7.2.15-redis-tests-complex.rdb` | 54,166 | 64 MB | ❌ Not in automated tests |
