# RDB Version 9 — Redis 6.0 / 6.2 LTS

**Magic:** `REDIS0009`
**Verified fixture:** `redis-6.2.23-bulk.rdb` (27,717 bytes), `redis-6.2.23-redis-tests-complex.rdb` (64,705,666 bytes)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

---

## What's New in v9

RDB v9 introduced proper **little-endian byte ordering** for millisecond timestamps on all architectures. Prior to v9, `rdbLoadSignedInteger` did not call `memrev64ifbe()`, causing big-endian systems to produce byte-swapped timestamps that could not be loaded on little-endian systems (or vice versa).

Source: [`src/rdb.c:156–161`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L156) @ `cbdad795d`:
```c
int64_t rdbLoadSignedInteger(rio *rdb, int rdbver) {
    int64_t val;
    if (rioRead(rdb, &val, 8) == 0) return INT64_MAX;
    if (rdbver >= 9) /* Fix BE/LE portability */
        memrev64ifbe(&val);
    return val;
}
```

## Available Type Bytes

| Type | Hex | Name | Encoding | Source (define) |
|------|-----|------|----------|-----------------|
| 0 | `0x00` | STRING | embstr / raw / int | [`rdb.h:55`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L55) |
| 1 | `0x01` | LIST | plain list | [`rdb.h:56`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L56) |
| 2 | `0x02` | SET | plain set | [`rdb.h:57`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L57) |
| 3 | `0x03` | ZSET | string scores | [`rdb.h:58`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L58) |
| 4 | `0x04` | HASH | plain hash | [`rdb.h:59`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L59) |
| 5 | `0x05` | ZSET_2 | binary double scores | [`rdb.h:60`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L60) |
| 6 | `0x06` | MODULE_PRE_GA | Redis 4.0 RC modules | [`rdb.h:61`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L61) |
| 7 | `0x07` | MODULE_2 | self-describing modules | [`rdb.h:62`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L62) |
| 8 | — | *(unassigned)* | — | [`rdb.h:95`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L95) — `rdbIsObjectType()` skips 8 |
| 9 | `0x09` | HASH_ZIPMAP | zipmap blob | [`rdb.h:64`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L64) |
| 10 | `0x0A` | LIST_ZIPLIST | ziplist blob | [`rdb.h:65`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L65) |
| 11 | `0x0B` | SET_INTSET | intset blob | [`rdb.h:66`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L66) |
| 12 | `0x0C` | ZSET_ZIPLIST | ziplist blob | [`rdb.h:67`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L67) |
| 13 | `0x0D` | HASH_ZIPLIST | ziplist blob | [`rdb.h:68`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L68) |
| 14 | `0x0E` | LIST_QUICKLIST | node-count + ziplist blobs | [`rdb.h:69`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L69) |
| 15 | `0x0F` | STREAM_LISTPACKS | radix tree + metadata + CGs | [`rdb.h:70`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L70) |
| 16 | `0x10` | HASH_LISTPACK | listpack blob | [`rdb.h:71`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L71) |
| 17 | `0x11` | ZSET_LISTPACK | listpack blob | [`rdb.h:72`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L72) |

**Types NOT available in v9:** 18–33 (introduced in v10+).

## Available Opcodes

| Opcode | Hex | Available | Source (define) |
|--------|-----|-----------|-----------------|
| MODULE_AUX | `0xF7` | ✅ | [`rdb.h:106`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L106) |
| IDLE | `0xF8` | ✅ | [`rdb.h:107`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L107) |
| FREQ | `0xF9` | ✅ | [`rdb.h:108`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L108) |
| AUX | `0xFA` | ✅ | [`rdb.h:109`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L109) |
| RESIZEDB | `0xFB` | ✅ | [`rdb.h:110`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L110) |
| EXPIRETIME_MS | `0xFC` | ✅ | [`rdb.h:111`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L111) |
| EXPIRETIME | `0xFD` | ✅ | [`rdb.h:112`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L112) |
| SELECTDB | `0xFE` | ✅ | [`rdb.h:113`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L113) |
| EOF | `0xFF` | ✅ | [`rdb.h:114`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L114) |

**Opcodes NOT available in v9:** `0xF2`–`0xF6` (introduced in v10+).

## Rediscope v1:beta Handling

| Aspect | Status | Rediscope Source |
|--------|--------|------------------|
| Header parsing | ✅ | [`parser.go:35–42`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/parser.go#L35) |
| All 17 type bytes | ✅ | [`types.go:36–76`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L36) `TypeName()` |
| Type 8 gap | ✅ Correct | [`types.go:120–151`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L120) `GeneralType()` — falls to `default: "unknown"` |
| Byte boundary tracking | ✅ | [`reader.go:180`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/reader.go#L180) `readKeyRecord()` |
| Value extraction | ⚠️ Skipped | [`reader.go:216`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/reader.go#L216) `skipValue()` — boundaries only |
| LZF decompression | ⚠️ Not decoded | [`reader.go:89`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/reader.go#L89) `readString()` — returns `<lzf>` |
| HLL/bitmap heuristic | ⚠️ Key suffix | [`types.go:122–131`](file:///home/keshav2002latest/dev/redis_lab/rediscope/internal/rdb/types.go#L122) |

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-6.2.23-bulk.rdb` | 468 | 27 KB | ✅ Parsed in CI |
| `redis-6.2.23-redis-tests-complex.rdb` | 53,671 | 64 MB | ❌ Not in automated tests |
