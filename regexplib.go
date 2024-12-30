package lua

import (
	regexp "regexp"
)

func OpenRegexp(L *LState) int {
	mod := L.RegisterModule(RegexpLibName, regexpFuncs)
	L.Push(mod)
	return 1
}

var regexpFuncs = map[string]LGFunction{
	"count":    regexpCount,
	"find_all": regexpFindAll,
	"is_match": regexpIsMatch,
	"match":    regexpMatch,
	"replace":  regexpReplace,
	"split":    regexpSplit,
}

func regexpCount(L *LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	matches := re.FindAllStringIndex(str, -1)
	L.Push(LNumber(len(matches)))
	return 1
}

func regexpFindAll(L *LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	matches := re.FindAllString(str, -1)
	result := L.CreateTable(len(matches), 0)
	for i := 0; i < len(matches); i++ {
		result.RawSetInt(i+1, LString(matches[i]))
	}
	L.Push(result)
	return 1
}

func regexpMatch(L *LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	matches := re.FindStringSubmatch(str)
	if matches == nil {
		L.Push(LNil)
		return 1
	}
	paracnt := len(matches) - 1
	L.Push(LString(matches[0]))
	paramtbl := L.CreateTable(paracnt, 0)
	if paracnt > 0 {
		for i := 1; i < len(matches); i++ {
			paramtbl.RawSetInt(i, LString(matches[i]))
		}
		L.Push(paramtbl)
	} else {
		L.Push(LNil)
	}
	return 2
}

func regexpIsMatch(L *LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	matched := re.MatchString(str)
	if matched {
		L.Push(LTrue)
	} else {
		L.Push(LFalse)
	}
	return 1
}

func regexpReplace(L *LState) int {
	pattern := L.CheckString(1)
	replacement := L.CheckString(2)
	str := L.CheckString(3)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	result := re.ReplaceAllString(str, replacement)
	L.Push(LString(result))
	return 1
}

func regexpSplit(L *LState) int {
	pattern := L.CheckString(1)
	str := L.CheckString(2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		L.RaiseError("%s", err.Error())
	}
	splits := re.Split(str, -1)
	result := L.CreateTable(len(splits), 0)
	for i := 0; i < len(splits); i++ {
		result.RawSetInt(i+1, LString(splits[i]))
	}
	L.Push(result)
	return 1
}
