package acb

import (
	"encoding/binary"
	"fmt"
)

// @UTF column flags
const (
	colFlagName    = 0x10
	colFlagDefault = 0x20
	colFlagRow     = 0x40
)

// @UTF column types
const (
	colTypeU8     = 0x00
	colTypeS8     = 0x01
	colTypeU16    = 0x02
	colTypeS16    = 0x03
	colTypeU32    = 0x04
	colTypeS32    = 0x05
	colTypeU64    = 0x06
	colTypeS64    = 0x07
	colTypeFloat  = 0x08
	colTypeString = 0x0a
	colTypeVLData = 0x0b
)

func colTypeSize(t byte) int {
	switch t {
	case colTypeU8, colTypeS8:
		return 1
	case colTypeU16, colTypeS16:
		return 2
	case colTypeU32, colTypeS32, colTypeFloat, colTypeString:
		return 4
	case colTypeU64, colTypeS64, colTypeVLData:
		return 8
	}
	return 0
}

type utfColumn struct {
	flag      byte
	typ       byte
	name      string
	schemaOff int // offset within schema data (for DEFAULT columns)
	rowOff    int // offset within a row (for ROW columns)
}

type utfTable struct {
	data          []byte // full file data (absolute addressing)
	tableOffset   int
	tableSize     int
	rowsOffset    int
	stringsOffset int
	dataOffset    int
	nameOffset    int
	rowWidth      int
	rows          int
	columns       []utfColumn
	stringTable   []byte
}

func openUTF(data []byte, tableOffset int) (*utfTable, error) {
	if tableOffset+0x20 > len(data) {
		return nil, fmt.Errorf("utf: table offset %x out of bounds", tableOffset)
	}
	if string(data[tableOffset:tableOffset+4]) != "@UTF" {
		return nil, fmt.Errorf("utf: bad magic at %x", tableOffset)
	}

	t := &utfTable{data: data, tableOffset: tableOffset}
	t.tableSize = int(binary.BigEndian.Uint32(data[tableOffset+0x04:])) + 8
	t.rowsOffset = int(binary.BigEndian.Uint16(data[tableOffset+0x0a:])) + 8
	t.stringsOffset = int(binary.BigEndian.Uint32(data[tableOffset+0x0c:])) + 8
	t.dataOffset = int(binary.BigEndian.Uint32(data[tableOffset+0x10:])) + 8
	t.nameOffset = int(binary.BigEndian.Uint32(data[tableOffset+0x14:]))
	numCols := int(binary.BigEndian.Uint16(data[tableOffset+0x18:]))
	t.rowWidth = int(binary.BigEndian.Uint16(data[tableOffset+0x1a:]))
	t.rows = int(binary.BigEndian.Uint32(data[tableOffset+0x1c:]))

	strStart := tableOffset + t.stringsOffset
	strEnd := tableOffset + t.dataOffset
	if strEnd > len(data) || strStart > strEnd {
		return nil, fmt.Errorf("utf: string section out of bounds")
	}
	t.stringTable = data[strStart:strEnd]

	schemaStart := tableOffset + 0x20
	schemaEnd := tableOffset + t.rowsOffset
	if schemaEnd > len(data) {
		return nil, fmt.Errorf("utf: schema section out of bounds")
	}
	schemaData := data[schemaStart:schemaEnd]

	t.columns = make([]utfColumn, numCols)
	schemaPos := 0
	columnOffset := 0 // running offset within a row

	for i := 0; i < numCols; i++ {
		if schemaPos+5 > len(schemaData) {
			return nil, fmt.Errorf("utf: schema pos out of bounds at column %d", i)
		}
		info := schemaData[schemaPos]
		nameOff := int(binary.BigEndian.Uint32(schemaData[schemaPos+1:]))
		schemaPos += 5

		col := &t.columns[i]
		col.flag = info & 0xf0
		col.typ = info & 0x0f

		if nameOff < len(t.stringTable) {
			end := nameOff
			for end < len(t.stringTable) && t.stringTable[end] != 0 {
				end++
			}
			col.name = string(t.stringTable[nameOff:end])
		}

		sz := colTypeSize(col.typ)
		if col.flag&colFlagDefault != 0 {
			col.schemaOff = schemaPos
			schemaPos += sz
		} else if col.flag&colFlagRow != 0 {
			col.rowOff = columnOffset
			columnOffset += sz
		}
	}

	return t, nil
}

