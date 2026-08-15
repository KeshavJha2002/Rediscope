package rdb

import (
	"encoding/binary"
	"fmt"
)

type reader struct {
	data           []byte
	pos            int
	templateFields map[uint64]uint64
}

func newReader(data []byte) *reader {
	return &reader{data: data, templateFields: map[uint64]uint64{}}
}

func (r *reader) remaining() int {
	return len(r.data) - r.pos
}

func (r *reader) err(message string) error {
	return fmt.Errorf("rdb offset %d: %s", r.pos, message)
}

func (r *reader) readByte() (byte, error) {
	if r.remaining() < 1 {
		return 0, r.err("unexpected eof")
	}
	v := r.data[r.pos]
	r.pos++
	return v, nil
}

func (r *reader) readLength() (uint64, bool, error) {
	first, err := r.readByte()
	if err != nil {
		return 0, false, err
	}
	mode := (first & 0xC0) >> 6
	switch mode {
	case 0:
		return uint64(first & 0x3F), false, nil
	case 1:
		next, err := r.readByte()
		if err != nil {
			return 0, false, err
		}
		return uint64(first&0x3F)<<8 | uint64(next), false, nil
	case 2:
		if first&0x3F == 0 {
			if r.remaining() < 4 {
				return 0, false, r.err("truncated 32-bit length")
			}
			v := binary.BigEndian.Uint32(r.data[r.pos : r.pos+4])
			r.pos += 4
			return uint64(v), false, nil
		}
		if first&0x3F == 1 {
			if r.remaining() < 8 {
				return 0, false, r.err("truncated 64-bit length")
			}
			v := binary.BigEndian.Uint64(r.data[r.pos : r.pos+8])
			r.pos += 8
			return v, false, nil
		}
		return 0, false, r.err("unknown extended length encoding")
	case 3:
		return uint64(first & 0x3F), true, nil
	default:
		return 0, false, r.err("invalid length")
	}
}

func (r *reader) readString() (string, int, int, error) {
	start := r.pos
	length, encoded, err := r.readLength()
	if err != nil {
		return "", start, r.pos, err
	}
	if !encoded {
		if length > uint64(r.remaining()) {
			return "", start, r.pos, r.err("truncated string")
		}
		bodyStart := r.pos
		r.pos += int(length)
		return string(r.data[bodyStart:r.pos]), start, r.pos, nil
	}

	switch length {
	case lenEncInt8:
		v, err := r.readByte()
		if err != nil {
			return "", start, r.pos, err
		}
		return fmt.Sprintf("%d", int8(v)), start, r.pos, nil
	case lenEncInt16:
		if r.remaining() < 2 {
			return "", start, r.pos, r.err("truncated int16 string")
		}
		v := int16(binary.LittleEndian.Uint16(r.data[r.pos : r.pos+2]))
		r.pos += 2
		return fmt.Sprintf("%d", v), start, r.pos, nil
	case lenEncInt32:
		if r.remaining() < 4 {
			return "", start, r.pos, r.err("truncated int32 string")
		}
		v := int32(binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4]))
		r.pos += 4
		return fmt.Sprintf("%d", v), start, r.pos, nil
	case lenEncLZF:
		compressedLen, _, err := r.readLength()
		if err != nil {
			return "", start, r.pos, err
		}
		_, _, err = r.readLength()
		if err != nil {
			return "", start, r.pos, err
		}
		if compressedLen > uint64(r.remaining()) {
			return "", start, r.pos, r.err("truncated lzf string")
		}
		r.pos += int(compressedLen)
		return "<lzf>", start, r.pos, nil
	default:
		return "", start, r.pos, r.err("unknown encoded string")
	}
}

func (r *reader) skipString() (int, int, error) {
	_, start, end, err := r.readString()
	return start, end, err
}

func (r *reader) skipBytes(n int) error {
	if n < 0 || n > r.remaining() {
		return r.err("truncated raw bytes")
	}
	r.pos += n
	return nil
}

func (r *reader) skipDoubleText() error {
	marker, err := r.readByte()
	if err != nil {
		return err
	}
	if marker == 253 || marker == 254 || marker == 255 {
		return nil
	}
	return r.skipBytes(int(marker))
}

func (r *reader) skipBinaryDouble() error {
	return r.skipBytes(8)
}

func (r *reader) skipBinaryFloat() error {
	return r.skipBytes(4)
}

