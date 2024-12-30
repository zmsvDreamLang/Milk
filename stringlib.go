package lua

import (
	"fmt"
	"strings"

	"MilkLua/pm"
)

const emptyLString LString = LString("")

func OpenString(L *LState) int {
	mod := L.RegisterModule(StringLibName, strFuncs).(*LTable)
	gmatch := L.NewClosure(strGmatch, L.NewFunction(strGmatchIter))
	mod.RawSetString("gmatch", gmatch)
	mod.RawSetString("gfind", gmatch)
	mod.RawSetString("__index", mod)
	L.G.builtinMts[int(LTString)] = mod
	L.Push(mod)
	return 1
}

var strFuncs = map[string]LGFunction{
	"byte":       strByte,
	"char":       strChar,
	"count":      strCount,
	"dump":       strDump,
	"end_with":   strEndsWith,
	"find":       strFind,
	"format":     strFormat,
	"gsub":       strGsub,
	"is_blank":   strIsBlank,
	"len":        strLen,
	"lower":      strLower,
	"match":      strMatch,
	"pad_end":    strPadEnd,
	"pad_start":  strPadStart,
	"rep":        strRep,
	"reverse":    strReverse,
	"split":      strSplit,
	"start_with": strStartsWith,
	"sub":        strSub,
	"trim":       strTrim,
	"trim_end":   strTrimEnd,
	"trim_start": strTrimStart,
	"truncate":   strTruncate,
	"upper":      strUpper,
}

func strByte(L *LState) int {
	str := L.CheckString(1)
	start := L.OptInt(2, 1) - 1
	end := L.OptInt(3, -1)
	l := len(str)
	if start < 0 {
		start = l + start + 1
	}
	if end < 0 {
		end = l + end + 1
	}

	if L.GetTop() == 2 {
		if start < 0 || start >= l {
			return 0
		}
		L.Push(LNumber(str[start]))
		return 1
	}

	start = intMax(start, 0)
	end = intMin(end, l)
	if end < 0 || end <= start || start >= l {
		return 0
	}

	for i := start; i < end; i++ {
		L.Push(LNumber(str[i]))
	}
	return end - start
}

func strChar(L *LState) int {
	top := L.GetTop()
	bytes := make([]byte, top)
	for i := 1; i <= top; i++ {
		bytes[i-1] = uint8(L.CheckInt(i))
	}
	L.Push(LString(string(bytes)))
	return 1
}

