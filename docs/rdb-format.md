# RDB Binary Format Specification

Common format invariant across all RDB versions (v9–v15). Per-version additions are documented in individual `rdb-vN.md` files.

> **Redis Source:** All references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2) on branch `unstable`.

---

## 1. File Layout

<!-- Source: rdbSaveRio() at src/rdb.c:2027 @ cbdad795d -->

```
┌─────────────────────────────────────────────────────┐
│  Magic Header (9 bytes)                             │
├─────────────────────────────────────────────────────┤
│  Global Metadata (AUX fields, Functions, Templates) │
├─────────────────────────────────────────────────────┤
│  Database 0                                         │
│    ├─ SELECTDB opcode + DB index                    │
│    ├─ RESIZEDB opcode + size hints                  │
│    └─ Key-Value records...                          │
├─────────────────────────────────────────────────────┤
│  Database 1..N  (repeat)                            │
├─────────────────────────────────────────────────────┤
│  EOF opcode (1 byte: 0xFF)                          │
├─────────────────────────────────────────────────────┤
│  CRC64 checksum (8 bytes)  [rdbver >= 5]            │
└─────────────────────────────────────────────────────┘
```

Source: [`src/rdb.c:2027`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2027) `rdbSaveRio()` — write path; [`src/rdb.c:4691`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4691) `rdbLoadRio()` magic check — read path.

## 2. Magic Header — 9 bytes, FIXED

| Offset | Size | Content | Hex Example (v15) |
|--------|------|---------|-------------------|
| 0–4 | 5 bytes | ASCII `REDIS` | `52 45 44 49 53` |
| 5–8 | 4 bytes | Zero-padded ASCII version | `30 30 31 35` |

The version string is always exactly 4 ASCII digits. Examples: `0009` (v9), `0010` (v10), `0015` (v15).

Source: [`src/rdb.h:21`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L21) — `#define RDB_VERSION 15`; [`src/rdb.c:4691`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4691) — `memcmp(buf,"REDIS",5)`.

## 3. Length Encoding

All counts, sizes, and string lengths use this variable-width encoding. The first 2 bits of the first byte determine the mode.

Source: [`src/rdb.h:37–42`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L37) — constant definitions; [`src/rdb.c:176`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L176) `rdbSaveLen()`; [`src/rdb.c:219`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L219) `rdbLoadLenByRef()`.

| Mode | First Byte Bits | Total Bytes | Value Range | Byte Layout | Source |
|------|----------------|-------------|-------------|-------------|--------|
| 6-bit | `00xxxxxx` | 1 | 0–63 | `[0x00..0x3F]` | `RDB_6BITLEN` [`rdb.h:37`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L37) |
| 14-bit | `01xxxxxx` | 2 | 0–16,383 | `[0x40..0x7F] [xx]` | `RDB_14BITLEN` [`rdb.h:38`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L38) |
| 32-bit | `10000000` | 5 | 0–2³² | `[0x80] [B3 B2 B1 B0]` (BE) | `RDB_32BITLEN` [`rdb.h:39`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L39) |
| 64-bit | `10000001` | 9 | 0–2⁶⁴ | `[0x81] [B7..B0]` (BE) | `RDB_64BITLEN` [`rdb.h:40`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L40) |
| Special | `11xxxxxx` | varies | N/A | String encoding (see §4) | `RDB_ENCVAL` [`rdb.h:41`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L41) |

**Big Endian (network byte order)** is used for 32-bit and 64-bit lengths.

### Examples

| Value | Encoded Bytes |
|-------|--------------|
| `5` | `05` |
| `100` | `40 64` |
| `1000` | `43 E8` |
| `100000` | `80 00 01 86 A0` |

## 4. String Encoding

Strings are either raw byte sequences or specially encoded integers/compressed data.

Source: [`src/rdb.h:47–50`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L47) — encoding constants; [`src/rdb.c:533`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L533) `rdbGenericLoadStringObjectUsable()`.

### Raw String
```
[length-encoded size] [raw bytes...]
```

### Special Encodings (first byte has `11` prefix)