func (r *reader) skipSignedInteger() error {
	return r.skipBytes(8)
}

func (r *reader) readKeyRecord(db int, typ byte, recordStart int, expiry *int64) (KeyRecord, error) {
	key, keyStart, keyEnd, err := r.readString()
	if err != nil {
		return KeyRecord{}, err
	}
	valueStart := r.pos
	if err := r.skipValue(typ); err != nil {
		return KeyRecord{}, fmt.Errorf("key %q type %s at offset %d: %w", key, TypeName(typ), valueStart, err)
	}
	valueEnd := r.pos
	return KeyRecord{
		DB:          db,
		Key:         key,
		TypeByte:    typ,
		TypeName:    TypeName(typ),
		Encoding:    EncodingName(typ),
		ExpiryMS:    expiry,
		RecordStart: recordStart,
		RecordEnd:   valueEnd,
		KeyStart:    keyStart,
		KeyEnd:      keyEnd,
		ValueStart:  valueStart,
		ValueEnd:    valueEnd,
	}, nil
}

func (r *reader) skipValue(typ byte) error {
	switch typ {
	case 0, 9, 10, 11, 12, 13, 16, 17, 20:
		_, _, err := r.skipString()
		return err
	case 1, 2, 3, 4, 5:
		count, _, err := r.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < count; i++ {
			if _, _, err := r.skipString(); err != nil {
				return err
			}
			switch typ {
			case 3:
				if err := r.skipDoubleText(); err != nil {
					return err
				}
			case 4:
				if _, _, err := r.skipString(); err != nil {
					return err
				}
			case 5:
				if err := r.skipBinaryDouble(); err != nil {
					return err
				}
			}
		}
		return nil
	case 14:
		nodeCount, _, err := r.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < nodeCount; i++ {
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		}
		return nil
	case 18:
		nodeCount, _, err := r.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < nodeCount; i++ {
			if _, _, err := r.readLength(); err != nil {
				return err
			}
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		}
		return nil
	case 15, 19, 21, 26, 27:
		return r.skipStream(typ)
	case 22, 24:
		if typ == 24 {
			if err := r.skipBytes(8); err != nil {
				return err
			}
		}
		count, _, err := r.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < count; i++ {
			if _, _, err := r.readLength(); err != nil {
				return err
			}
			if _, _, err := r.skipString(); err != nil {
				return err
			}
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		}
		return nil
	case 23, 25:
		if typ == 25 {
			if err := r.skipBytes(8); err != nil {
				return err
			}
		}
		_, _, err := r.skipString()
		return err
	case 28:
		return r.skipArray()
	case 29:
		if _, err := r.skipTemplateFields(); err != nil {
			return err
		}
		_, _, err := r.skipString()
		return err
	case 30:
		_, _, err := r.skipString()
		return err
	case 31:
		return r.skipTemplateArray(false)
	case 32:
		return r.skipTemplateArray(true)
	case 7:
		if _, _, err := r.readLength(); err != nil {
			return err
		}
		return r.skipModulePayload()
	case 33:
		_, _, err := r.readLength()
		return err
	default:
		return r.err(fmt.Sprintf("unsupported RDB value type 0x%02x", typ))
	}
}

func (r *reader) skipStream(typ byte) error {
	listpackCount, _, err := r.readLength()
	if err != nil {
		return err
	}
	for i := uint64(0); i < listpackCount; i++ {
		if _, _, err := r.skipString(); err != nil {
			return err
		}
		if _, _, err := r.skipString(); err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		if _, _, err := r.readLength(); err != nil {
			return err
		}
	}
	if typ >= 19 {
		for i := 0; i < 5; i++ {
			if _, _, err := r.readLength(); err != nil {
				return err
			}
		}
	}
	cgroups, _, err := r.readLength()
	if err != nil {
		return err
	}
	for i := uint64(0); i < cgroups; i++ {
		if err := r.skipStreamConsumerGroup(typ); err != nil {
			return err
		}
	}
	if typ >= 26 {
		if _, _, err := r.readLength(); err != nil {
			return err
		}
		if _, _, err := r.readLength(); err != nil {
			return err
		}
		if err := r.skipStreamIDMPEntries(); err != nil {
			return err
		}
		if _, _, err := r.readLength(); err != nil {
			return err
		}
		if _, _, err := r.readLength(); err != nil {
			return err
		}
	}
	return nil
}

