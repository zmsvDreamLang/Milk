package lua

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"

	toml "github.com/pelletier/go-toml"
)

func OpenBase64(L *LState) int {
	mod := L.RegisterModule(Base64LibName, base64Funcs)
	L.Push(mod)
	return 1
}

func OpenJson(L *LState) int {
	mod := L.RegisterModule(JsonLibName, jsonFuncs)
	L.Push(mod)
	return 1
}

func OpenXml(L *LState) int {
	mod := L.RegisterModule(XmlLibName, xmlFuncs)
	L.Push(mod)
	return 1
}

func OpenToml(L *LState) int {
	mod := L.RegisterModule(TomlLibName, tomlFuncs)
	L.Push(mod)
	return 1
}

func OpenHex(L *LState) int {
	mod := L.RegisterModule(HexLibName, hexFuncs)
	L.Push(mod)
	return 1
}

var base64Funcs = map[string]LGFunction{
	"encode": encodingBase64,
	"decode": decodingBase64,
}

var jsonFuncs = map[string]LGFunction{
	"encode": encodingJson,
	"decode": decodingJson,
}

var xmlFuncs = map[string]LGFunction{
	"encode": encodingXml,
	"decode": decodingXml,
}

var tomlFuncs = map[string]LGFunction{
	"encode": encodingToml,
	"decode": decodingToml,
}

var hexFuncs = map[string]LGFunction{
	"encode": encodingHex,
}

func encodingBase64(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(base64.StdEncoding.EncodeToString([]byte(str))))
	return 1
}

func decodingBase64(L *LState) int {
	str := L.CheckString(1)
	data, _ := base64.StdEncoding.DecodeString(str)
	resstr := string(data)
	L.Push(LString(resstr))
	return 1
}

func encodingHex(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(hex.EncodeToString([]byte(str))))
	return 1
}

func encodingXml(L *LState) int {
	tbl := L.CheckTable(1)
	root := L.OptString(2, "root")

	xmlStruct := tableToXmlStruct(tbl, root)

	data, err := xml.MarshalIndent(xmlStruct, "", "  ")
	if err != nil {
		L.Push(LNil)
		return 1
	}

	result := []byte(xml.Header + string(data))

	L.Push(LString(result))
	return 1
}

type xmlElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr   `xml:",attr"`
	Content []xmlElement `xml:",any"`
	Value   string       `xml:",chardata"`
}

func tableToXmlStruct(tbl *LTable, name string) xmlElement {
	elem := xmlElement{XMLName: xml.Name{Local: name}}

	tbl.ForEach(func(k, v LValue) {
		switch v.Type() {
		case LTTable:
			// 递归处理嵌套表
			childName := fmt.Sprintf("%v", k)
			elem.Content = append(elem.Content, tableToXmlStruct(v.(*LTable), childName))
		case LTString, LTNumber, LTBool:
			// 处理简单类型
			if k.Type() == LTString {
				// 作为属性处理
				elem.Attrs = append(elem.Attrs, xml.Attr{Name: xml.Name{Local: k.String()}, Value: fmt.Sprintf("%v", v)})
			} else {
				// 作为内容处理
				elem.Value += fmt.Sprintf("%v", v)
			}
		}
	})

	return elem
}

func decodingXml(L *LState) int {
	str := L.CheckString(1)
	var tbl LTable
	xml.Unmarshal([]byte(str), &tbl)
	L.Push(&tbl)
	return 1
}

func encodingJson(LuaVM *LState) int {
	lv := LuaVM.CheckTable(1)
	goMap := toGoMap(lv)
	jsonData, err := json.Marshal(goMap)
	if err != nil {
		LuaVM.Push(LNil)
		return 2
	}

	LuaVM.Push(LString(string(jsonData)))
	return 1
}

func decodingJson(LuaVM *LState) int {
	jsonStr := LuaVM.CheckString(1)

	var goMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &goMap)
	if err != nil {
		LuaVM.Push(LNil)
		return 1
	}

	luaTable := toLuaTable(LuaVM, goMap)
	LuaVM.Push(luaTable)
	return 1
}

func toGoMap(lv *LTable) map[string]interface{} {
	goMap := make(map[string]interface{})
	lv.ForEach(func(key LValue, value LValue) {
		goMap[key.String()] = toGoValue(value)
	})
	return goMap
}

func toLuaTable(LuaVM *LState, goMap map[string]interface{}) *LTable {
	luaTable := LuaVM.NewTable()
	for key, value := range goMap {
		luaTable.RawSetString(key, toLuaValue(LuaVM, value))
	}
	return luaTable
}

func toLuaValue(LuaVM *LState, value interface{}) LValue {
	switch v := value.(type) {
	case string:
		return LString(v)
	case float64:
		return LNumber(v)
	case bool:
		return LBool(v)
	case map[string]interface{}:
		return toLuaTable(LuaVM, v)
	default:
		return LNil
	}
}

func toGoValue(lv LValue) interface{} {
	switch lv.Type() {
	case LTString:
		return lv.String()
	case LTNumber:
		return float64(LVAsNumber(lv))
	case LTBool:
		return LVAsBool(lv)
	case LTTable:
		return toGoMap(lv.(*LTable))
	default:
		return nil
	}
}

func encodingToml(L *LState) int {
	tbl := L.CheckTable(1)
	goMap := toGoMap(tbl)

	tomlData, err := toml.Marshal(goMap)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}

	L.Push(LString(string(tomlData)))
	return 1
}

func decodingToml(L *LState) int {
	tomlStr := L.CheckString(1)

	var goMap map[string]interface{}
	err := toml.Unmarshal([]byte(tomlStr), &goMap)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}

	luaTable := toLuaTable(L, goMap)
	L.Push(luaTable)
	return 1
}