func strCount(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	init := luaIndex2StringIndex(str, L.OptInt(3, 1), true)
	plain := L.OptBool(4, false)

	if plain {
		count := 0
		for {
			pos := strings.Index(str[init:], pattern)
			if pos < 0 {
				break
			}
			count++
			init += pos + len(pattern)
		}
		L.Push(LNumber(count))
		return 1
	}

	mds, err := pm.Find(pattern, unsafeFastStringToReadOnlyBytes(str), init, -1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	L.Push(LNumber(len(mds)))
	return 1
}

func strDump(L *LState) int {
	L.RaiseError("GopherLua does not support the string.dump")
	return 0
}

func strEndsWith(L *LState) int {
	str := L.CheckString(1)
	sub := L.CheckString(2)
	if len(sub) > len(str) {
		L.Push(LFalse)
		return 1
	}
	L.Push(LBool(strings.HasSuffix(str, sub)))
	return 1
}

func strFind(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	if len(pattern) == 0 {
		L.Push(LNumber(1))
		L.Push(LNumber(0))
		return 2
	}
	init := luaIndex2StringIndex(str, L.OptInt(3, 1), true)
	plain := L.OptBool(4, false)

	if plain {
		pos := strings.Index(str[init:], pattern)
		if pos < 0 {
			L.Push(LNil)
			return 1
		}
		L.Push(LNumber(init+pos) + 1)
		L.Push(LNumber(init + pos + len(pattern)))
		return 2
	}

	mds, err := pm.Find(pattern, unsafeFastStringToReadOnlyBytes(str), init, 1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.Push(LNil)
		return 1
	}
	md := mds[0]
	L.Push(LNumber(md.Capture(0) + 1))
	L.Push(LNumber(md.Capture(1)))
	for i := 2; i < md.CaptureLength(); i += 2 {
		if md.IsPosCapture(i) {
			L.Push(LNumber(md.Capture(i)))
		} else {
			L.Push(LString(str[md.Capture(i):md.Capture(i+1)]))
		}
	}
	return md.CaptureLength()/2 + 1
}

func strFormat(L *LState) int {
	str := L.CheckString(1)
	args := make([]interface{}, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		args[i-2] = L.Get(i)
	}
	npat := strings.Count(str, "%") - strings.Count(str, "%%")
	L.Push(LString(fmt.Sprintf(str, args[:intMin(npat, len(args))]...)))
	return 1
}

func strGsub(L *LState) int {
	str := L.CheckString(1)
	pat := L.CheckString(2)
	L.CheckTypes(3, LTString, LTTable, LTFunction)
	repl := L.CheckAny(3)
	limit := L.OptInt(4, -1)

	mds, err := pm.Find(pat, unsafeFastStringToReadOnlyBytes(str), 0, limit)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.SetTop(1)
		L.Push(LNumber(0))
		return 2
	}
	switch lv := repl.(type) {
	case LString:
		L.Push(LString(strGsubStr(L, str, string(lv), mds)))
	case *LTable:
		L.Push(LString(strGsubTable(L, str, lv, mds)))
	case *LFunction:
		L.Push(LString(strGsubFunc(L, str, lv, mds)))
	}
	L.Push(LNumber(len(mds)))
	return 2
}

func strIsBlank(L *LState) int {
	str := L.CheckString(1)
	for _, c := range str {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			L.Push(LFalse)
			return 1
		}
	}
	L.Push(LTrue)
	return 1
}

type replaceInfo struct {
	Indices []int
	String  string
}

func checkCaptureIndex(L *LState, m *pm.MatchData, idx int) {
	if idx <= 2 {
		return
	}
	if idx >= m.CaptureLength() {
		L.RaiseError("invalid capture index")
	}
}

func capturedString(L *LState, m *pm.MatchData, str string, idx int) string {
	checkCaptureIndex(L, m, idx)
	if idx >= m.CaptureLength() && idx == 2 {
		idx = 0
	}
	if m.IsPosCapture(idx) {
		return fmt.Sprint(m.Capture(idx))
	} else {
		return str[m.Capture(idx):m.Capture(idx+1)]
	}
}

func strGsubDoReplace(str string, info []replaceInfo) string {
	offset := 0
	buf := []byte(str)
	for _, replace := range info {
		oldlen := len(buf)
		b1 := append([]byte(""), buf[0:offset+replace.Indices[0]]...)
		b2 := []byte("")
		index2 := offset + replace.Indices[1]
		if index2 <= len(buf) {
			b2 = append(b2, buf[index2:]...)
		}
		buf = append(b1, replace.String...)
		buf = append(buf, b2...)
		offset += len(buf) - oldlen
	}
	return string(buf)
}

func strGsubStr(L *LState, str string, repl string, matches []*pm.MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		sc := newFlagScanner('%', "", "", repl)
		for c, eos := sc.Next(); !eos; c, eos = sc.Next() {
			if !sc.ChangeFlag {
				if sc.HasFlag {
					if c >= '0' && c <= '9' {
						sc.AppendString(capturedString(L, match, str, 2*(int(c)-48)))
					} else {
						sc.AppendChar('%')
						sc.AppendChar(c)
					}
					sc.HasFlag = false
				} else {
					sc.AppendChar(c)
				}
			}
		}
		infoList = append(infoList, replaceInfo{[]int{start, end}, sc.String()})
	}

	return strGsubDoReplace(str, infoList)
}