func (r *reader) skipStreamConsumerGroup(typ byte) error {
	if _, _, err := r.skipString(); err != nil {
		return err
	}
	if _, _, err := r.readLength(); err != nil {
		return err
	}
	if _, _, err := r.readLength(); err != nil {
		return err
	}
	if typ >= 19 {
		if _, _, err := r.readLength(); err != nil {
			return err
		}
	}
	pelSize, _, err := r.readLength()
	if err != nil {
		return err
	}
	for i := uint64(0); i < pelSize; i++ {
		if err := r.skipBytes(16); err != nil {
			return err
		}
		if err := r.skipBytes(8); err != nil {
			return err
		}
		if _, _, err := r.readLength(); err != nil {
			return err
		}
	}
	consumers, _, err := r.readLength()
	if err != nil {
		return err
	}
	for i := uint64(0); i < consumers; i++ {
		if _, _, err := r.skipString(); err != nil {
			return err
		}
		if err := r.skipBytes(8); err != nil {
			return err
		}
		if typ >= 21 {
			if err := r.skipBytes(8); err != nil {
				return err
			}
		}
		consumerPEL, _, err := r.readLength()
		if err != nil {
			return err
		}
		for j := uint64(0); j < consumerPEL; j++ {
			if err := r.skipBytes(16); err != nil {
				return err
			}
		}
	}
	if typ >= 27 {
		nacked, _, err := r.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < nacked; i++ {
			if err := r.skipBytes(16); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *reader) skipStreamIDMPEntries() error {
	producers, _, err := r.readLength()
	if err != nil {
		return err
	}
	for i := uint64(0); i < producers; i++ {
		if _, _, err := r.skipString(); err != nil {
			return err
		}
		count, _, err := r.readLength()
		if err != nil {
			return err
		}
		for j := uint64(0); j < count; j++ {
			if _, _, err := r.skipString(); err != nil {
				return err
			}
			if _, _, err := r.readLength(); err != nil {
				return err
			}
			if _, _, err := r.readLength(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *reader) skipArray() error {
	count, _, err := r.readLength()
	if err != nil {
		return err
	}
	insertFlag, _, err := r.readLength()
	if err != nil {
		return err
	}
	if insertFlag == 1 {
		if _, _, err := r.readLength(); err != nil {
			return err
		}
	} else if insertFlag != 0 {
		return r.err("invalid array insert index flag")
	}
	for i := uint64(0); i < count; i++ {
		if _, _, err := r.readLength(); err != nil {
			return err
		}
		tag, _, err := r.readLength()
		if err != nil {
			return err
		}
		switch tag {
		case arrayTagSDS, arrayTagSmallStr:
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		case arrayTagInt:
			if err := r.skipSignedInteger(); err != nil {
				return err
			}
		case arrayTagFloat:
			if err := r.skipBinaryDouble(); err != nil {
				return err
			}
		default:
			return r.err(fmt.Sprintf("unknown array element tag %d", tag))
		}
	}
	return nil
}

func (r *reader) skipTemplateFields() (uint64, error) {
	count, _, err := r.readLength()
	if err != nil {
		return 0, err
	}
	for i := uint64(0); i < count; i++ {
		if _, _, err := r.skipString(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (r *reader) skipTemplateArray(ref bool) error {
	if ref {
		templateID, _, err := r.readLength()
		if err != nil {
			return err
		}
		count, ok := r.templateFields[templateID]
		if !ok {
			return r.err(fmt.Sprintf("template-array-ref missing template id %d", templateID))
		}
		for i := uint64(0); i < count; i++ {
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		}
		return nil
	}
	count, err := r.skipTemplateFields()
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		if _, _, err := r.skipString(); err != nil {
			return err
		}
	}
	return nil
}

func (r *reader) skipModulePayload() error {
	for {
		opcode, _, err := r.readLength()
		if err != nil {
			return err
		}
		switch opcode {
		case moduleOpEOF:
			return nil
		case moduleOpSInt, moduleOpUInt:
			if _, _, err := r.readLength(); err != nil {
				return err
			}
		case moduleOpString:
			if _, _, err := r.skipString(); err != nil {
				return err
			}
		case moduleOpFloat:
			if err := r.skipBinaryFloat(); err != nil {
				return err
			}
		case moduleOpDouble:
			if err := r.skipBinaryDouble(); err != nil {
				return err
			}
		default:
			return r.err(fmt.Sprintf("unknown module payload opcode %d", opcode))
		}
	}
}