func (t *utfTable) tableName() string {
	if t.nameOffset >= len(t.stringTable) {
		return ""
	}
	end := t.nameOffset
	for end < len(t.stringTable) && t.stringTable[end] != 0 {
		end++
	}
	return string(t.stringTable[t.nameOffset:end])
}

func (t *utfTable) findColumn(name string) int {
	for i, col := range t.columns {
		if col.name == name {
			return i
		}
	}
	return -1
}

// colBytes returns a byte slice pointing to the raw value of a column for a given row.
func (t *utfTable) colBytes(row, colIdx int) ([]byte, error) {
	if colIdx < 0 || colIdx >= len(t.columns) {
		return nil, fmt.Errorf("utf: column index %d out of range", colIdx)
	}
	col := &t.columns[colIdx]
	var absOff int
	if col.flag&colFlagDefault != 0 {
		// schema data is at tableOffset + 0x20; schemaOff is relative to schema data start
		absOff = t.tableOffset + 0x20 + col.schemaOff
	} else if col.flag&colFlagRow != 0 {
		absOff = t.tableOffset + t.rowsOffset + row*t.rowWidth + col.rowOff
	} else {
		return nil, fmt.Errorf("utf: column %d has no data", colIdx)
	}
	sz := colTypeSize(col.typ)
	if absOff+sz > len(t.data) {
		return nil, fmt.Errorf("utf: column data out of bounds at %x+%d", absOff, sz)
	}
	return t.data[absOff : absOff+sz], nil
}

func (t *utfTable) queryU8(row int, colName string) (uint8, bool) {
	idx := t.findColumn(colName)
	if idx < 0 {
		return 0, false
	}
	b, err := t.colBytes(row, idx)
	if err != nil {
		return 0, false
	}
	return b[0], true
}

func (t *utfTable) queryU16(row int, colName string) (uint16, bool) {
	idx := t.findColumn(colName)
	if idx < 0 {
		return 0, false
	}
	b, err := t.colBytes(row, idx)
	if err != nil {
		return 0, false
	}
	return binary.BigEndian.Uint16(b), true
}

func (t *utfTable) queryString(row int, colName string) (string, bool) {
	idx := t.findColumn(colName)
	if idx < 0 {
		return "", false
	}
	b, err := t.colBytes(row, idx)
	if err != nil {
		return "", false
	}
	off := int(binary.BigEndian.Uint32(b))
	if off >= len(t.stringTable) {
		return "", false
	}
	end := off
	for end < len(t.stringTable) && t.stringTable[end] != 0 {
		end++
	}
	return string(t.stringTable[off:end]), true
}

// queryVLData returns the absolute offset into data and size for a VLDATA column.
// absOffset = tableOffset + dataOffset + stored_relative_offset
func (t *utfTable) queryVLData(row int, colName string) (absOffset int, size int, ok bool) {
	idx := t.findColumn(colName)
	if idx < 0 {
		return 0, 0, false
	}
	b, err := t.colBytes(row, idx)
	if err != nil {
		return 0, 0, false
	}
	relOff := int(binary.BigEndian.Uint32(b[0:4]))
	sz := int(binary.BigEndian.Uint32(b[4:8]))
	return t.tableOffset + t.dataOffset + relOff, sz, true
}

// File is the result of loading an ACB.
type File struct {
	Names map[uint16]string // waveID → cue name
}

