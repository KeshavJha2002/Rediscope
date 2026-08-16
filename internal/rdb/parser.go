package rdb

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(path string) (FileModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileModel{}, fmt.Errorf("failed to read RDB file %s: %w", path, err)
	}
	return p.Parse(path, data)
}

func (p *Parser) Parse(path string, data []byte) (model FileModel, err error) {
	reader := newReader(data)

	defer func() {
		if r := recover(); r != nil {
			err = &ParseError{
				Path:    path,
				Offset:  reader.pos,
				Message: fmt.Sprintf("panic recovered during RDB parsing: %v", r),
				Err:     ErrCorruptPayload,
			}
		}
	}()

	if len(data) < 9 {
		return FileModel{}, &ParseError{
			Path:    path,
			Offset:  len(data),
			Message: "file size is less than 9 bytes",
			Err:     ErrTooSmall,
		}
	}
	if string(data[:5]) != "REDIS" {
		magic := string(data[:min(len(data), 5)])
		return FileModel{}, &ParseError{
			Path:    path,
			Offset:  0,
			Message: fmt.Sprintf("invalid header magic %q (expected \"REDIS\")", magic),
			Err:     ErrInvalidSignature,
		}
	}

	rawVersion := string(data[5:9])
	verNum, _ := strconv.Atoi(strings.TrimLeft(rawVersion, "0"))
	versionLabel := fmt.Sprintf("RDB v%d", verNum)
	if verNum == 0 {
		versionLabel = fmt.Sprintf("RDB %s", rawVersion)
	}

	baseName := filepath.Base(path)
	fileID := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(baseName, ".", "-"), "_", "-"))

	maxHexBytes := len(data)
	if maxHexBytes > 65536 {
		maxHexBytes = 65536
	}

	model = FileModel{
		ID:         fileID,
		Name:       baseName,
		Path:       path,
		Bytes:      len(data),
		Size:       len(data),
		Version:    versionLabel,
		RawVersion: rawVersion,
		Hex:        hex.EncodeToString(data[:maxHexBytes]),
		Groups:     nil,
		Sections:   nil,
		Keys:       nil,
	}

	model.addSection("signature", "header", "signature REDIS", 0, 5, "")
	model.addSection("version", "header", "rdb version "+rawVersion, 5, 9, "")
	reader.pos = 9

	metaRecords := []Record{
		{
			ID:       "header",
			Start:    0,
			End:      9,
			Label:    "signature",
			Type:     "metadata",
			Color:    "var(--header)",
			Value:    fmt.Sprintf("signature=\"REDIS%s\"", rawVersion),
			JSON:     fmt.Sprintf("{\"signature\": \"REDIS%s\"}", rawVersion),
			Encoding: "ascii",
			Size:     "9B",
			Summary:  fmt.Sprintf("REDIS%s file signature and RDB version.", rawVersion),
		},
	}

	db := 0
	var pendingExpiry *int64
	var trailerRecords []Record

	for reader.pos < len(data) {
		start := reader.pos
		op, err := reader.readByte()
		if err != nil {
			return FileModel{}, err
		}

		switch op {
		case opIdle:
			idle, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("idle", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "idle",
				Type:     "metadata",
				Color:    "var(--aux)",
				Value:    fmt.Sprintf("idle=%d", idle),
				JSON:     fmt.Sprintf("{\"idle\": %d}", idle),
				Encoding: "opcode+len",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  fmt.Sprintf("LRU idle time: %d seconds.", idle),
			})
			model.addSection(sectionID("idle", len(model.Sections)), "metadata", fmt.Sprintf("idle=%d", idle), start, reader.pos, fmt.Sprintf("%d", idle))

		case opFreq:
			freq, err := reader.readByte()
			if err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("freq", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "freq",
				Type:     "metadata",
				Color:    "var(--aux)",
				Value:    fmt.Sprintf("freq=%d", freq),
				JSON:     fmt.Sprintf("{\"freq\": %d}", freq),
				Encoding: "opcode+byte",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  fmt.Sprintf("LFU access frequency counter: %d.", freq),
			})
			model.addSection(sectionID("freq", len(model.Sections)), "metadata", fmt.Sprintf("freq=%d", freq), start, reader.pos, fmt.Sprintf("%d", freq))

		case opAux:
			key, _, _, err := reader.readString()
			if err != nil {
				return FileModel{}, err
			}
			value, _, _, err := reader.readString()
			if err != nil {
				return FileModel{}, err
			}

			enc := "rdb-string"
			summary := fmt.Sprintf("AUX field: %s = %s", key, value)
			switch key {
			case "redis-ver":
				summary = "RDB producer version."
			case "redis-bits":
				enc = "int8"
				summary = "Redis architecture bitness."
			case "ctime":
				enc = "int32"
				summary = "Creation time stored in the RDB AUX section."
			case "used-mem":
				enc = "int32"
				summary = "Memory usage value stored in the RDB AUX section."
			case "aof-base":
				enc = "int8"
				if value == "0" {
					summary = "This file is not marked as an AOF base."
				} else {
					summary = "This file is marked as an AOF base file."
				}
			}

			metaRecords = append(metaRecords, Record{
				ID:       sectionID("aux", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    key,
				Type:     "metadata",
				Color:    "var(--aux)",
				Value:    fmt.Sprintf("%s=%s", key, quoteValue(value)),
				JSON:     fmt.Sprintf("{\"%s\": %s}", key, quoteValue(value)),
				Encoding: enc,
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  summary,
			})
			model.addSection(sectionID("aux", len(model.Sections)), "aux", "aux "+key+"="+value, start, reader.pos, key+"="+value)

		case opSelectDB:
			selected, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			db = int(selected)
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("selectdb", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "selected-db",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("selected-db=%d", db),
				JSON:     fmt.Sprintf("{\"selected-db\": %d}", db),
				Encoding: "opcode+len",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  fmt.Sprintf("Selects logical database %d for following key records.", db),
			})
			model.addSection(sectionID("selectdb", len(model.Sections)), "database", fmt.Sprintf("db=%d", db), start, reader.pos, "")

		case opResizeDB:
			dbSize, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			expireSize, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("resizedb", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "resize-hint",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("db-size=%d, expires-size=%d", dbSize, expireSize),
				JSON:     fmt.Sprintf("{\"db-size\": %d, \"expires-size\": %d}", dbSize, expireSize),
				Encoding: "opcode+len",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  "Hash table size hints for fast loading.",
			})
			model.addSection(sectionID("resizedb", len(model.Sections)), "database", fmt.Sprintf("resize db=%d expires=%d", dbSize, expireSize), start, reader.pos, "")

		case opSlotInfo:
			slotID, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			slotSize, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			expiresSize, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("slot", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "slot-info",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("slot-id=%d, keys=%d, expires=%d", slotID, slotSize, expiresSize),
				JSON:     fmt.Sprintf("{\"slot-id\": %d, \"keys\": %d, \"expires\": %d}", slotID, slotSize, expiresSize),
				Encoding: "opcode+len",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  fmt.Sprintf("Cluster slot metadata for slot %d.", slotID),
			})
			model.addSection(sectionID("slot", len(model.Sections)), "database", fmt.Sprintf("slot-id=%d keys=%d expires=%d", slotID, slotSize, expiresSize), start, reader.pos, "")

		case opExpireMS:
			if reader.remaining() < 8 {
				return FileModel{}, reader.err("truncated millisecond expiry", ErrTruncated)
			}
			v := int64(binary.LittleEndian.Uint64(data[reader.pos : reader.pos+8]))
			reader.pos += 8
			pendingExpiry = &v
			model.addSection(sectionID("expire", len(model.Sections)), "expiry", "expire-ms", start, reader.pos, fmt.Sprintf("%d", v))

		case opExpireSec:
			if reader.remaining() < 4 {
				return FileModel{}, reader.err("truncated second expiry", ErrTruncated)
			}
			sec := int64(binary.LittleEndian.Uint32(data[reader.pos : reader.pos+4]))
			ms := sec * 1000
			reader.pos += 4
			pendingExpiry = &ms
			model.addSection(sectionID("expire", len(model.Sections)), "expiry", "expire-sec", start, reader.pos, fmt.Sprintf("%d", sec))

		case opFunction2:
			if _, _, err := reader.skipString(); err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("function", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "function",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    "function=library",
				Encoding: "function",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  "Redis function library definition.",
			})
			model.addSection(sectionID("function", len(model.Sections)), "function", "function=library", start, reader.pos, "")

		case opFunction:
			return FileModel{}, &ParseError{
				Path:    path,
				Offset:  start,
				Opcode:  opFunction,
				Message: "pre-GA function opcode 0xF6 is not supported by Redis trunk",
				Err:     ErrUnsupportedOpcode,
			}

		case opModuleAux:
			moduleID, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			whenOpcode, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			when, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			if whenOpcode != moduleOpUInt {
				return FileModel{}, reader.err("module aux when opcode is not UINT", ErrCorruptPayload)
			}
			if err := reader.skipModulePayload(); err != nil {
				return FileModel{}, err
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("module-aux", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "module-aux",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("module-id=%d, when=%d", moduleID, when),
				Encoding: "module",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  "Module auxiliary state.",
			})
			model.addSection(sectionID("module-aux", len(model.Sections)), "module", fmt.Sprintf("module-id=%d when=%d", moduleID, when), start, reader.pos, "")

		case opKeyMeta:
			count, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			for i := uint64(0); i < count; i++ {
				if err := reader.skipBytes(4); err != nil {
					return FileModel{}, err
				}
				if err := reader.skipModulePayload(); err != nil {
					return FileModel{}, err
				}
			}
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("key-meta", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "key-meta",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("key-meta-classes=%d", count),
				Encoding: "metadata",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  "Key metadata definitions.",
			})
			model.addSection(sectionID("key-meta", len(model.Sections)), "metadata", fmt.Sprintf("key-meta-classes=%d", count), start, reader.pos, "")

		case opHashTemplate:
			templateID, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			count, err := reader.skipTemplateFields()
			if err != nil {
				return FileModel{}, err
			}
			reader.templateFields[templateID] = count
			metaRecords = append(metaRecords, Record{
				ID:       sectionID("hash-template", len(metaRecords)),
				Start:    start,
				End:      reader.pos,
				Label:    "hash-template",
				Type:     "metadata",
				Color:    "var(--db)",
				Value:    fmt.Sprintf("hash-template-id=%d, fields=%d", templateID, count),
				Encoding: "template",
				Size:     fmt.Sprintf("%dB", reader.pos-start),
				Summary:  "Hash field schema template for listpack compression.",
			})
			model.addSection(sectionID("hash-template", len(model.Sections)), "metadata", fmt.Sprintf("hash-template-id=%d fields=%d", templateID, count), start, reader.pos, "")

		case opEOF:
			trailerRecords = append(trailerRecords, Record{
				ID:       "eof",
				Start:    start,
				End:      reader.pos,
				Label:    "eof",
				Type:     "trailer",
				Color:    "var(--eof)",
				Value:    "eof=0xff",
				JSON:     "{\"eof\": \"0xff\"}",
				Encoding: "opcode",
				Size:     "1B",
				Summary:  "End of RDB stream.",
			})
			model.addSection("eof", "footer", "eof", start, reader.pos, "")

			if reader.remaining() >= 8 {
				checksumStart := reader.pos
				ckBytes := data[reader.pos : reader.pos+8]
				reader.pos += 8
				ckHex := hex.EncodeToString(ckBytes)
				trailerRecords = append(trailerRecords, Record{
					ID:       "checksum",
					Start:    checksumStart,
					End:      reader.pos,
					Label:    "checksum",
					Type:     "trailer",
					Color:    "var(--checksum)",
					Value:    fmt.Sprintf("checksum=%s", ckHex),
					JSON:     fmt.Sprintf("{\"checksum\": \"%s\"}", ckHex),
					Encoding: "crc64",
					Size:     "8B",
					Summary:  "CRC64 checksum bytes at the end of the file.",
				})
				model.addSection("checksum", "footer", "checksum", checksumStart, reader.pos, "")
			}
			break

		default:
			record, err := reader.readKeyRecord(db, op, start, pendingExpiry)
			if err != nil {
				return FileModel{}, err
			}
			pendingExpiry = nil
			keyIndex := len(model.Keys)
			model.Keys = append(model.Keys, record)
			model.addSection(sectionID("record", keyIndex), "record", record.Key+" "+record.TypeName, record.RecordStart, record.RecordEnd, "", keyIndex)
			model.addSection(sectionID("key", keyIndex), "key", "key "+record.Key, record.KeyStart, record.KeyEnd, record.Key, keyIndex)
			model.addSection(sectionID("value", keyIndex), "value", "value "+record.Key, record.ValueStart, record.ValueEnd, record.TypeName, keyIndex)
		}
	}

	maxKeysToRecord := len(model.Keys)
	if maxKeysToRecord > 1000 {
		maxKeysToRecord = 1000
	}

	keyRecords := make([]Record, 0, maxKeysToRecord)
	for i := 0; i < maxKeysToRecord; i++ {
		k := model.Keys[i]
		genType := GeneralType(k.TypeByte, k.Key)
		color := TypeColor(k.TypeByte, k.Key)
		sizeStr := fmt.Sprintf("%dB", k.RecordEnd-k.RecordStart)
		summary := fmt.Sprintf("%s stored with %s encoding.", k.TypeName, k.Encoding)
		valStr := fmt.Sprintf("{\"%s\": ...}", k.Key)

		var strSegments []RecordString
		keySegments := extractStringSegments(data, k.KeyStart, k.KeyEnd, "key")
		if len(keySegments) > 0 {
			strSegments = append(strSegments, keySegments...)
		} else {
			strSegments = append(strSegments, RecordString{Kind: "key", Start: k.KeyStart, End: k.KeyEnd, Text: k.Key})
		}

		valueSegments := extractStringSegments(data, k.ValueStart, k.ValueEnd, "value")
		if len(valueSegments) > 0 {
			strSegments = append(strSegments, valueSegments...)
		}

		record := Record{
			ID:       fmt.Sprintf("k-%d", i),
			Start:    k.RecordStart,
			End:      k.RecordEnd,
			Label:    k.Key,
			Type:     genType,
			RDBType:  k.TypeName,
			Color:    color,
			Value:    valStr,
			Encoding: k.Encoding,
			Size:     sizeStr,
			Summary:  summary,
			Parts: []RecordPart{
				{Kind: "type", Start: k.RecordStart, End: k.KeyStart},
				{Kind: "key", Start: k.KeyStart, End: k.KeyEnd},
				{Kind: "value", Start: k.ValueStart, End: k.ValueEnd},
			},
			Strings: strSegments,
		}
		keyRecords = append(keyRecords, record)
	}

	model.Groups = []RecordGroup{
		{
			Title:   "File metadata",
			Records: metaRecords,
		},
	}
	if len(keyRecords) > 0 {
		model.Groups = append(model.Groups, RecordGroup{
			Title:   "Key value pairs",
			Records: keyRecords,
		})
	}
	if len(trailerRecords) > 0 {
		model.Groups = append(model.Groups, RecordGroup{
			Title:   "Trailer",
			Records: trailerRecords,
		})
	}

	if len(model.Keys) == 1 {
		model.CountLabel = "1 key"
	} else {
		model.CountLabel = fmt.Sprintf("%d keys", len(model.Keys))
	}

	return model, nil
}