| Encoding | Prefix Byte | Payload | Source |
|----------|-------------|---------|--------|
| INT8 | `0xC0` | 1 byte | `RDB_ENC_INT8` [`rdb.h:47`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L47) |
| INT16 | `0xC1` | 2 bytes (LE) | `RDB_ENC_INT16` [`rdb.h:48`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L48) |
| INT32 | `0xC2` | 4 bytes (LE) | `RDB_ENC_INT32` [`rdb.h:49`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L49) |
| LZF | `0xC3` | `[clen] [ulen] [data]` | `RDB_ENC_LZF` [`rdb.h:50`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L50) |

For LZF: `clen` (compressed length) and `ulen` (uncompressed length) are each length-encoded, followed by `clen` bytes of compressed data. Source: [`src/rdb.c:411`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L411).

### Examples

| Value | Encoding | Bytes |
|-------|----------|-------|
| Integer `64` | INT8 | `C0 40` |
| Integer `1167376` | INT32 | `C2 10 D0 11 00` |
| String `"REDIS"` | Raw, len=5 | `05 52 45 44 49 53` |

## 5. Opcodes

Special bytes (`>= 0xF2`) that control parsing flow. Any byte `< 0xF2` in opcode position is a **type byte** (see §6).

Source: [`src/rdb.h:101–114`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L101) — opcode definitions; [`src/rdb.c:4715–4911`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4715) — opcode handlers in `rdbLoadRio()`.

| Opcode | Hex | Name | Payload | Source (define) | Source (handler) |
|--------|-----|------|---------|-----------------|------------------|
| `0xF2` | 242 | HASH_TEMPLATE | `[field_count] [field_name_0]...[field_name_N-1]` | [`rdb.h:101`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L101) | [`rdb.c:4911`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4911) |
| `0xF3` | 243 | KEY_META | `[length-encoded metadata payload]` | [`rdb.h:102`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L102) | [`rdb.c:2348`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2348) |
| `0xF4` | 244 | SLOT_INFO | `[slot_id] [slot_size] [expires_size]` — all length-encoded | [`rdb.h:103`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L103) | [`rdb.c:4767`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4767) |
| `0xF5` | 245 | FUNCTION2 | `[string-encoded function library data]` | [`rdb.h:104`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L104) | [`rdb.c:4903`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4903) |
| `0xF6` | 246 | FUNCTION_PRE_GA | `[string-encoded function library data]` — RC only | [`rdb.h:105`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L105) | [`rdb.c:4900`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4900) |
| `0xF7` | 247 | MODULE_AUX | `[module_id (uint64)] [when] [module data...]` | [`rdb.h:106`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L106) | [`rdb.c:4847`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4847) |
| `0xF8` | 248 | IDLE | `[length-encoded LRU idle time]` | [`rdb.h:107`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L107) | [`rdb.c:4737`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4737) |
| `0xF9` | 249 | FREQ | `[1 byte: LFU frequency]` | [`rdb.h:108`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L108) | [`rdb.c:4731`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4731) |
| `0xFA` | 250 | AUX | `[string key] [string value]` | [`rdb.h:109`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L109) | [`rdb.c:4783`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4783) |
| `0xFB` | 251 | RESIZEDB | `[length db_size] [length expires_size]` | [`rdb.h:110`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L110) | [`rdb.c:4758`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4758) |
| `0xFC` | 252 | EXPIRETIME_MS | `[8 bytes: int64 milliseconds, LE]` | [`rdb.h:111`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L111) | [`rdb.c:4724`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4724) |
| `0xFD` | 253 | EXPIRETIME | `[4 bytes: uint32 seconds, LE]` | [`rdb.h:112`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L112) | [`rdb.c:4715`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4715) |
| `0xFE` | 254 | SELECTDB | `[length-encoded DB index]` | [`rdb.h:113`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L113) | [`rdb.c:4746`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4746) |
| `0xFF` | 255 | EOF | *(no payload — followed by 8-byte CRC64)* | [`rdb.h:114`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L114) | [`rdb.c:4743`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4743) |

### Common AUX Keys