// acbParser holds working state while walking the cue graph.
type acbParser struct {
	data          []byte
	header        *utfTable
	cueTable      *utfTable
	cuenameTable  *utfTable
	synthTable    *utfTable
	waveformTable *utfTable
	sequenceTable *utfTable
	trackTable    *utfTable
	trackCmdTable *utfTable
	names         map[uint16]string
}

func (p *acbParser) openSubtable(columnName string) (*utfTable, error) {
	absOff, size, ok := p.header.queryVLData(0, columnName)
	if !ok || size == 0 {
		return nil, fmt.Errorf("acb: subtable column %q not found or empty", columnName)
	}
	return openUTF(p.data, absOff)
}

func (p *acbParser) getCueTable() *utfTable {
	if p.cueTable != nil {
		return p.cueTable
	}
	t, _ := p.openSubtable("CueTable")
	p.cueTable = t
	return t
}

func (p *acbParser) getCueNameTable() *utfTable {
	if p.cuenameTable != nil {
		return p.cuenameTable
	}
	t, _ := p.openSubtable("CueNameTable")
	p.cuenameTable = t
	return t
}

func (p *acbParser) getSynthTable() *utfTable {
	if p.synthTable != nil {
		return p.synthTable
	}
	t, _ := p.openSubtable("SynthTable")
	p.synthTable = t
	return t
}

func (p *acbParser) getWaveformTable() *utfTable {
	if p.waveformTable != nil {
		return p.waveformTable
	}
	t, _ := p.openSubtable("WaveformTable")
	p.waveformTable = t
	return t
}

func (p *acbParser) getSequenceTable() *utfTable {
	if p.sequenceTable != nil {
		return p.sequenceTable
	}
	t, _ := p.openSubtable("SequenceTable")
	p.sequenceTable = t
	return t
}

func (p *acbParser) getTrackTable() *utfTable {
	if p.trackTable != nil {
		return p.trackTable
	}
	t, _ := p.openSubtable("TrackTable")
	p.trackTable = t
	return t
}

func (p *acbParser) getTrackCmdTable() *utfTable {
	if p.trackCmdTable != nil {
		return p.trackCmdTable
	}
	t, _ := p.openSubtable("TrackEventTable")
	if t == nil {
		t, _ = p.openSubtable("CommandTable")
	}
	p.trackCmdTable = t
	return t
}

func (p *acbParser) recordWaveform(idx int, cueName string) {
	wt := p.getWaveformTable()
	if wt == nil || idx >= wt.rows {
		return
	}
	// Try "Id" first (older), then "MemoryAwbId" (newer split)
	waveID, ok := wt.queryU16(idx, "Id")
	if !ok {
		waveID, ok = wt.queryU16(idx, "MemoryAwbId")
		if !ok {
			return
		}
	}
	if _, exists := p.names[waveID]; !exists {
		p.names[waveID] = cueName
	}
}

func (p *acbParser) loadSynth(idx int, cueName string, depth int) {
	if depth > 3 {
		return
	}
	st := p.getSynthTable()
	if st == nil || idx >= st.rows {
		return
	}
	absOff, size, ok := st.queryVLData(idx, "ReferenceItems")
	if !ok || size == 0 {
		return
	}
	count := size / 4
	for i := 0; i < count; i++ {
		off := absOff + i*4
		if off+4 > len(p.data) {
			break
		}
		itemType := binary.BigEndian.Uint16(p.data[off:])
		itemIndex := int(binary.BigEndian.Uint16(p.data[off+2:]))
		switch itemType {
		case 0x00: // no reference
			return
		case 0x01: // Waveform
			p.recordWaveform(itemIndex, cueName)
		case 0x02: // Synth (recursive)
			p.loadSynth(itemIndex, cueName, depth+1)
		case 0x03: // Sequence
			p.loadSequence(itemIndex, cueName, depth+1)
		}
	}
}

