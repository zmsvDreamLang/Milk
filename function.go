package lua

import (
	"fmt"
	"strings"
)

const (
	VarArgHasArg   uint8 = 1
	VarArgIsVarArg uint8 = 2
	VarArgNeedsArg uint8 = 4
)

type DbgLocalInfo struct {
	Name    string
	StartPc int
	EndPc   int
}

type DbgCall struct {
	Name string
	Pc   int
}

type FunctionProto struct {
	SourceName         string
	LineDefined        int
	LastLineDefined    int
	NumUpvalues        uint8
	NumParameters      uint8
	IsVarArg           uint8
	NumUsedRegisters   uint8
	Code               []uint32
	Constants          []LValue
	FunctionPrototypes []*FunctionProto

	DbgSourcePositions []int
	DbgLocals          []*DbgLocalInfo
	DbgCalls           []DbgCall
	DbgUpvalues        []string

	stringConstants []string
}

/* Upvalue {{{ */

type Upvalue struct {
	next   *Upvalue
	reg    *registry
	index  int
	value  LValue
	closed bool
}

func (uv *Upvalue) Value() LValue {
	//if uv.IsClosed() {
	if uv.closed || uv.reg == nil {
		return uv.value
	}
	//return uv.reg.Get(uv.index)
	return uv.reg.array[uv.index]
}

func (uv *Upvalue) SetValue(value LValue) {
	if uv.IsClosed() {
		uv.value = value
	} else {
		uv.reg.Set(uv.index, value)
	}
}

func (uv *Upvalue) Close() {
	value := uv.Value()
	uv.closed = true
	uv.value = value
}

func (uv *Upvalue) IsClosed() bool {
	return uv.closed || uv.reg == nil
}

func UpvalueIndex(i int) int {
	return GlobalsIndex - i
}

/* }}} */

/* FunctionProto {{{ */

func newFunctionProto(name string) *FunctionProto {
	return &FunctionProto{
		SourceName:         name,
		LineDefined:        0,
		LastLineDefined:    0,
		NumUpvalues:        0,
		NumParameters:      0,
		IsVarArg:           0,
		NumUsedRegisters:   2,
		Code:               make([]uint32, 0, 128),
		Constants:          make([]LValue, 0, 32),
		FunctionPrototypes: make([]*FunctionProto, 0, 16),

		DbgSourcePositions: make([]int, 0, 128),
		DbgLocals:          make([]*DbgLocalInfo, 0, 16),
		DbgCalls:           make([]DbgCall, 0, 128),
		DbgUpvalues:        make([]string, 0, 16),

		stringConstants: make([]string, 0, 32),
	}
}

func (fp *FunctionProto) RawBytecodeString() string {
	return string(fp.RawBytecode())
}

func (fp *FunctionProto) RawBytecode() []byte {
	buf := make([]byte, 0, len(fp.Code)*4+12)
	buf = append(buf, 0x1b, 0x4c, 0x75, 0x61) // Lua signature
	buf = append(buf, 0x51)                   // Version
	buf = append(buf, 0x00)                   // Format
	buf = append(buf, 0x01)                   // Endianness
	buf = append(buf, 0x04)                   // Size of int
	buf = append(buf, 0x04)                   // Size of size_t
	buf = append(buf, 0x04)                   // Size of Instruction
	buf = append(buf, 0x08)                   // Size of lua_Number
	buf = append(buf, 0x00)                   // Integral flag

	// Function header
	buf = append(buf, encodeString(fp.SourceName)...)
	buf = append(buf, encodeInt(fp.LineDefined)...)
	buf = append(buf, encodeInt(fp.LastLineDefined)...)
	buf = append(buf, fp.NumUpvalues)
	buf = append(buf, fp.NumParameters)
	buf = append(buf, fp.IsVarArg)
	buf = append(buf, fp.NumUsedRegisters)

	// Code
	buf = append(buf, encodeInt(len(fp.Code))...)
	for _, code := range fp.Code {
		buf = append(buf, byte(code), byte(code>>8), byte(code>>16), byte(code>>24))
	}

	// Constants
	buf = append(buf, encodeInt(len(fp.Constants))...)
	for _, constant := range fp.Constants {
		buf = append(buf, encodeConstant(constant)...)
	}

	// Function prototypes
	buf = append(buf, encodeInt(len(fp.FunctionPrototypes))...)
	for _, proto := range fp.FunctionPrototypes {
		buf = append(buf, proto.RawBytecode()...)
	}

	// Debug information
	buf = append(buf, encodeInt(len(fp.DbgSourcePositions))...)
	for _, pos := range fp.DbgSourcePositions {
		buf = append(buf, encodeInt(pos)...)
	}

	buf = append(buf, encodeInt(len(fp.DbgLocals))...)
	for _, local := range fp.DbgLocals {
		buf = append(buf, encodeString(local.Name)...)
		buf = append(buf, encodeInt(local.StartPc)...)
		buf = append(buf, encodeInt(local.EndPc)...)
	}

	buf = append(buf, encodeInt(len(fp.DbgUpvalues))...)
	for _, upvalue := range fp.DbgUpvalues {
		buf = append(buf, encodeString(upvalue)...)
	}

	return buf
}

