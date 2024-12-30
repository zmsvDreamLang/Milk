package lua

const (
	// BaseLibName is here for consistency; the base functions have no namespace/library.
	BaseLibName = ""
	// LoadLibName is here for consistency; the loading system has no namespace/library.
	LoadLibName = "package"
	// TabLibName is the name of the table Library.
	TabLibName = "table"
	// IoLibName is the name of the io Library.
	IoLibName = "io"
	// OsLibName is the name of the os Library.
	OsLibName = "os"
	// StringLibName is the name of the string Library.
	StringLibName = "string"
	// MathLibName is the name of the math Library.
	MathLibName = "math"
	// DebugLibName is the name of the debug Library.
	DebugLibName = "debug"
	// ChannelLibName is the name of the channel Library.
	ChannelLibName = "channel"
	// CoroutineLibName is the name of the coroutine Library.
	CoroutineLibName = "coroutine"
	// Base64LibName is the name of the base64 Library.
	Base64LibName = "base64"
	// JsonLibName is the name of the json Library.
	JsonLibName = "json"
	// XmlLibName is the name of the xml Library.
	XmlLibName = "xml"
	// TomlLibName is the name of the toml Library.
	TomlLibName = "toml"
	// HexLibName is the name of the hex Library.
	HexLibName = "hex"
	// RegexpLibName is the name of the regexp Library.
	RegexpLibName = "regexp"
	// MatrixLibName is the name of the matrix Library.
	MatrixLibName = "matrix"
	// StatisticLibName is the name of the statistic Library.
	StatisticLibName = "statistic"
	// CalculusLibName is the name of the calculus Library.
	CalculusLibName = "calculus"
	// NeurolibName is the name of the neurolib Library.
	NeurolibName = "neuro"
	// FFILibName is the name of the FFI Library.
	FFILibName = "ffi"
	// HttpLibName is the name of the http Library.
	HttpLibName = "http"
	// DefaultExportLibName is the name of the default export Library.
	DefaultExportLibName = "lib"
)

type luaLib struct {
	libName string
	libFunc LGFunction
}

var luaLibs = []luaLib{
	{LoadLibName, OpenPackage},
	{BaseLibName, OpenBase},
	{TabLibName, OpenTable},
	{IoLibName, OpenIo},
	{OsLibName, OpenOs},
	{StringLibName, OpenString},
	{MathLibName, OpenMath},
	{DebugLibName, OpenDebug},
	{ChannelLibName, OpenChannel},
	{CoroutineLibName, OpenCoroutine},
	{Base64LibName, OpenBase64},
	{JsonLibName, OpenJson},
	{XmlLibName, OpenXml},
	{TomlLibName, OpenToml},
	{HexLibName, OpenHex},
	{RegexpLibName, OpenRegexp},
	{MatrixLibName, OpenMatrix},
	{StatisticLibName, OpenStatistic},
	{CalculusLibName, OpenCalculus},
	{NeurolibName, OpenNeurolib},
	{FFILibName, OpenFFI},
	{HttpLibName, OpenHttp},
	{DefaultExportLibName, OpenLib},
}

// OpenLibs loads the built-in libraries. It is equivalent to running OpenLoad,
// then OpenBase, then iterating over the other OpenXXX functions in any order.
func (ls *LState) OpenLibs() {
	// NB: Map iteration order in Go is deliberately randomized, so must open Load/Base
	// prior to iterating.
	for _, lib := range luaLibs {
		ls.Push(ls.NewFunction(lib.libFunc))
		ls.Push(LString(lib.libName))
		ls.Call(1, 0)
	}
}