func quoteValue(v string) string {
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return v
	}
	return fmt.Sprintf("%q", v)
}

func (m *FileModel) addSection(id, kind, label string, start, end int, meta string, keyIndex ...int) {
	section := Section{
		ID:    id,
		Kind:  kind,
		Label: label,
		Start: start,
		End:   end,
		Size:  end - start,
		Meta:  meta,
	}
	if len(keyIndex) > 0 {
		section.KeyIndex = keyIndex[0]
	}
	m.Sections = append(m.Sections, section)
}

func sectionID(prefix string, index int) string {
	return fmt.Sprintf("%s-%d", prefix, index)
}

func extractStringSegments(data []byte, start, end int, kind string) []RecordString {
	if start >= end || start < 0 || end > len(data) {
		return nil
	}
	var segments []RecordString
	inRun := false
	runStart := start

	for i := start; i < end; i++ {
		b := data[i]
		if b >= 32 && b <= 126 { // printable ascii
			if !inRun {
				inRun = true
				runStart = i
			}
		} else {
			if inRun {
				if i-runStart >= 1 {
					segments = append(segments, RecordString{
						Kind:  kind,
						Start: runStart,
						End:   i,
						Text:  string(data[runStart:i]),
					})
				}
				inRun = false
			}
		}
	}
	if inRun && end-runStart >= 1 {
		segments = append(segments, RecordString{
			Kind:  kind,
			Start: runStart,
			End:   end,
			Text:  string(data[runStart:end]),
		})
	}
	return segments
}
