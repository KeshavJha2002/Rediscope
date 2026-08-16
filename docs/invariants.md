# RDB Byte-Level Invariants

Encoding rules and structural guarantees that hold across all RDB versions (v9–v15). These invariants are the foundations that the rediscope parser relies on for correct byte-boundary tracking.

> **Redis source:** All references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2) on branch `unstable`.
> **Rediscope source:** References cite the current `rediscope/` tree.

---

## 1. File Structure Invariants

| # | Invariant | Byte Detail | Source |
|---|-----------|-------------|--------|
| F1 | Magic header is always exactly **9 bytes** | 5 bytes `REDIS` (`52 45 44 49 53`) + 4 bytes zero-padded ASCII version | [`rdb.c:4691`](src/rdb.c#L4691) — `memcmp(buf,"REDIS",5)` |
| F2 | EOF opcode (`0xFF`) always precedes checksum | Last 9 bytes of any v5+ file: `[FF] [8-byte CRC64]` | [`rdb.h:114`](src/rdb.h#L114), [`rdb.c:4743`](src/rdb.c#L4743) |
| F3 | Checksum is always **8 bytes** | Present when `rdbver >= 5`. CRC64 over all preceding bytes. | [`rdb.c:5049`](src/rdb.c#L5049) — `if (rdbver >= 5)` |
| F4 | SELECTDB always precedes key records | At least one `0xFE` opcode appears before any KV record | [`rdb.c:4746`](src/rdb.c#L4746) |
| F5 | RESIZEDB follows SELECTDB | When present, `0xFB` always appears after `0xFE` for the same DB | [`rdb.c:4758`](src/rdb.c#L4758) |
| F6 | AUX fields appear only in global metadata | `0xFA` opcodes appear before any SELECTDB, never between KV records | [`rdb.c:4783`](src/rdb.c#L4783) |
| F7 | File version is monotonically increasing | A higher version number is always a strict superset of lower versions | [`rdb.h:19–21`](src/rdb.h#L19) — version comment |

## 2. Length Encoding Invariants

Source: [`src/rdb.h:23–42`](src/rdb.h#L23) (format description + constants); [`src/rdb.c:176`](src/rdb.c#L176) `rdbSaveLen()`; [`src/rdb.c:219`](src/rdb.c#L219) `rdbLoadLenByRef()`.
Rediscope: [`reader.go:49`](../internal/rdb/reader.go#L49) `readLength()`.

| # | Invariant | Detail |
|---|-----------|--------|
| L1 | First 2 bits determine encoding mode | `00` = 6-bit, `01` = 14-bit, `10` = 32/64-bit, `11` = special |
| L2 | 6-bit length: exactly **1 byte** | Values 0–63. Byte: `0x00`–`0x3F`. Constant: `RDB_6BITLEN` [`rdb.h:37`](src/rdb.h#L37) |
| L3 | 14-bit length: exactly **2 bytes** | Values 0–16,383. First byte: `0x40`–`0x7F`. Constant: `RDB_14BITLEN` [`rdb.h:38`](src/rdb.h#L38) |
| L4 | 32-bit length: exactly **5 bytes** | Discriminator: `0x80`, then 4 bytes **Big Endian**. Constant: `RDB_32BITLEN` [`rdb.h:39`](src/rdb.h#L39) |
| L5 | 64-bit length: exactly **9 bytes** | Discriminator: `0x81`, then 8 bytes **Big Endian**. Constant: `RDB_64BITLEN` [`rdb.h:40`](src/rdb.h#L40) |
| L6 | Special encoding (`11` prefix) only valid in string context | Never appears as a standalone length for counts, sizes, or DB indexes. Constant: `RDB_ENCVAL` [`rdb.h:41`](src/rdb.h#L41) |
| L7 | Minimum encoding for a length: 1 byte | Maximum: 9 bytes (64-bit) |
| L8 | 32-bit and 64-bit use network byte order | MSB first (Big Endian) — unlike timestamps which are LE |

## 3. String Encoding Invariants

Source: [`src/rdb.h:47–50`](src/rdb.h#L47) (constants); [`src/rdb.c:533`](src/rdb.c#L533) `rdbGenericLoadStringObjectUsable()`.
Rediscope: [`reader.go:89`](../internal/rdb/reader.go#L89) `readString()`.

| # | Invariant | Byte Layout | Source |
|---|-----------|-------------|--------|
| S1 | INT8 prefix is always `0xC0` | `[C0] [1 byte signed int8]` — total 2 bytes | `RDB_ENC_INT8` [`rdb.h:47`](src/rdb.h#L47) |
| S2 | INT16 prefix is always `0xC1` | `[C1] [2 bytes signed int16 LE]` — total 3 bytes | `RDB_ENC_INT16` [`rdb.h:48`](src/rdb.h#L48) |
| S3 | INT32 prefix is always `0xC2` | `[C2] [4 bytes signed int32 LE]` — total 5 bytes | `RDB_ENC_INT32` [`rdb.h:49`](src/rdb.h#L49) |
| S4 | LZF prefix is always `0xC3` | `[C3] [compressed_len] [uncompressed_len] [compressed_bytes]` | `RDB_ENC_LZF` [`rdb.h:50`](src/rdb.h#L50) |
| S5 | Raw string: length then exact bytes | `[length-encoded N] [N raw bytes]` | [`rdb.c:533`](src/rdb.c#L533) |
| S6 | Integer strings use **Little Endian** | INT16 and INT32 payloads are LE — opposite of 32/64-bit lengths | [`rdb.c:533`](src/rdb.c#L533) |
| S7 | LZF sub-lengths use standard length encoding | `compressed_len` and `uncompressed_len` each use the §2 scheme | [`rdb.c:411`](src/rdb.c#L411) |

## 4. Opcode Invariants

Source: [`src/rdb.h:101–114`](src/rdb.h#L101) (definitions); [`src/rdb.h:95`](src/rdb.h#L95) `rdbIsObjectType()` (type range validation).
Rediscope: [`parser.go:114–411`](../internal/rdb/parser.go#L114) main parse loop.

| # | Invariant | Detail |
|---|-----------|--------|
| O1 | Opcode byte range: `0xF2`–`0xFF` | Any byte `< 0xF2` in opcode position is a **type byte**. Source: opcodes start at [`rdb.h:101`](src/rdb.h#L101) (`0xF2`) |
| O2 | Type byte range: `0x00`–`0x07` and `0x09`–`0x21` | Byte `0x08` is **never** valid. Source: [`rdb.h:95`](src/rdb.h#L95) — `rdbIsObjectType()` skips 8 |
| O3 | Expiry precedes its key | `0xFC` or `0xFD` always immediately before the type byte. Source: [`rdb.c:4715–4728`](src/rdb.c#L4715) |
| O4 | IDLE/FREQ precede their key | `0xF8` or `0xF9` between expiry (if any) and the type byte. Source: [`rdb.c:4731–4741`](src/rdb.c#L4731) |
| O5 | Order within a KV: `[expiry?] [idle/freq?] [type] [key] [value]` | Optional prefix opcodes are always in this order. Source: [`rdb.c:4715–4945`](src/rdb.c#L4715) |
| O6 | EOF is always exactly 1 byte: `0xFF` | No payload follows the opcode itself. Source: [`rdb.h:114`](src/rdb.h#L114) |

## 5. Timestamp Invariants

Source: [`src/rdb.c:156–161`](src/rdb.c#L156) `rdbLoadSignedInteger()`.
Rediscope: [`parser.go:281–298`](../internal/rdb/parser.go#L281).

| # | Invariant | Byte Layout |
|---|-----------|-------------|
| T1 | EXPIRETIME (`0xFD`): **4 bytes** fixed | `[FD] [uint32 seconds LE]` — total 5 bytes. Handler: [`rdb.c:4715`](src/rdb.c#L4715) |
| T2 | EXPIRETIME_MS (`0xFC`): **8 bytes** fixed | `[FC] [int64 milliseconds LE]` — total 9 bytes. Handler: [`rdb.c:4724`](src/rdb.c#L4724) |
| T3 | v9+ timestamps are architecture-portable | `rdbver >= 9`: proper LE via `memrev64ifbe()`. Source: [`rdb.c:159`](src/rdb.c#L159) |
| T4 | v1–v8 timestamps may be swapped on BE | `rdbLoadSignedInteger` did NOT call `memrev64ifbe()` for `rdbver < 9`. Source: [`rdb.c:140–161`](src/rdb.c#L140) |

## 6. Key-Value Record Invariants

Source: [`src/rdb.c:4945`](src/rdb.c#L4945) — `val = rdbLoadObject(type,rdb,key,...)`.
Rediscope: [`reader.go:180`](../internal/rdb/reader.go#L180) `readKeyRecord()`.

| # | Invariant | Detail |
|---|-----------|--------|
| K1 | Type byte is always exactly **1 byte** | Never length-encoded. Source: [`rdb.c:111`](src/rdb.c#L111) `rdbSaveType()` — single byte write |
| K2 | Key is always **string-encoded** | Uses standard RDB string encoding |
| K3 | Value format determined solely by type byte | No additional metadata needed. Source: [`rdb.c:2898`](src/rdb.c#L2898) `rdbLoadObject(int rdbtype, ...)` |
| K4 | Every KV has 3 logical byte regions | `[type: 1 byte]` + `[key: variable]` + `[value: variable]` |
| K5 | Type byte offset = record start | First byte of the KV record is the type byte |
| K6 | Key immediately follows type byte | No padding or alignment |

## 7. Type-Specific Value Invariants

### Single-Blob Types

These types encode their value as a single string-encoded blob. The blob length fully determines the byte boundary.

| Types | Name | Source (handler) |
|-------|------|------------------|
| 9 | HASH_ZIPMAP | [`rdb.c:3626`](src/rdb.c#L3626) |
| 10 | LIST_ZIPLIST | [`rdb.c:3683`](src/rdb.c#L3683) |
| 11 | SET_INTSET | [`rdb.c:3713`](src/rdb.c#L3713) |
| 12 | ZSET_ZIPLIST | [`rdb.c:3748`](src/rdb.c#L3748) |
| 13 | HASH_ZIPLIST | [`rdb.c:3794`](src/rdb.c#L3794) |
| 16 | HASH_LISTPACK | [`rdb.c:3821`](src/rdb.c#L3821) |
| 17 | ZSET_LISTPACK | [`rdb.c:3775`](src/rdb.c#L3775) |
| 20 | SET_LISTPACK | [`rdb.c:3727`](src/rdb.c#L3727) |
| 23 | HASH_LISTPACK_EX_PRE_GA | [`rdb.c:3822`](src/rdb.c#L3822) |

For type 25 (HASH_LISTPACK_EX), an 8-byte `minExpire` prefix precedes the blob. Source: [`rdb.c:3602`](src/rdb.c#L3602).

### Counted-Element Types

| Type | Format | Source |
|------|--------|--------|
| 1 (LIST) | `[count] [string]×count` | [`rdb.c:2922`](src/rdb.c#L2922) |
| 2 (SET) | `[count] [string]×count` | [`rdb.c:2943`](src/rdb.c#L2943) |
| 3 (ZSET) | `[count] ([string member] [string score])×count` | [`rdb.c:3043`](src/rdb.c#L3043) |
| 4 (HASH) | `[count] ([string field] [string value])×count` | [`rdb.c:3114`](src/rdb.c#L3114) |
| 5 (ZSET_2) | `[count] ([string member] [8-byte double])×count` | [`rdb.c:3072`](src/rdb.c#L3072) |

**ZSET_2 invariant:** Each score is always exactly 8 bytes (IEEE 754 double-precision).

### Quicklist Types

| Type | Format | Source |
|------|--------|--------|
| 14 (QUICKLIST) | `[node_count] [string blob]×node_count` | [`rdb.c:3513`](src/rdb.c#L3513) |
| 18 (QUICKLIST_2) | `[node_count] ([container_fmt] [string blob])×node_count` | [`rdb.c:3523`](src/rdb.c#L3523) |

**QUICKLIST_2 invariant:** `container_fmt` is length-encoded: `1` = ziplist, `2` = listpack. Source: [`rdb.c:3549`](src/rdb.c#L3549).

### Array Type

| Type | Format | Source |
|------|--------|--------|
| 28 (ARRAY) | `[count] ([tag] [payload])×count` | [`rdb.c:4393`](src/rdb.c#L4393) |

**Tag invariants:**

| Tag | Payload Size | Source |
|-----|-------------|--------|
| 0 (SDS) | string-encoded (variable) | Rediscope: [`reader.go:492`](../internal/rdb/reader.go#L492) `skipArray()` |
| 1 (int64) | exactly 8 bytes, FIXED | |
| 2 (double) | exactly 8 bytes, FIXED | |
| 3 (small_str) | `[1-byte length] [N bytes]`, variable | |

## 8. Rediscope Parser Invariants

These are invariants of the parser implementation itself.

| # | Invariant | Source |
|---|-----------|--------|
| P1 | Position counter is monotonically increasing | [`reader.go:36`](../internal/rdb/reader.go#L36) `readByte()`, [`reader.go:49`](../internal/rdb/reader.go#L49) `readLength()`, [`reader.go:149`](../internal/rdb/reader.go#L149) `skipBytes()` — all advance `r.pos` |
| P2 | `Record.Start < Record.End` | [`model.go`](../internal/rdb/model.go) — every record has non-zero byte span |
| P3 | KV records have exactly 3 parts | `{kind:"type"}`, `{kind:"key"}`, `{kind:"value"}` — [`parser.go:471–496`](../internal/rdb/parser.go#L471) |
| P4 | Type part is always 1 byte | `part.end - part.start == 1` for `kind:"type"` |
| P5 | Groups are ordered: metadata → KV → trailer | "File metadata", "Key value pairs", "Trailer" — [`parser.go:503–530`](../internal/rdb/parser.go#L503) |
| P6 | `model.Hex` is full-file hex (lowercase) | `hex.EncodeToString(data)` — [`parser.go:58`](../internal/rdb/parser.go#L58) |
| P7 | Section offsets are contiguous | [`parser.go:547`](../internal/rdb/parser.go#L547) `addSection()` |

## 9. What the Parser Does NOT Guarantee

| # | Non-Guarantee | Detail | Source |
|---|---------------|--------|--------|
| N1 | No value decoding | Values skipped for byte boundaries, not decoded | [`reader.go:216`](../internal/rdb/reader.go#L216) `skipValue()` |
| N2 | No LZF decompression | Returns `<lzf>` literal | [`reader.go:89`](../internal/rdb/reader.go#L89) `readString()` |
| N3 | No CRC64 verification | Checksum read but not validated | [`parser.go:411`](../internal/rdb/parser.go#L411) vs [`rdb.c:5049`](src/rdb.c#L5049) |
| N4 | No big-endian host support | Assumes LE architecture | [`rdb.c:159`](src/rdb.c#L159) — BE fix only for timestamps |
| N5 | No FUNCTION_PRE_GA tolerance | Opcode `0xF6` → hard failure | [`parser.go:318`](../internal/rdb/parser.go#L318) |
| N6 | No cross-file template state | Template IDs scoped to single parse | [`parser.go:396`](../internal/rdb/parser.go#L396) — `reader.templateFields` |