func strGsubTable(L *LState, str string, repl *LTable, matches []*pm.MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		idx := 0
		if match.CaptureLength() > 2 { // has captures
			idx = 2
		}
		var value LValue
		if match.IsPosCapture(idx) {
			value = L.GetTable(repl, LNumber(match.Capture(idx)))
		} else {
			value = L.GetField(repl, str[match.Capture(idx):match.Capture(idx+1)])
		}
		if !LVIsFalse(value) {
			infoList = append(infoList, replaceInfo{[]int{match.Capture(0), match.Capture(1)}, LVAsString(value)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

func strGsubFunc(L *LState, str string, repl *LFunction, matches []*pm.MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		L.Push(repl)
		nargs := 0
		if match.CaptureLength() > 2 { // has captures
			for i := 2; i < match.CaptureLength(); i += 2 {
				if match.IsPosCapture(i) {
					L.Push(LNumber(match.Capture(i)))
				} else {
					L.Push(LString(capturedString(L, match, str, i)))
				}
				nargs++
			}
		} else {
			L.Push(LString(capturedString(L, match, str, 0)))
			nargs++
		}
		L.Call(nargs, 1)
		ret := L.reg.Pop()
		if !LVIsFalse(ret) {
			infoList = append(infoList, replaceInfo{[]int{start, end}, LVAsString(ret)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

type strMatchData struct {
	str     string
	pos     int
	matches []*pm.MatchData
}

func strGmatchIter(L *LState) int {
	md := L.CheckUserData(1).Value.(*strMatchData)
	str := md.str
	matches := md.matches
	idx := md.pos
	md.pos++
	if idx == len(matches) {
		return 0
	}
	L.Push(L.Get(1))
	match := matches[idx]
	if match.CaptureLength() == 2 {
		L.Push(LString(str[match.Capture(0):match.Capture(1)]))
		return 1
	}

	for i := 2; i < match.CaptureLength(); i += 2 {
		if match.IsPosCapture(i) {
			L.Push(LNumber(match.Capture(i)))
		} else {
			L.Push(LString(str[match.Capture(i):match.Capture(i+1)]))
		}
	}
	return match.CaptureLength()/2 - 1
}

func strGmatch(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	mds, err := pm.Find(pattern, []byte(str), 0, -1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	L.Push(L.Get(UpvalueIndex(1)))
	ud := L.NewUserData()
	ud.Value = &strMatchData{str, 0, mds}
	L.Push(ud)
	return 2
}

func strLen(L *LState) int {
	str := L.CheckString(1)
	L.Push(LNumber(len(str)))
	return 1
}

func strLower(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(strings.ToLower(str)))
	return 1
}

func strMatch(L *LState) int {
	str := L.CheckString(1)
	pattern := L.CheckString(2)
	offset := L.OptInt(3, 1)
	l := len(str)
	if offset < 0 {
		offset = l + offset + 1
	}
	offset--
	if offset < 0 {
		offset = 0
	}

	mds, err := pm.Find(pattern, unsafeFastStringToReadOnlyBytes(str), offset, 1)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	if len(mds) == 0 {
		L.Push(LNil)
		return 0
	}
	md := mds[0]
	nsubs := md.CaptureLength() / 2
	switch nsubs {
	case 1:
		L.Push(LString(str[md.Capture(0):md.Capture(1)]))
		return 1
	default:
		for i := 2; i < md.CaptureLength(); i += 2 {
			if md.IsPosCapture(i) {
				L.Push(LNumber(md.Capture(i)))
			} else {
				L.Push(LString(str[md.Capture(i):md.Capture(i+1)]))
			}
		}
		return nsubs - 1
	}
}

func strPadEnd(L *LState) int {
	str := L.CheckString(1)
	n := L.CheckInt(2)
	if n <= len(str) {
		L.Push(LString(str))
	} else {
		pad := L.OptString(3, " ")
		L.Push(LString(str + strings.Repeat(pad, n-len(str))))
	}
	return 1
}

func strPadStart(L *LState) int {
	str := L.CheckString(1)
	n := L.CheckInt(2)
	if n <= len(str) {
		L.Push(LString(str))
	} else {
		pad := L.OptString(3, " ")
		L.Push(LString(strings.Repeat(pad, n-len(str)) + str))
	}
	return 1
}

func strRep(L *LState) int {
	str := L.CheckString(1)
	n := L.CheckInt(2)
	if n < 0 {
		L.Push(emptyLString)
	} else {
		L.Push(LString(strings.Repeat(str, n)))
	}
	return 1
}

func strReverse(L *LState) int {
	str := L.CheckString(1)
	bts := []byte(str)
	out := make([]byte, len(bts))
	for i, j := 0, len(bts)-1; j >= 0; i, j = i+1, j-1 {
		out[i] = bts[j]
	}
	L.Push(LString(string(out)))
	return 1
}

func strSplit(L *LState) int {
	str := L.CheckString(1)
	var seps []string
	if L.GetTop() == 1 {
		seps = []string{","}
	} else {
		seps = make([]string, L.GetTop()-1)
		for i := 2; i <= L.GetTop(); i++ {
			seps[i-2] = L.CheckString(i)
		}
	}
	maxSplits := L.OptInt(L.GetTop()+1, -1)
	if maxSplits == 0 {
		L.Push(L.NewTable())
		return 1
	}
	if maxSplits < 0 {
		maxSplits = -1
	}

	tbl := L.NewTable()
	L.Push(tbl)
	start := 0
	index := 1
	for start < len(str) && maxSplits != 0 {
		minPos := len(str)
		var minSep string
		for _, sep := range seps {
			pos := strings.Index(str[start:], sep)
			if pos != -1 && pos < minPos {
				minPos = pos
				minSep = sep
			}
		}
		if minPos == len(str) {
			break
		}
		tbl.RawSetInt(index, LString(str[start:start+minPos]))
		index++
		start += minPos + len(minSep)
		maxSplits--
	}
	if start < len(str) {
		tbl.RawSetInt(index, LString(str[start:]))
	}
	return 1
}

func strStartsWith(L *LState) int {
	str := L.CheckString(1)
	sub := L.CheckString(2)
	if len(sub) > len(str) {
		L.Push(LFalse)
		return 1
	}
	L.Push(LBool(strings.HasPrefix(str, sub)))
	return 1
}

func strSub(L *LState) int {
	str := L.CheckString(1)
	start := luaIndex2StringIndex(str, L.CheckInt(2), true)
	end := luaIndex2StringIndex(str, L.OptInt(3, -1), false)
	l := len(str)
	if start >= l || end < start {
		L.Push(emptyLString)
	} else {
		L.Push(LString(str[start:end]))
	}
	return 1
}

func strTrim(L *LState) int {
	str := L.CheckString(1)
	sep := L.OptString(2, " ")
	str = strings.Trim(str, sep)
	L.Push(LString(str))
	return 1
}

func strTrimEnd(L *LState) int {
	str := L.CheckString(1)
	sep := L.OptString(2, " ")
	str = strings.TrimRight(str, sep)
	L.Push(LString(str))
	return 1
}

func strTrimStart(L *LState) int {
	str := L.CheckString(1)
	sep := L.OptString(2, " ")
	str = strings.TrimLeft(str, sep)
	L.Push(LString(str))
	return 1
}

func strTruncate(L *LState) int {
	str := L.CheckString(1)
	n := L.CheckInt(2)
	suffix := L.OptString(3, "")
	if n < 0 || n >= len(str) {
		L.Push(LString(str))
	} else {
		if len(suffix) > 0 && n > len(suffix) {
			L.Push(LString(str[:n-len(suffix)] + suffix))
		} else {
			L.Push(LString(str[:n]))
		}
	}
	return 1
}

func strUpper(L *LState) int {
	str := L.CheckString(1)
	L.Push(LString(strings.ToUpper(str)))
	return 1
}

func luaIndex2StringIndex(str string, i int, start bool) int {
	if start && i != 0 {
		i -= 1
	}
	l := len(str)
	if i < 0 {
		i = l + i + 1
	}
	i = intMax(0, i)
	if !start && i > l {
		i = l
	}
	return i
}
