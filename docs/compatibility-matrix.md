# Compatibility Matrix — rediscope v1:beta

Exhaustive compatibility report for all RDB versions, opcodes, type bytes, and encoding modes.

> **Redis source:** All references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2) on branch `unstable`.
> **Rediscope source:** References cite the current `rediscope/` tree.

---

## 1. Version Support

| RDB | Redis Versions | Magic | Fixture (bulk) | Fixture (complex) | Status |
|-----|---------------|-------|----------------|-------------------|--------|
| v9 | 6.0, 6.2 LTS | `REDIS0009` | `redis-6.2.23-bulk.rdb` (27,717 B) | `redis-6.2.23-redis-tests-complex.rdb` (64,705,666 B, 53,671 keys) | ✅ |
| v10 | 7.0 LTS | `REDIS0010` | `redis-7.0.15-bulk.rdb` (27,691 B) | `redis-7.0.15-redis-tests-complex.rdb` (64,667,999 B, 53,844 keys) | ✅ |
| v11 | 7.2 LTS | `REDIS0011` | `redis-7.2.15-bulk.rdb` (27,691 B) | `redis-7.2.15-redis-tests-complex.rdb` (64,659,135 B, 54,166 keys) | ✅ |
| v12 | 7.4 LTS, 8.0, 8.2, 8.4 | `REDIS0012` | 4 fixtures (27 KB each) | 4 fixtures (~64 MB each) | ✅ |
| v13 | 8.6 | `REDIS0013` | `redis-8.6.5-bulk.rdb` (26,976 B) | `redis-8.6.5-redis-tests-complex.rdb` (64,696,888 B, 53,939 keys) | ✅ |
| v14 | 8.8 | `REDIS0014` | `redis-8.8.1-bulk.rdb` (26,976 B) | `redis-8.8.1-redis-tests-complex.rdb` (64,658,953 B, 53,930 keys) | ✅ |
| v15 | 8.9+ (trunk) | `REDIS0015` | `native-types.rdb` (483 B, 11 keys) | — | ✅ |

