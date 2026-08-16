# RDB Version 14 — Redis 8.8

**Magic:** `REDIS0014`
**Verified fixture:** `redis-8.8.1-bulk.rdb` (26,976 bytes), `redis-8.8.1-redis-tests-complex.rdb` (64,658,953 bytes)

> All Redis source references cite commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2). Rediscope references cite the current `rediscope/` tree.

---

## What's New in v14 (vs v13)

### New Type Bytes

| Type | Hex | Name | Description | Source (define) | Source (handler) |
|------|-----|------|-------------|-----------------|------------------|
| 29 | `0x1D` | HASH_TMPL_LP | Template hash, listpack, self-contained (DUMP) | [`rdb.h:84`](src/rdb.h#L84) | [`rdb.c:3253`](src/rdb.c#L3253) |
| 30 | `0x1E` | HASH_TMPL_LP_REF | Template hash, listpack, template reference (RDB save) | [`rdb.h:85`](src/rdb.h#L85) | [`rdb.c:3273`](src/rdb.c#L3273) |
| 31 | `0x1F` | HASH_TMPL_ARRAY | Template hash, array, self-contained (DUMP) | [`rdb.h:86`](src/rdb.h#L86) | [`rdb.c:3303`](src/rdb.c#L3303) |
| 32 | `0x20` | HASH_TMPL_ARRAY_REF | Template hash, array, template reference (RDB save) | [`rdb.h:87`](src/rdb.h#L87) | [`rdb.c:3321`](src/rdb.c#L3321) |

### New Opcodes

| Opcode | Hex | Description | Source (define) | Source (handler) |
|--------|-----|-------------|-----------------|------------------|
| HASH_TEMPLATE | `0xF2` | Defines a hash template record (field name schema) | [`rdb.h:101`](src/rdb.h#L101) | [`rdb.c:4911`](src/rdb.c#L4911) |

## Hash Templates Architecture

Source: [`src/rdb.c:1886`](src/rdb.c#L1886) — *"Each template is written as its own RDB_OPCODE_HASH_TEMPLATE record"*; [`src/rdb.c:2615`](src/rdb.c#L2615) — template section description.

Hash templates deduplicate field names across many hashes sharing the same schema. A single `HASH_TEMPLATE` opcode defines the field names, and subsequent keys use `_REF` type bytes to reference the template by ID.

### HASH_TEMPLATE opcode (`0xF2`)

Source: [`src/rdb.c:4911`](src/rdb.c#L4911) (handler)

```
[0xF2]
[template_id: length-encoded]
[field_count: length-encoded]
  [field_name_0: string-encoded]
  ...
  [field_name_N-1: string-encoded]
```

### Self-Contained Types (for DUMP payloads)

#### HASH_TMPL_LP (type 29)

Source: [`src/rdb.c:3253`](src/rdb.c#L3253); inline comment at [`rdb.h:84`](src/rdb.h#L84): `[count][f0]...[fN-1][lp_blob]`

```
[field_count: length-encoded]
[field_name_0: string-encoded]
...
[field_name_N-1: string-encoded]
[listpack_blob: string-encoded]
```

#### HASH_TMPL_ARRAY (type 31)

Source: [`src/rdb.c:3303`](src/rdb.c#L3303); inline comment at [`rdb.h:86`](src/rdb.h#L86): `[count][f0][v0]...[fN-1][vN-1]`

```
[field_count: length-encoded]
  for each field:
    [field_name: string-encoded]
    [field_value: string-encoded]
```

### Reference Types (for RDB save)

#### HASH_TMPL_LP_REF (type 30)

Source: [`src/rdb.c:3273`](src/rdb.c#L3273); inline comment at [`rdb.h:85`](src/rdb.h#L85): `raw lp blob, first entry is tid`

```
[listpack_blob: string-encoded]     ← first entry in the listpack is template ID
```

#### HASH_TMPL_ARRAY_REF (type 32)

Source: [`src/rdb.c:3321`](src/rdb.c#L3321); inline comment at [`rdb.h:87`](src/rdb.h#L87): `[tid][v0]...[vN-1]`

```
[template_id: length-encoded]
[value_0: string-encoded]
...
[value_N-1: string-encoded]
```

## All Types Available in v14

Types 0–7, 9–32. Plus type 33 (GCRA) if compiled with `ENABLE_GCRA` — [`rdb.h:89`](src/rdb.h#L89).

## All Opcodes Available in v14

`0xF2`–`0xFF` (the full opcode set).

## Rediscope v1:beta Handling

| Feature | Status | Rediscope Source |
|---------|--------|------------------|
| Types 29–32 registered | ✅ | [`types.go:66–69`](../internal/rdb/types.go#L66) |
| HASH_TEMPLATE opcode | ✅ | [`parser.go:387–409`](../internal/rdb/parser.go#L387) — registers in `reader.templateFields` at [`parser.go:396`](../internal/rdb/parser.go#L396) |
| Self-contained (29, 31) | ✅ | [`reader.go:536`](../internal/rdb/reader.go#L536) `skipTemplateFields()` |
| Reference types (30, 32) | ✅ | [`reader.go:549`](../internal/rdb/reader.go#L549) `skipTemplateArray()` uses `reader.templateFields` |
| GCRA (33) registered | ✅ | [`types.go:70`](../internal/rdb/types.go#L70) |

### ⚠️ Template Dependency

Reference types (30, 32) require the `HASH_TEMPLATE` opcode to have been parsed earlier in the same RDB file. Redis always emits templates before references (source: [`src/rdb.c:2615`](src/rdb.c#L2615)), so this is safe for valid RDB files.

## Test Coverage

| Fixture | Keys | Size | Automated Test |
|---------|------|------|----------------|
| `redis-8.8.1-bulk.rdb` | 468 | 27 KB | ✅ |
| `redis-8.8.1-redis-tests-complex.rdb` | 53,930 | 64 MB | ❌ |