func (p *acbParser) loadTrackCommand(idx int, cueName string, depth int) {
	tct := p.getTrackCmdTable()
	if tct == nil || idx >= tct.rows {
		return
	}
	absOff, size, ok := tct.queryVLData(idx, "Command")
	if !ok || size == 0 {
		return
	}
	pos := 0
	for pos < size {
		if absOff+pos+3 > len(p.data) {
			break
		}
		tlvCode := binary.BigEndian.Uint16(p.data[absOff+pos:])
		tlvSize := int(p.data[absOff+pos+2])
		pos += 3
		switch tlvCode {
		case 2000, 2003: // noteOn
			if tlvSize >= 4 && absOff+pos+4 <= len(p.data) {
				tlvType := binary.BigEndian.Uint16(p.data[absOff+pos:])
				tlvIndex := int(binary.BigEndian.Uint16(p.data[absOff+pos+2:]))
				switch tlvType {
				case 0x02: // Synth
					p.loadSynth(tlvIndex, cueName, depth+1)
				case 0x03: // Sequence
					p.loadSequence(tlvIndex, cueName, depth+1)
				}
			}
		}
		pos += tlvSize
	}
}

func (p *acbParser) loadTrack(idx int, cueName string, depth int) {
	tt := p.getTrackTable()
	if tt == nil || idx >= tt.rows {
		return
	}
	eventIdx, ok := tt.queryU16(idx, "EventIndex")
	if !ok || eventIdx == 0xFFFF {
		return
	}
	p.loadTrackCommand(int(eventIdx), cueName, depth)
}

func (p *acbParser) loadSequence(idx int, cueName string, depth int) {
	if depth > 3 {
		return
	}
	st := p.getSequenceTable()
	if st == nil || idx >= st.rows {
		return
	}
	numTracks, ok := st.queryU16(idx, "NumTracks")
	if !ok {
		return
	}
	absOff, size, ok := st.queryVLData(idx, "TrackIndex")
	if !ok || size == 0 {
		return
	}
	count := int(numTracks)
	if count*2 > size {
		count = size / 2
	}
	for i := 0; i < count; i++ {
		off := absOff + i*2
		if off+2 > len(p.data) {
			break
		}
		trackIdx := int(binary.BigEndian.Uint16(p.data[off:]))
		p.loadTrack(trackIdx, cueName, depth+1)
	}
}

func (p *acbParser) loadCue(idx int, cueName string) {
	ct := p.getCueTable()
	if ct == nil || idx >= ct.rows {
		return
	}
	refType, _ := ct.queryU8(idx, "ReferenceType")
	refIdx, _ := ct.queryU16(idx, "ReferenceIndex")
	switch refType {
	case 1: // Cue → Waveform
		p.recordWaveform(int(refIdx), cueName)
	case 2: // Cue → Synth → Waveform
		p.loadSynth(int(refIdx), cueName, 0)
	case 3: // Cue → Sequence → Track → Command → Synth → Waveform
		p.loadSequence(int(refIdx), cueName, 0)
	}
}

// LoadACB parses an ACB file. It returns the embedded AWB data and a waveID→cueName map.
func LoadACB(data []byte) (*File, error) {
	header, err := openUTF(data, 0)
	if err != nil {
		return nil, fmt.Errorf("acb: %w", err)
	}
	if header.tableName() != "Header" || header.rows != 1 {
		return nil, fmt.Errorf("acb: expected Header table with 1 row, got %q rows=%d", header.tableName(), header.rows)
	}

	p := &acbParser{
		data:   data,
		header: header,
		names:  make(map[uint16]string),
	}

	file := File{}
	file.Names = make(map[uint16]string)

	cnt := p.getCueNameTable()
	if cnt != nil {
		for i := 0; i < cnt.rows; i++ {
			cueName, ok := cnt.queryString(i, "CueName")
			if !ok {
				continue
			}
			cueIdx, ok := cnt.queryU16(i, "CueIndex")
			if !ok {
				continue
			}
			file.Names[cueIdx] = cueName
			p.loadCue(int(cueIdx), cueName)
		}
	}

	return &file, nil
}
