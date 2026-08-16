package rdb

type FileModel struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Bytes      int           `json:"bytes"`
	Version    string        `json:"version"`
	RawVersion string        `json:"rawVersion,omitempty"`
	CountLabel string        `json:"countLabel"`
	Hex        string        `json:"hex"`
	Groups     []RecordGroup `json:"groups"`

	// Internal fields omitted from JSON to keep index.html lightweight
	Size     int         `json:"-"`
	Sections []Section   `json:"-"`
	Keys     []KeyRecord `json:"-"`
}

type RecordGroup struct {
	Title   string   `json:"title"`
	Records []Record `json:"records"`
}

type Record struct {
	ID       string         `json:"id"`
	Start    int            `json:"start"`
	End      int            `json:"end"`
	Label    string         `json:"label"`
	Type     string         `json:"type"`
	RDBType  string         `json:"rdbType,omitempty"`
	Color    string         `json:"color"`
	Value    string         `json:"value"`
	JSON     string         `json:"json,omitempty"`
	Encoding string         `json:"encoding"`
	Size     string         `json:"size"`
	Summary  string         `json:"summary"`
	Parts    []RecordPart   `json:"parts,omitempty"`
	Strings  []RecordString `json:"strings,omitempty"`
}

type RecordPart struct {
	Kind  string `json:"kind"` // "type", "key", "value"
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type RecordString struct {
	Kind  string `json:"kind"` // "key", "value"
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
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