| Key | Value Type | Example |
|-----|-----------|---------|
| `redis-ver` | string | `"8.9.241"` |
| `redis-bits` | integer | `64` |
| `ctime` | integer | Unix timestamp (seconds) |
| `used-mem` | integer | Memory usage in bytes |
| `aof-base` | integer | `0` or `1` |
| `repl-id` | string | Replication ID |
| `repl-offset` | integer | Replication offset |

## 6. Type Bytes

Type bytes identify the data structure and encoding of a key-value record. Valid range: `0x00`–`0x07` and `0x09`–`0x21`. **Type 8 (`0x08`) is intentionally unassigned** in all Redis versions.

Source: [`src/rdb.h:55–90`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L55) — type definitions; [`src/rdb.h:95–97`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L95) — `rdbIsObjectType()` macro confirming type 8 gap; [`src/rdb.c:2898`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2898) `rdbLoadObject()` — type dispatch.

### String Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 0 | `0x00` | STRING | Single string-encoded value | [`rdb.h:55`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L55) | [`rdb.c:2918`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2918) |

### List Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 1 | `0x01` | LIST | `[count] [string element]...` | [`rdb.h:56`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L56) | [`rdb.c:2922`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2922) |
| 10 | `0x0A` | LIST_ZIPLIST | Single string blob (ziplist) | [`rdb.h:65`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L65) | [`rdb.c:3683`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3683) |
| 14 | `0x0E` | LIST_QUICKLIST | `[node_count] [string blob]...` | [`rdb.h:69`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L69) | [`rdb.c:3513`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3513) |
| 18 | `0x12` | LIST_QUICKLIST_2 | `[node_count] ([container_fmt] [string blob])...` | [`rdb.h:73`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L73) | [`rdb.c:3523`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3523) |

### Set Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 2 | `0x02` | SET | `[count] [string element]...` | [`rdb.h:57`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L57) | [`rdb.c:2943`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L2943) |
| 11 | `0x0B` | SET_INTSET | Single string blob (intset) | [`rdb.h:66`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L66) | [`rdb.c:3713`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3713) |
| 20 | `0x14` | SET_LISTPACK | Single string blob (listpack) | [`rdb.h:75`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L75) | [`rdb.c:3727`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3727) |

### Sorted Set Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 3 | `0x03` | ZSET | `[count] ([string member] [string score])...` | [`rdb.h:58`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L58) | [`rdb.c:3043`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3043) |
| 5 | `0x05` | ZSET_2 | `[count] ([string member] [8-byte IEEE754 double])...` | [`rdb.h:60`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L60) | [`rdb.c:3072`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3072) |
| 12 | `0x0C` | ZSET_ZIPLIST | Single string blob (ziplist) | [`rdb.h:67`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L67) | [`rdb.c:3748`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3748) |
| 17 | `0x11` | ZSET_LISTPACK | Single string blob (listpack) | [`rdb.h:72`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L72) | [`rdb.c:3775`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3775) |

### Hash Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 4 | `0x04` | HASH | `[count] ([string field] [string value])...` | [`rdb.h:59`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L59) | [`rdb.c:3114`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3114) |
| 9 | `0x09` | HASH_ZIPMAP | Single string blob (zipmap) | [`rdb.h:64`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L64) | [`rdb.c:3626`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3626) |
| 13 | `0x0D` | HASH_ZIPLIST | Single string blob (ziplist) | [`rdb.h:68`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L68) | [`rdb.c:3794`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3794) |
| 16 | `0x10` | HASH_LISTPACK | Single string blob (listpack) | [`rdb.h:71`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L71) | [`rdb.c:3821`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3821) |
| 22 | `0x16` | HASH_METADATA_PRE_GA | `[count] ([string field] [uint64 TTL] [string value])...` | [`rdb.h:77`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L77) | [`rdb.c:3343`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3343) |
| 23 | `0x17` | HASH_LISTPACK_EX_PRE_GA | Single string blob (listpack with TTLs) | [`rdb.h:78`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L78) | [`rdb.c:3822`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3822) |
| 24 | `0x18` | HASH_METADATA | `[uint64 minExpire] [count] ([field] [TTL delta] [value])...` | [`rdb.h:79`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L79) | [`rdb.c:3356`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3356) |
| 25 | `0x19` | HASH_LISTPACK_EX | `[uint64 minExpire] [string blob (listpack with TTLs)]` | [`rdb.h:80`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L80) | [`rdb.c:3602`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3602) |
| 29 | `0x1D` | HASH_TMPL_LP | `[field_count] [fields...] [listpack blob]` — self-contained | [`rdb.h:84`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L84) | [`rdb.c:3253`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3253) |
| 30 | `0x1E` | HASH_TMPL_LP_REF | Listpack blob, first entry is template ID | [`rdb.h:85`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L85) | [`rdb.c:3273`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3273) |
| 31 | `0x1F` | HASH_TMPL_ARRAY | `[field_count] ([field] [value])...` — self-contained | [`rdb.h:86`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L86) | [`rdb.c:3303`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3303) |
| 32 | `0x20` | HASH_TMPL_ARRAY_REF | `[template_id] [value_0]...[value_N-1]` | [`rdb.h:87`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L87) | [`rdb.c:3321`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3321) |

