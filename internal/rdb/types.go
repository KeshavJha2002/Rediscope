package rdb

const (
	opHashTemplate byte = 0xF2
	opKeyMeta      byte = 0xF3
	opSlotInfo     byte = 0xF4
	opFunction2    byte = 0xF5
	opFunction     byte = 0xF6
	opModuleAux    byte = 0xF7
	opIdle         byte = 0xF8
	opFreq         byte = 0xF9
	opAux          byte = 0xFA
	opResizeDB     byte = 0xFB
	opExpireMS     byte = 0xFC
	opExpireSec    byte = 0xFD
	opSelectDB     byte = 0xFE
	opEOF          byte = 0xFF
	lenEncInt8          = 0
	lenEncInt16         = 1
	lenEncInt32         = 2
	lenEncLZF           = 3

	moduleOpEOF    = 0
	moduleOpSInt   = 1
	moduleOpUInt   = 2
	moduleOpFloat  = 3
	moduleOpDouble = 4
	moduleOpString = 5

	arrayTagSDS      = 0
	arrayTagInt      = 1
	arrayTagFloat    = 2
	arrayTagSmallStr = 3
)

func TypeName(t byte) string {
	names := map[byte]string{
		0:  "RDB_TYPE_STRING",
		1:  "RDB_TYPE_LIST",
		2:  "RDB_TYPE_SET",
		3:  "RDB_TYPE_ZSET",
		4:  "RDB_TYPE_HASH",
		5:  "RDB_TYPE_ZSET_2",
		6:  "RDB_TYPE_MODULE_PRE_GA",
		7:  "RDB_TYPE_MODULE_2",
		9:  "RDB_TYPE_HASH_ZIPMAP",
		10: "RDB_TYPE_LIST_ZIPLIST",
		11: "RDB_TYPE_SET_INTSET",
		12: "RDB_TYPE_ZSET_ZIPLIST",
		13: "RDB_TYPE_HASH_ZIPLIST",
		14: "RDB_TYPE_LIST_QUICKLIST",
		15: "RDB_TYPE_STREAM_LISTPACKS",
		16: "RDB_TYPE_HASH_LISTPACK",
		17: "RDB_TYPE_ZSET_LISTPACK",
		18: "RDB_TYPE_LIST_QUICKLIST_2",
		19: "RDB_TYPE_STREAM_LISTPACKS_2",
		20: "RDB_TYPE_SET_LISTPACK",
		21: "RDB_TYPE_STREAM_LISTPACKS_3",
		22: "RDB_TYPE_HASH_METADATA_PRE_GA",
		23: "RDB_TYPE_HASH_LISTPACK_EX_PRE_GA",
		24: "RDB_TYPE_HASH_METADATA",
		25: "RDB_TYPE_HASH_LISTPACK_EX",
		26: "RDB_TYPE_STREAM_LISTPACKS_4",
		27: "RDB_TYPE_STREAM_LISTPACKS_5",
		28: "RDB_TYPE_ARRAY",
		29: "RDB_TYPE_HASH_TMPL_LP",
		30: "RDB_TYPE_HASH_TMPL_LP_REF",
		31: "RDB_TYPE_HASH_TMPL_ARRAY",
		32: "RDB_TYPE_HASH_TMPL_ARRAY_REF",
		33: "RDB_TYPE_GCRA",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "UNKNOWN"
}

func EncodingName(t byte) string {
	names := map[byte]string{
		0:  "string",
		1:  "list",
		2:  "set",
		3:  "zset",
		4:  "hash",
		5:  "zset",
		6:  "module-pre-ga",
		7:  "module",
		9:  "zipmap",
		10: "ziplist",
		11: "intset",
		12: "ziplist",
		13: "ziplist",
		14: "quicklist-ziplist",
		15: "stream-listpacks",
		16: "listpack",
		17: "listpack",
		18: "quicklist",
		19: "stream-listpacks",
		20: "listpack",
		21: "stream-listpacks",
		22: "hash-metadata",
		23: "listpack-ex",
		24: "hash-metadata",
		25: "listpack-ex",
		26: "stream-listpacks",
		27: "stream-listpacks",
		28: "array",
		29: "template-listpack",
		30: "template-listpack-ref",
		31: "template-array",
		32: "template-array-ref",
		33: "gcra",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "unknown"
}
