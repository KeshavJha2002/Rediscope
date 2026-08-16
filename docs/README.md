# RDB File Format Documentation

`rediscope` v1:beta parses Redis RDB snapshot files and renders them as interactive, byte-level viewers. This directory contains the authoritative documentation for the RDB binary format as understood and handled by rediscope.

## Canonical Source of Truth

All specifications in this documentation are derived from the official Redis source code at commit [`cbdad795d`](https://github.com/redis/redis/commit/cbdad795d8d75746e501aae06f14a3398bd190a2) — specifically:

- [`src/rdb.h`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h) — Type, opcode, and encoding constant definitions
- [`src/rdb.c`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.c) — Serialization and deserialization implementation

Every architectural claim in these docs is annotated with `file:line @ cbdad795d` citations.

## RDB Version → Redis Version Mapping

Verified from the 9-byte magic headers of test fixture files:

| RDB Version | Redis Versions | Magic String | Verification Fixture |
|-------------|---------------|--------------|---------------------|
| **v9**  | 6.0, 6.2 LTS | `REDIS0009` | `redis-6.2.23-bulk.rdb` |
| **v10** | 7.0 LTS | `REDIS0010` | `redis-7.0.15-bulk.rdb` |
| **v11** | 7.2 LTS | `REDIS0011` | `redis-7.2.15-bulk.rdb` |
| **v12** | 7.4 LTS, 8.0, 8.2, 8.4 | `REDIS0012` | `redis-7.4.10-bulk.rdb`, `redis-8.0.6-bulk.rdb`, `redis-8.2.8-bulk.rdb`, `redis-8.4.5-bulk.rdb` |
| **v13** | 8.6 | `REDIS0013` | `redis-8.6.5-bulk.rdb` |
| **v14** | 8.8 | `REDIS0014` | `redis-8.8.1-bulk.rdb` |
| **v15** | 8.9+ (trunk) | `REDIS0015` | `native-types.rdb` (lab fixture) |

Current RDB version: `RDB_VERSION 15` — [`src/rdb.h:21`](file:///home/keshav2002latest/dev/redis_lab/redis/src/rdb.h#L21) @ `cbdad795d`

## Document Index

| Document | Description |
|----------|-------------|
| [`rdb-format.md`](rdb-format.md) | Common RDB binary format — invariant across all versions |
| [`rdb-v9.md`](rdb-v9.md) | RDB v9 — Redis 6.0 / 6.2 LTS |
| [`rdb-v10.md`](rdb-v10.md) | RDB v10 — Redis 7.0 LTS |
| [`rdb-v11.md`](rdb-v11.md) | RDB v11 — Redis 7.2 LTS |
| [`rdb-v12.md`](rdb-v12.md) | RDB v12 — Redis 7.4 LTS / 8.0 / 8.2 / 8.4 |
| [`rdb-v13.md`](rdb-v13.md) | RDB v13 — Redis 8.6 |
| [`rdb-v14.md`](rdb-v14.md) | RDB v14 — Redis 8.8 |
| [`compatibility-matrix.md`](compatibility-matrix.md) | What rediscope handles vs what the spec defines |
| [`invariants.md`](invariants.md) | Byte-level invariants and encoding rules |

## Test Fixtures

Test fixtures live at `test/rdb/` (relative to the repository root). For each Redis version, two fixture types are provided:

| Type | Size | Keys | Purpose |
|------|------|------|---------|
| **Bulk** | ~27 KB | 468 | Fast unit-test parsing across all type bytes |
| **Complex** | ~64 MB | ~54,000 | Stress testing with LZF, large streams, modules |

An additional lab fixture (`lab_artifacts/redis_persistence/native-types.rdb`, 483 bytes, RDB v15, 11 keys) serves as the primary unit test target.

## Reference Documentation

- [Redis Persistence](https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/) — Snapshotting lifecycle, AOF hybrid
- [RDB Internals / AUX Design](https://redis.io/docs/latest/operate/oss_and_stack/reference/internals/rdd/) — AUX field format, metadata keys
- [Modules API](https://redis.io/docs/latest/develop/reference/modules/modules-api-ref/) — Module type serialization
- [DUMP Command](https://redis.io/docs/latest/commands/dump/) — Single-key RDB payload wire format
- [Durable Redis](https://redis.io/technology/durable-redis/) — Persistence architecture overview
