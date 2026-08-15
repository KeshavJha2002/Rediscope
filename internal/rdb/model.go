package rdb

type FileModel struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Size     int         `json:"size"`
	Version  string      `json:"version"`
	Hex      string      `json:"hex"`
	Sections []Section   `json:"sections"`
	Keys     []KeyRecord `json:"keys"`
}

type Section struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Size     int    `json:"size"`
	Meta     string `json:"meta,omitempty"`
	KeyIndex int    `json:"keyIndex,omitempty"`
}

type KeyRecord struct {
	DB          int    `json:"db"`
	Key         string `json:"key"`
	TypeByte    byte   `json:"typeByte"`
	TypeName    string `json:"typeName"`
	Encoding    string `json:"encoding"`
	ExpiryMS    *int64 `json:"expiryMs,omitempty"`
	RecordStart int    `json:"recordStart"`
	RecordEnd   int    `json:"recordEnd"`
	KeyStart    int    `json:"keyStart"`
	KeyEnd      int    `json:"keyEnd"`
	ValueStart  int    `json:"valueStart"`
	ValueEnd    int    `json:"valueEnd"`
}