RDB version constant: [`src/rdb.h:21`](src/rdb.h#L21) — `#define RDB_VERSION 15`

## 2. Opcode Support

Source (definitions): [`src/rdb.h:101–114`](src/rdb.h#L101)
Source (handlers): [`src/rdb.c:4715–4911`](src/rdb.c#L4715)

| Opcode | Hex | Introduced | Redis Source | Rediscope Handling | Rediscope Source | Status |
|--------|-----|------------|-------------|-------------------|------------------|--------|
| EOF | `0xFF` | v1 | [`rdb.h:114`](src/rdb.h#L114), [`rdb.c:4743`](src/rdb.c#L4743) | Terminates parse loop | [`parser.go:411`](../internal/rdb/parser.go#L411) | ✅ |
| SELECTDB | `0xFE` | v1 | [`rdb.h:113`](src/rdb.h#L113), [`rdb.c:4746`](src/rdb.c#L4746) | Reads length-encoded DB index | [`parser.go:208`](../internal/rdb/parser.go#L208) | ✅ |
| EXPIRETIME | `0xFD` | v1 | [`rdb.h:112`](src/rdb.h#L112), [`rdb.c:4715`](src/rdb.c#L4715) | Reads 4 bytes LE → `pendingExpiry` | [`parser.go:290`](../internal/rdb/parser.go#L290) | ✅ |
| EXPIRETIME_MS | `0xFC` | v1 | [`rdb.h:111`](src/rdb.h#L111), [`rdb.c:4724`](src/rdb.c#L4724) | Reads 8 bytes LE → `pendingExpiry` | [`parser.go:281`](../internal/rdb/parser.go#L281) | ✅ |
| RESIZEDB | `0xFB` | v7 | [`rdb.h:110`](src/rdb.h#L110), [`rdb.c:4758`](src/rdb.c#L4758) | Reads two length-encoded values | [`parser.go:229`](../internal/rdb/parser.go#L229) | ✅ |
| AUX | `0xFA` | v7 | [`rdb.h:109`](src/rdb.h#L109), [`rdb.c:4783`](src/rdb.c#L4783) | Reads two strings; special summaries for known keys | [`parser.go:160`](../internal/rdb/parser.go#L160) | ✅ |
| FREQ | `0xF9` | v8 | [`rdb.h:108`](src/rdb.h#L108), [`rdb.c:4731`](src/rdb.c#L4731) | Reads 1 byte | [`parser.go:140`](../internal/rdb/parser.go#L140) | ✅ |
| IDLE | `0xF8` | v8 | [`rdb.h:107`](src/rdb.h#L107), [`rdb.c:4737`](src/rdb.c#L4737) | Reads length-encoded value | [`parser.go:120`](../internal/rdb/parser.go#L120) | ✅ |
| MODULE_AUX | `0xF7` | v9 | [`rdb.h:106`](src/rdb.h#L106), [`rdb.c:4847`](src/rdb.c#L4847) | Reads and skips module payload | [`parser.go:327`](../internal/rdb/parser.go#L327) | ✅ |
| FUNCTION_PRE_GA | `0xF6` | v10 | [`rdb.h:105`](src/rdb.h#L105), [`rdb.c:4900`](src/rdb.c#L4900) | **Returns immediate error** | [`parser.go:318`](../internal/rdb/parser.go#L318) | ⚠️ Aborts |
| FUNCTION2 | `0xF5` | v10 | [`rdb.h:104`](src/rdb.h#L104), [`rdb.c:4903`](src/rdb.c#L4903) | Reads and skips function library string | [`parser.go:300`](../internal/rdb/parser.go#L300) | ✅ |
| SLOT_INFO | `0xF4` | v12 | [`rdb.h:103`](src/rdb.h#L103), [`rdb.c:4767`](src/rdb.c#L4767) | Reads slot ID and size hints | [`parser.go:253`](../internal/rdb/parser.go#L253) | ✅ |
| KEY_META | `0xF3` | v12 | [`rdb.h:102`](src/rdb.h#L102), [`rdb.c:2348`](src/rdb.c#L2348) | Reads and skips metadata payload | [`parser.go:360`](../internal/rdb/parser.go#L360) | ✅ |
| HASH_TEMPLATE | `0xF2` | v14 | [`rdb.h:101`](src/rdb.h#L101), [`rdb.c:4911`](src/rdb.c#L4911) | Reads fields, registers in parser state | [`parser.go:387`](../internal/rdb/parser.go#L387) | ✅ |

## 3. Type Byte Support

Source (definitions): [`src/rdb.h:55–90`](src/rdb.h#L55)
Source (type dispatch): [`src/rdb.c:2898`](src/rdb.c#L2898) `rdbLoadObject()`
Source (type 8 gap): [`src/rdb.h:95`](src/rdb.h#L95) `rdbIsObjectType()` — explicitly skips `(t) == 8`
Rediscope type maps: [`types.go:36`](../internal/rdb/types.go#L36) `TypeName()`, [`types.go:120`](../internal/rdb/types.go#L120) `GeneralType()`

| Type | Hex | Name | Intro | Redis define | Redis handler | Rediscope | Status |
|------|-----|------|-------|-------------|---------------|-----------|--------|
| 0 | `0x00` | STRING | v1 | [`rdb.h:55`](src/rdb.h#L55) | [`rdb.c:2918`](src/rdb.c#L2918) | [`types.go:38`](../internal/rdb/types.go#L38) | ✅ |
| 1 | `0x01` | LIST | v1 | [`rdb.h:56`](src/rdb.h#L56) | [`rdb.c:2922`](src/rdb.c#L2922) | [`types.go:39`](../internal/rdb/types.go#L39) | ✅ |
| 2 | `0x02` | SET | v1 | [`rdb.h:57`](src/rdb.h#L57) | [`rdb.c:2943`](src/rdb.c#L2943) | [`types.go:40`](../internal/rdb/types.go#L40) | ✅ |
| 3 | `0x03` | ZSET | v1 | [`rdb.h:58`](src/rdb.h#L58) | [`rdb.c:3043`](src/rdb.c#L3043) | [`types.go:41`](../internal/rdb/types.go#L41) | ✅ |
| 4 | `0x04` | HASH | v1 | [`rdb.h:59`](src/rdb.h#L59) | [`rdb.c:3114`](src/rdb.c#L3114) | [`types.go:42`](../internal/rdb/types.go#L42) | ✅ |
| 5 | `0x05` | ZSET_2 | v2 | [`rdb.h:60`](src/rdb.h#L60) | [`rdb.c:3072`](src/rdb.c#L3072) | [`types.go:43`](../internal/rdb/types.go#L43) | ✅ |
| 6 | `0x06` | MODULE_PRE_GA | v8 | [`rdb.h:61`](src/rdb.h#L61) | [`rdb.c:4323`](src/rdb.c#L4323) | [`types.go:44`](../internal/rdb/types.go#L44) | ✅ |
| 7 | `0x07` | MODULE_2 | v8 | [`rdb.h:62`](src/rdb.h#L62) | [`rdb.c:4326`](src/rdb.c#L4326) | [`types.go:45`](../internal/rdb/types.go#L45) | ✅ |
| **8** | — | **(unassigned)** | — | [`rdb.h:95`](src/rdb.h#L95) | N/A | N/A | **Intentionally unused** |
| 9 | `0x09` | HASH_ZIPMAP | v1 | [`rdb.h:64`](src/rdb.h#L64) | [`rdb.c:3626`](src/rdb.c#L3626) | [`types.go:46`](../internal/rdb/types.go#L46) | ✅ |
| 10 | `0x0A` | LIST_ZIPLIST | v1 | [`rdb.h:65`](src/rdb.h#L65) | [`rdb.c:3683`](src/rdb.c#L3683) | [`types.go:47`](../internal/rdb/types.go#L47) | ✅ |
| 11 | `0x0B` | SET_INTSET | v1 | [`rdb.h:66`](src/rdb.h#L66) | [`rdb.c:3713`](src/rdb.c#L3713) | [`types.go:48`](../internal/rdb/types.go#L48) | ✅ |
| 12 | `0x0C` | ZSET_ZIPLIST | v1 | [`rdb.h:67`](src/rdb.h#L67) | [`rdb.c:3748`](src/rdb.c#L3748) | [`types.go:49`](../internal/rdb/types.go#L49) | ✅ |
| 13 | `0x0D` | HASH_ZIPLIST | v1 | [`rdb.h:68`](src/rdb.h#L68) | [`rdb.c:3794`](src/rdb.c#L3794) | [`types.go:50`](../internal/rdb/types.go#L50) | ✅ |
| 14 | `0x0E` | LIST_QUICKLIST | v7 | [`rdb.h:69`](src/rdb.h#L69) | [`rdb.c:3513`](src/rdb.c#L3513) | [`types.go:51`](../internal/rdb/types.go#L51) | ✅ |
| 15 | `0x0F` | STREAM_LISTPACKS | v9 | [`rdb.h:70`](src/rdb.h#L70) | [`rdb.c:3869`](src/rdb.c#L3869) | [`types.go:52`](../internal/rdb/types.go#L52) | ✅ |
| 16 | `0x10` | HASH_LISTPACK | v9 | [`rdb.h:71`](src/rdb.h#L71) | [`rdb.c:3821`](src/rdb.c#L3821) | [`types.go:53`](../internal/rdb/types.go#L53) | ✅ |
| 17 | `0x11` | ZSET_LISTPACK | v9 | [`rdb.h:72`](src/rdb.h#L72) | [`rdb.c:3775`](src/rdb.c#L3775) | [`types.go:54`](../internal/rdb/types.go#L54) | ✅ |
| 18 | `0x12` | LIST_QUICKLIST_2 | v10 | [`rdb.h:73`](src/rdb.h#L73) | [`rdb.c:3523`](src/rdb.c#L3523) | [`types.go:55`](../internal/rdb/types.go#L55) | ✅ |
| 19 | `0x13` | STREAM_LISTPACKS_2 | v10 | [`rdb.h:74`](src/rdb.h#L74) | [`rdb.c:3870`](src/rdb.c#L3870) | [`types.go:56`](../internal/rdb/types.go#L56) | ✅ |
| 20 | `0x14` | SET_LISTPACK | v10 | [`rdb.h:75`](src/rdb.h#L75) | [`rdb.c:3727`](src/rdb.c#L3727) | [`types.go:57`](../internal/rdb/types.go#L57) | ✅ |
| 21 | `0x15` | STREAM_LISTPACKS_3 | v11 | [`rdb.h:76`](src/rdb.h#L76) | [`rdb.c:3871`](src/rdb.c#L3871) | [`types.go:58`](../internal/rdb/types.go#L58) | ✅ |
| 22 | `0x16` | HASH_METADATA_PRE_GA | v12 | [`rdb.h:77`](src/rdb.h#L77) | [`rdb.c:3343`](src/rdb.c#L3343) | [`types.go:59`](../internal/rdb/types.go#L59) | ✅ |
| 23 | `0x17` | HASH_LISTPACK_EX_PRE_GA | v12 | [`rdb.h:78`](src/rdb.h#L78) | [`rdb.c:3822`](src/rdb.c#L3822) | [`types.go:60`](../internal/rdb/types.go#L60) | ✅ |
| 24 | `0x18` | HASH_METADATA | v12 | [`rdb.h:79`](src/rdb.h#L79) | [`rdb.c:3356`](src/rdb.c#L3356) | [`types.go:61`](../internal/rdb/types.go#L61) | ✅ |
| 25 | `0x19` | HASH_LISTPACK_EX | v12 | [`rdb.h:80`](src/rdb.h#L80) | [`rdb.c:3602`](src/rdb.c#L3602) | [`types.go:62`](../internal/rdb/types.go#L62) | ✅ |
| 26 | `0x1A` | STREAM_LISTPACKS_4 | v12 | [`rdb.h:81`](src/rdb.h#L81) | [`rdb.c:3872`](src/rdb.c#L3872) | [`types.go:63`](../internal/rdb/types.go#L63) | ✅ |
| 27 | `0x1B` | STREAM_LISTPACKS_5 | v13 | [`rdb.h:82`](src/rdb.h#L82) | [`rdb.c:3873`](src/rdb.c#L3873) | [`types.go:64`](../internal/rdb/types.go#L64) | ✅ |
| 28 | `0x1C` | ARRAY | v13 | [`rdb.h:83`](src/rdb.h#L83) | [`rdb.c:4393`](src/rdb.c#L4393) | [`types.go:65`](../internal/rdb/types.go#L65) | ✅ |
| 29 | `0x1D` | HASH_TMPL_LP | v14 | [`rdb.h:84`](src/rdb.h#L84) | [`rdb.c:3253`](src/rdb.c#L3253) | [`types.go:66`](../internal/rdb/types.go#L66) | ✅ |
| 30 | `0x1E` | HASH_TMPL_LP_REF | v14 | [`rdb.h:85`](src/rdb.h#L85) | [`rdb.c:3273`](src/rdb.c#L3273) | [`types.go:67`](../internal/rdb/types.go#L67) | ✅ |
| 31 | `0x1F` | HASH_TMPL_ARRAY | v14 | [`rdb.h:86`](src/rdb.h#L86) | [`rdb.c:3303`](src/rdb.c#L3303) | [`types.go:68`](../internal/rdb/types.go#L68) | ✅ |
| 32 | `0x20` | HASH_TMPL_ARRAY_REF | v14 | [`rdb.h:87`](src/rdb.h#L87) | [`rdb.c:3321`](src/rdb.c#L3321) | [`types.go:69`](../internal/rdb/types.go#L69) | ✅ |
| 33 | `0x21` | GCRA | conditional | [`rdb.h:89`](src/rdb.h#L89) | [`rdb.c:4385`](src/rdb.c#L4385) | [`types.go:70`](../internal/rdb/types.go#L70) | ✅ |

## 4. Length Encoding Support

Source: [`src/rdb.h:37–42`](src/rdb.h#L37); [`src/rdb.c:219`](src/rdb.c#L219) `rdbLoadLenByRef()`
Rediscope: [`reader.go:49`](../internal/rdb/reader.go#L49) `readLength()`

| Mode | First Byte | Total Bytes | Status |
|------|-----------|-------------|--------|
| 6-bit | `00xxxxxx` | 1 | ✅ |
| 14-bit | `01xxxxxx` | 2 | ✅ |
| 32-bit BE | `10000000` | 5 | ✅ |
| 64-bit BE | `10000001` | 9 | ✅ |
| Special | `11xxxxxx` | varies | ✅ |

## 5. String Encoding Support

Source: [`src/rdb.h:47–50`](src/rdb.h#L47); [`src/rdb.c:533`](src/rdb.c#L533) `rdbGenericLoadStringObjectUsable()`
Rediscope: [`reader.go:89`](../internal/rdb/reader.go#L89) `readString()`

| Encoding | Prefix | Payload | Status |
|----------|--------|---------|--------|
| Raw bytes | length + data | variable | ✅ |
| INT8 | `0xC0` | 1 byte | ✅ |
| INT16 | `0xC1` | 2 bytes (LE) | ✅ |
| INT32 | `0xC2` | 4 bytes (LE) | ✅ |
| LZF | `0xC3` | clen + ulen + data | ⚠️ **Skips only** — returns `<lzf>` literal |

## 6. Known Limitations

| # | Issue | Severity | Redis Source | Rediscope Source |
|---|-------|----------|-------------|------------------|
| 1 | Values not extracted | By Design | — | [`reader.go:216`](../internal/rdb/reader.go#L216) `skipValue()` |
| 2 | LZF not decompressed | Medium | [`rdb.h:50`](src/rdb.h#L50) | [`reader.go:89`](../internal/rdb/reader.go#L89) returns `<lzf>` |
| 3 | FUNCTION_PRE_GA abort | Low | [`rdb.h:105`](src/rdb.h#L105) | [`parser.go:318–325`](../internal/rdb/parser.go#L318) |
| 4 | Key name type heuristic | Cosmetic | — | [`types.go:122–131`](../internal/rdb/types.go#L122) |
| 5 | Type 8 gap | Correct | [`rdb.h:95`](src/rdb.h#L95) | [`types.go:149–150`](../internal/rdb/types.go#L149) `default: "unknown"` |
| 6 | No checksum validation | Low | [`rdb.c:5049`](src/rdb.c#L5049) | [`parser.go:411`](../internal/rdb/parser.go#L411) reads but doesn't verify |
| 7 | Integer overflow risk | Low | — | [`reader.go:149`](../internal/rdb/reader.go#L149) `skipBytes(n int)` from `uint64` |
| 8 | No big-endian support | Out of Scope | [`rdb.c:159`](src/rdb.c#L159) | — |

## 7. Test Fixture Coverage

| Redis Version | RDB | Bulk Fixture | Complex Fixture | Automated |
|--------------|-----|-------------|-----------------|-----------|
| 6.2.23 | v9 | 27,717 B / 468 keys | 64,705,666 B / 53,671 keys | ✅ Bulk & Complex |
| 7.0.15 | v10 | 27,691 B / 468 keys | 64,667,999 B / 53,844 keys | ✅ Bulk & Complex |
| 7.2.15 | v11 | 27,691 B / 468 keys | 64,659,135 B / 54,166 keys | ✅ Bulk & Complex |
| 7.4.10 | v12 | 27,691 B / 468 keys | 64,663,880 B / 53,923 keys | ✅ Bulk & Complex |
| 8.0.6 | v12 | 26,969 B / 468 keys | 64,687,183 B / 54,183 keys | ✅ Bulk & Complex |
| 8.2.8 | v12 | 26,969 B / 468 keys | 64,612,059 B / 53,806 keys | ✅ Bulk & Complex |
| 8.4.5 | v12 | 26,969 B / 468 keys | 64,637,409 B / 53,739 keys | ✅ Bulk & Complex |
| 8.6.5 | v13 | 26,976 B / 468 keys | 64,696,888 B / 53,939 keys | ✅ Bulk & Complex |
| 8.8.1 | v14 | 26,976 B / 468 keys | 64,658,953 B / 53,930 keys | ✅ Bulk & Complex |
| 8.9+ | v15 | `native-types.rdb` (483 B / 11 keys) | — | ✅ Unit test target |

> Both bulk fixtures and 64 MB complex stress-test fixtures (~54,000 keys each) are tested and verified in `test/rdb_parser_test.go`.
