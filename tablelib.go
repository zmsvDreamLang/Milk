package lua

import (
	"fmt"
	"sort"
	"strings"
)

// OpenTable registers the table library and pushes it onto the stack.
func OpenTable(L *LState) int {
	tableModule := L.RegisterModule(TabLibName, tableFuncs)
	L.Push(tableModule)
	return 1
}

// tableFuncs maps Lua function names to Go functions.
var tableFuncs = map[string]LGFunction{
	"concat":   tableConcat,
	"equals":   tableEquals,
	"getn":     tableGetN,
	"insert":   tableInsert,
	"maxn":     tableMaxN,
	"remove":   tableRemove,
	"sort":     tableSort,
	"tostring": tableToString,
}

// tableSort sorts the elements of the table in place.
func tableSort(L *LState) int {
	table := L.CheckTable(1)
	sorter := lValueArraySorter{L, nil, table.array}
	if L.GetTop() != 1 {
		sorter.Fn = L.CheckFunction(2)
	}
	sort.Sort(sorter)
	return 0
}

// tableGetN returns the length of the table.
func tableGetN(L *LState) int {
	L.Push(LNumber(L.CheckTable(1).Len()))
	return 1
}

// tableMaxN returns the maximum numerical index in the table.
func tableMaxN(L *LState) int {
	L.Push(LNumber(L.CheckTable(1).MaxN()))
	return 1
}

// tableRemove removes an element from the table at the given position.
func tableRemove(L *LState) int {
	table := L.CheckTable(1)
	if L.GetTop() == 1 {
		L.Push(table.Remove(-1))
	} else {
		L.Push(table.Remove(L.CheckInt(2)))
	}
	return 1
}

// tableConcat concatenates the elements of the table into a string.
func tableConcat(L *LState) int {
	table := L.CheckTable(1)
	separator := LString(L.OptString(2, ""))
	start := L.OptInt(3, 1)
	end := L.OptInt(4, table.Len())

	start = intMax(intMin(start, table.Len()), 1)
	end = intMin(intMin(end, table.Len()), table.Len())
	if start > end {
		L.Push(emptyLString)
		return 1
	}

	var sb strings.Builder
	for ; start <= end; start++ {
		value := table.RawGetInt(start)
		if !LVCanConvToString(value) {
			L.RaiseError("invalid value (%s) at index %d in table for concat", value.Type().String(), start)
		}
		sb.WriteString(LVAsString(value))
		if start != end {
			sb.WriteString(string(separator))
		}
	}
	L.Push(LString(sb.String()))
	return 1
}

// tableInsert inserts an element into the table at the given position.
func tableInsert(L *LState) int {
	table := L.CheckTable(1)
	numArgs := L.GetTop()
	if numArgs == 2 {
		table.Append(L.Get(2))
	} else if numArgs == 3 {
		table.Insert(int(L.CheckInt(2)), L.CheckAny(3))
	} else {
		L.RaiseError("wrong number of arguments")
	}
	return 0
}

// tableToString converts the table to a string representation.
func tableToString(L *LState) int {
	table := L.CheckTable(1)
	result, err := tableToStringRecursive(L, table, "", make(map[*LTable]bool))
	if err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}
	L.Push(LString(result))
	return 1
}

// tableToStringRecursive recursively converts the table to a string representation.
func tableToStringRecursive(L *LState, table *LTable, indent string, visited map[*LTable]bool) (string, error) {
	if visited[table] {
		L.RaiseError("circular references")
		return "", nil
	}
	visited[table] = true

	var sb strings.Builder
	sb.WriteString("{")
	newIndent := indent + "  "

	// Collect keys and sort them
	keys := make([]LValue, 0, table.Len())
	table.ForEach(func(key LValue, value LValue) {
		keys = append(keys, key)
	})

	sort.Slice(keys, func(i, j int) bool {
		ki, kj := keys[i], keys[j]
		if ki.Type() == LTNumber && kj.Type() == LTNumber {
			return LVAsNumber(ki) < LVAsNumber(kj)
		}
		if ki.Type() == LTNumber {
			return true
		}
		if kj.Type() == LTNumber {
			return false
		}
		return LVAsString(ki) < LVAsString(kj)
	})

	for _, key := range keys {
		value := table.RawGet(key)
		keyStr := valueToStringInTable(key)
		valueStr := ""
		if value.Type() == LTTable {
			if visited[value.(*LTable)] {
				valueStr = "circular reference"
			} else {
				var err error
				valueStr, err = tableToStringRecursive(L, value.(*LTable), newIndent, visited)
				if err != nil {
					valueStr = "error"
				}
			}
		} else {
			valueStr = valueToStringInTable(value)
		}
		sb.WriteString(fmt.Sprintf("\n%s[%s] -> %s", newIndent, keyStr, valueStr))
	}

	if sb.Len() > 1 {
		sb.WriteString(fmt.Sprintf("\n%s}", indent))
	} else {
		sb.WriteString("}")
	}

	return sb.String(), nil
}

// valueToStringInTable converts a Lua value to its string representation.
func valueToStringInTable(value LValue) string {
	switch value.Type() {
	case LTNil:
		return "nil"
	case LTBool:
		return fmt.Sprintf("%t", LVAsBool(value))
	case LTNumber:
		return fmt.Sprintf("%v", LVAsNumber(value))
	case LTString:
		return fmt.Sprintf("%q", LVAsString(value))
	case LTFunction:
		return "function"
	case LTUserData:
		return "userdata"
	case LTThread:
		return "thread"
	case LTTable:
		return "table"
	case LTChannel:
		return "channel"
	default:
		return "unknown"
	}
}

func tableEquals(L *LState) int {
	t1 := L.CheckTable(1)
	t2 := L.CheckTable(2)
	visited := make(map[*LTable]bool)
	if deepTableEqual(L, t1, t2, visited) {
		L.Push(LTrue)
	} else {
		L.Push(LFalse)
	}
	return 1
}

func deepTableEqual(L *LState, t1, t2 *LTable, visited map[*LTable]bool) bool {
	if t1 == t2 {
		return true
	}
	if visited[t1] || visited[t2] {
		return true
	}
	if t1.Len() != t2.Len() {
		return false
	}
	visited[t1] = true
	visited[t2] = true
	equal := true
	t1.ForEach(func(key LValue, value LValue) {
		if !Equal(L, value, t2.RawGet(key)) {
			equal = false
			return
		}
		if value.Type() == LTTable && t2.RawGet(key).Type() == LTTable {
			if !deepTableEqual(L, value.(*LTable), t2.RawGet(key).(*LTable), visited) {
				equal = false
				return
			}
		} else if !Equal(L, value, t2.RawGet(key)) {
			equal = false
			return
		}
	})
	return equal
}

func Equal(L *LState, a, b LValue) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch a.Type() {
	case LTNil:
		return true
	case LTBool:
		return LVAsBool(a) == LVAsBool(b)
	case LTNumber:
		return LVAsNumber(a) == LVAsNumber(b)
	case LTString:
		return LVAsString(a) == LVAsString(b)
	case LTTable:
		deepTableEqual(L, a.(*LTable), b.(*LTable), make(map[*LTable]bool))
	default:
		return false
	}
	return false
}