### Stream Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 15 | `0x0F` | STREAM_LISTPACKS | Radix tree nodes + metadata + consumer groups | [`rdb.h:70`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L70) | [`rdb.c:3869`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3869) |
| 19 | `0x13` | STREAM_LISTPACKS_2 | + entries_added, first/max_deleted IDs | [`rdb.h:74`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L74) | [`rdb.c:3870`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3870) |
| 21 | `0x15` | STREAM_LISTPACKS_3 | + per-consumer active_time, entries_read | [`rdb.h:76`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L76) | [`rdb.c:3871`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3871) |
| 26 | `0x1A` | STREAM_LISTPACKS_4 | + IDMP support | [`rdb.h:81`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L81) | [`rdb.c:3872`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3872) |
| 27 | `0x1B` | STREAM_LISTPACKS_5 | + XNACK support | [`rdb.h:82`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L82) | [`rdb.c:3873`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L3873) |

### Module Family

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 6 | `0x06` | MODULE_PRE_GA | Module-specific (Redis 4.0 RC) | [`rdb.h:61`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L61) | [`rdb.c:4323`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4323) |
| 7 | `0x07` | MODULE_2 | `[uint64 module_id]` + self-describing opcodes ending with EOF(0) | [`rdb.h:62`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L62) | [`rdb.c:4326`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4326) |

### Other

| Type | Hex | Name | Value Format | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 28 | `0x1C` | ARRAY | `[count] ([tag_byte] [payload])...` — tags: 0=SDS, 1=int64, 2=double, 3=small_str | [`rdb.h:83`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L83) | [`rdb.c:4393`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4393) |
| 33 | `0x21` | GCRA | GCRA object (compile-time gated via `ENABLE_GCRA`) | [`rdb.h:89`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L89) | [`rdb.c:4385`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4385) |

## 7. Key-Value Record Format

```
[optional: 0xFC + 8-byte ms timestamp  OR  0xFD + 4-byte sec timestamp]
[optional: 0xF8 + length-encoded idle  OR  0xF9 + 1-byte freq]
[1 byte: Type Byte]
[String-encoded Key]
[Type-specific Value Payload]
```

The type byte is always exactly 1 byte. The key is always string-encoded. The value format is determined solely by the type byte.

Source: [`src/rdb.c:4715–4945`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L4715) — opcode/type dispatch loop in `rdbLoadRio()`.

## 8. EOF and Checksum

```
[0xFF]                    ← EOF opcode (1 byte)
[8 bytes: CRC64]          ← present when rdbver >= 5
```

The CRC64 is computed over all bytes from the start of the file through the EOF opcode (inclusive). For `rdbver >= 5`, the checksum is always present. Redis validates `cksum != 0` (a zero checksum means validation was disabled at save time).

Source: [`src/rdb.c:5049`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L5049) — `if (rdbver >= 5)` checksum guard.

## 9. Version-Specific Notes

- **rdbver >= 5**: CRC64 checksum is appended after EOF. Source: [`src/rdb.c:5049`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L5049).
- **rdbver >= 9**: Millisecond timestamps (`rdbLoadSignedInteger`) use proper little-endian byte order on all architectures. Pre-v9 files on big-endian systems may have swapped timestamp bytes. Source: [`src/rdb.c:156–161`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c#L156).
