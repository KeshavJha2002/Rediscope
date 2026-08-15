package rdb

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(path string) (FileModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileModel{}, err
	}
	model, err := p.Parse(path, data)
	if err != nil {
		return FileModel{}, err
	}
	return model, nil
}

func (p *Parser) Parse(path string, data []byte) (FileModel, error) {
	if len(data) < 10 {
		return FileModel{}, fmt.Errorf("%s: too small to be an RDB file", path)
	}
	if string(data[:5]) != "REDIS" {
		return FileModel{}, fmt.Errorf("%s: missing REDIS signature", path)
	}

	reader := newReader(data)
	model := FileModel{
		Name:    filepath.Base(path),
		Path:    path,
		Size:    len(data),
		Version: string(data[5:9]),
		Hex:     hex.EncodeToString(data),
	}
	model.addSection("signature", "header", "signature REDIS", 0, 5, "")
	model.addSection("version", "header", "rdb version "+model.Version, 5, 9, "")
	reader.pos = 9

	db := 0
	var pendingExpiry *int64
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
			model.addSection(sectionID("idle", len(model.Sections)), "metadata", fmt.Sprintf("idle=%d", idle), start, reader.pos, fmt.Sprintf("%d", idle))
		case opFreq:
			freq, err := reader.readByte()
			if err != nil {
				return FileModel{}, err
			}
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
			model.addSection(sectionID("aux", len(model.Sections)), "aux", "aux "+key+"="+value, start, reader.pos, key+"="+value)
		case opSelectDB:
			selected, _, err := reader.readLength()
			if err != nil {
				return FileModel{}, err
			}
			db = int(selected)
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
			model.addSection(sectionID("slot", len(model.Sections)), "database", fmt.Sprintf("slot-id=%d keys=%d expires=%d", slotID, slotSize, expiresSize), start, reader.pos, "")
		case opExpireMS:
			if reader.remaining() < 8 {
				return FileModel{}, reader.err("truncated millisecond expiry")
			}
			v := int64(binary.LittleEndian.Uint64(data[reader.pos : reader.pos+8]))
			reader.pos += 8
			pendingExpiry = &v
			model.addSection(sectionID("expire", len(model.Sections)), "expiry", "expire-ms", start, reader.pos, fmt.Sprintf("%d", v))
		case opExpireSec:
			if reader.remaining() < 4 {
				return FileModel{}, reader.err("truncated second expiry")
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
			model.addSection(sectionID("function", len(model.Sections)), "function", "function=library", start, reader.pos, "")
		case opFunction:
			return FileModel{}, reader.err("pre-GA function opcode is not supported by Redis trunk")
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
				return FileModel{}, reader.err("module aux when opcode is not UINT")
			}
			if err := reader.skipModulePayload(); err != nil {
				return FileModel{}, err
			}
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
			model.addSection(sectionID("hash-template", len(model.Sections)), "metadata", fmt.Sprintf("hash-template-id=%d fields=%d", templateID, count), start, reader.pos, "")
		case opEOF:
			model.addSection("eof", "footer", "eof", start, reader.pos, "")
			if reader.remaining() >= 8 {
				checksumStart := reader.pos
				reader.pos += 8
				model.addSection("checksum", "footer", "checksum", checksumStart, reader.pos, "")
			}
			return model, nil
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

	return model, nil
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