func encodeString(s string) []byte {
	b := []byte(s)
	return append(encodeInt(len(b)), b...)
}

func encodeInt(i int) []byte {
	return []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
}

func encodeConstant(c LValue) []byte {
	// This function should encode the constant based on its type
	// For simplicity, let's assume all constants are strings
	return append([]byte{0x04}, encodeString(c.String())...)
}

func (fp *FunctionProto) String() string {
	return fp.str(1, 0)
}

func (fp *FunctionProto) str(level int, count int) string {
	indent := strings.Repeat("  ", level-1)
	buf := []string{}
	buf = append(buf, fmt.Sprintf("%v; function [%v] definition (level %v)\n",
		indent, count, level))
	buf = append(buf, fmt.Sprintf("%v; %v upvalues, %v params, %v stacks\n",
		indent, fp.NumUpvalues, fp.NumParameters, fp.NumUsedRegisters))
	for reg, linfo := range fp.DbgLocals {
		buf = append(buf, fmt.Sprintf("%v.local %v ; %v\n", indent, linfo.Name, reg))
	}
	for reg, upvalue := range fp.DbgUpvalues {
		buf = append(buf, fmt.Sprintf("%v.upvalue %v ; %v\n", indent, upvalue, reg))
	}
	for reg, conzt := range fp.Constants {
		buf = append(buf, fmt.Sprintf("%v.const %v ; %v\n", indent, conzt.String(), reg))
	}
	buf = append(buf, "\n")

	protono := 0
	for no, code := range fp.Code {
		inst := opGetOpCode(code)
		if inst == OP_CLOSURE {
			buf = append(buf, "\n")
			buf = append(buf, fp.FunctionPrototypes[protono].str(level+1, protono))
			buf = append(buf, "\n")
			protono++
		}
		buf = append(buf, fmt.Sprintf("%v[%03d] %v (line:%v)\n",
			indent, no+1, opToString(code), fp.DbgSourcePositions[no]))

	}
	buf = append(buf, fmt.Sprintf("%v; end of function\n", indent))
	return strings.Join(buf, "")
}

/* }}} */

/* LFunction {{{ */

func newLFunctionL(proto *FunctionProto, env *LTable, nupvalue int) *LFunction {
	return &LFunction{
		IsG: false,
		Env: env,

		Proto:     proto,
		GFunction: nil,
		Upvalues:  make([]*Upvalue, nupvalue),
	}
}

func newLFunctionG(gfunc LGFunction, env *LTable, nupvalue int) *LFunction {
	return &LFunction{
		IsG: true,
		Env: env,

		Proto:     nil,
		GFunction: gfunc,
		Upvalues:  make([]*Upvalue, nupvalue),
	}
}

func (fn *LFunction) LocalName(regno, pc int) (string, bool) {
	if fn.IsG {
		return "", false
	}
	p := fn.Proto
	for i := 0; i < len(p.DbgLocals) && p.DbgLocals[i].StartPc < pc; i++ {
		if pc < p.DbgLocals[i].EndPc {
			regno--
			if regno == 0 {
				return p.DbgLocals[i].Name, true
			}
		}
	}
	return "", false
}

/* }}} */
