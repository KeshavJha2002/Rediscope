package rdb

import (
	"errors"
	"fmt"
)

var (
	ErrTooSmall          = errors.New("file is too small to be a valid RDB")
	ErrInvalidSignature  = errors.New("missing or invalid REDIS magic signature")
	ErrTruncated         = errors.New("rdb stream is truncated")
	ErrUnsupportedType   = errors.New("unsupported RDB value type")
	ErrCorruptPayload    = errors.New("corrupted or malformed RDB payload")
	ErrUnsupportedOpcode = errors.New("unsupported or deprecated RDB opcode")
)

type ParseError struct {
	Path    string
	Offset  int
	Opcode  byte
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Path != "" && e.Offset >= 0 {
		if e.Err != nil {
			return fmt.Sprintf("%s at offset %d: %s: %v", e.Path, e.Offset, e.Message, e.Err)
		}
		return fmt.Sprintf("%s at offset %d: %s", e.Path, e.Offset, e.Message)
	}
	if e.Offset >= 0 {
		if e.Err != nil {
			return fmt.Sprintf("rdb offset %d: %s: %v", e.Offset, e.Message, e.Err)
		}
		return fmt.Sprintf("rdb offset %d: %s", e.Offset, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
