package gopdf

import (
	"bytes"
	"log"
	"sort"

	"github.com/signintech/gopdf/fontmaker/core"
)

var EntrySelectors = []int{0, 0, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}

type PdfDictionaryObj struct {
	buffer             bytes.Buffer
	PtrToSubsetFontObj *SubsetFontObj
}

func (me *PdfDictionaryObj) Init(funcGetRoot func() *GoPdf) {
}

func (me *PdfDictionaryObj) Build() {
	err := me.makeFont()
	if err != nil {
		log.Panicf("%s", err.Error())
	}
}

func (me *PdfDictionaryObj) GetType() string {
	return "PdfDictionary"
}

func (me *PdfDictionaryObj) GetObjBuff() *bytes.Buffer {
	return &me.buffer
}

func (me *PdfDictionaryObj) SetPtrToSubsetFontObj(ptr *SubsetFontObj) {
	me.PtrToSubsetFontObj = ptr
}

func (me *PdfDictionaryObj) makeFont() error {
	var buff Buff
	ttfp := me.PtrToSubsetFontObj.GetTTFParser()
	tables := make(map[string]core.TableDirectoryEntry)
	tables["cvt "] = ttfp.GetTables()["cvt "] //มีช่องว่างด้วยนะ
	tables["fpgm"] = ttfp.GetTables()["fpgm"]
	tables["glyf"] = ttfp.GetTables()["glyf"]
	tables["head"] = ttfp.GetTables()["head"]
	tables["hhea"] = ttfp.GetTables()["hhea"]
	tables["hmtx"] = ttfp.GetTables()["hmtx"]
	tables["loca"] = ttfp.GetTables()["loca"]
	tables["maxp"] = ttfp.GetTables()["maxp"]
	tables["prep"] = ttfp.GetTables()["prep"]
	tableCount := len(tables)
	selector := EntrySelectors[tableCount]

	WriteUInt32(&buff, 0x00010000)
	WriteUInt16(&buff, uint(tableCount))
	WriteUInt16(&buff, ((1 << uint(selector)) * 16))
	WriteUInt16(&buff, uint(selector))
	WriteUInt16(&buff, (uint(tableCount)-(1<<uint(selector)))*16)
	//fmt.Printf("%#v\n\n", buff)
	//fmt.Printf("%#v\n\n%#v\n", tables, ttfp.GetTables())
	//fmt.Printf("tableCount = %d\n", tableCount)
	var tags []string
	for tag, _ := range tables {
		tags = append(tags, tag) //copy all tag
	}
	sort.Strings(tags) //order
	idx := 0
	tablePosition := int(12 + 16*tableCount)
	for idx < tableCount {
		entry := tables[tags[idx]]
		//write data
		entry.Offset = uint64(tablePosition)
		buff.SetPosition(tablePosition)
		if tags[idx] == "glyf" {
			//
		} else {
			WriteBytes(&buff, ttfp.FontData(), int(entry.Offset), entry.PaddedLength())
		}
		endPosition := buff.Position()
		tablePosition = endPosition

		//write table
		buff.SetPosition(idx*16 + 12)
		WriteTag(&buff, tags[idx])
		WriteUInt32(&buff, uint(entry.CheckSum))
		WriteUInt32(&buff, uint(entry.Offset)) //offset
		WriteUInt32(&buff, uint(entry.Length))

		tablePosition = endPosition
		//fmt.Printf("====tag %s entry.Offset = %d entry.Offset = %d PaddedLength = %d\n", tags[idx], entry.Offset, entry.Offset, entry.PaddedLength())
		idx++
	}
	//fmt.Printf("buff= %#v\n", buff)
	DebugSubType(buff.Bytes())
	return nil
}