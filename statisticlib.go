package lua

import (
	"fmt"
	"math"
)

// OpenStatistic 注册统计库函数
func OpenStatistic(L *LState) int {
	mod := L.RegisterModule(StatisticLibName, statisticFuncs)

	mt := L.NewTypeMetatable("statistic")
	L.SetField(mt, "__index", L.NewFunction(statisticIndex))
	L.SetField(mt, "__newindex", L.NewFunction(statisticNewIndex))

	// 将方法添加到元表中
	for name, fn := range statisticMethods {
		L.SetField(mt, name, L.NewFunction(fn))
	}

	L.Push(mod)
	return 1
}

// 模块函数，支持简单调用
var statisticFuncs = map[string]LGFunction{
	"new":      statisticNew,
	"mean":     statisticMean,
	"median":   statisticMedian,
	"mode":     statisticMode,
	"stddev":   statisticStddev,
	"variance": statisticVariance,
}

// 方法函数，支持面向对象调用
var statisticMethods = map[string]LGFunction{
	"mean":     statisticMean,
	"median":   statisticMedian,
	"mode":     statisticMode,
	"stddev":   statisticStddev,
	"variance": statisticVariance,
	"add":      statisticAdd,
}

// statisticNew 创建一个新的统计对象
func statisticNew(L *LState) int {
	ud := L.NewUserData()
	ud.Value = make(map[string]float64)

	mt := L.GetTypeMetatable("statistic")
	L.SetMetatable(ud, mt)

	L.Push(ud)
	return 1
}

// statisticIndex 处理统计对象的索引操作
func statisticIndex(L *LState) int {
	ud := L.CheckUserData(1)
	key := L.CheckString(2)

	if method := L.GetField(L.GetMetatable(ud), key); method != LNil {
		L.Push(method)
		return 1
	}

	data, ok := ud.Value.(map[string]float64)
	if !ok {
		L.RaiseError("invalid statistic object")
	}
	if value, exists := data[key]; exists {
		L.Push(LNumber(value))
	} else {
		L.Push(LNil)
	}
	return 1
}

// statisticNewIndex 处理统计对象的新索引操作
func statisticNewIndex(L *LState) int {
	ud := L.CheckUserData(1)
	key := L.CheckString(2)
	value := L.CheckNumber(3)
	data, ok := ud.Value.(map[string]float64)
	if !ok {
		L.RaiseError("invalid statistic object")
	}
	data[key] = float64(value)
	return 0
}

func getValues(L *LState) []float64 {
	var values []float64

	if L.GetTop() == 0 {
		L.Push(LNumber(math.NaN()))
		return nil
	}

	firstArg := L.Get(1)

	switch firstArg.Type() {
	case LTUserData:
		if ud, ok := firstArg.(*LUserData); ok {
			data, ok := ud.Value.(map[string]float64)
			if !ok {
				L.RaiseError("invalid statistic object")
			}
			for _, v := range data {
				values = append(values, v)
			}
		} else {
			L.RaiseError("invalid user data")
		}
	case LTTable:
		tbl := L.CheckTable(1)
		tbl.ForEach(func(key, value LValue) {
			if num, ok := value.(LNumber); ok {
				values = append(values, float64(num))
			} else {
				L.RaiseError("table contains non-number value")
			}
		})
		if len(values) == 0 {
			L.Push(LNumber(math.NaN()))
			return nil
		}
	default:
		for i := 1; i <= L.GetTop(); i++ {
			values = append(values, float64(L.CheckNumber(i)))
		}
	}

	if len(values) == 0 {
		L.Push(LNumber(math.NaN()))
		return nil
	}

	return values
}

// statisticAdd 向统计对象添加数据
func statisticAdd(L *LState) int {
	ud := L.CheckUserData(1)
	data, ok := ud.Value.(map[string]float64)
	if !ok {
		L.RaiseError("invalid statistic object")
	}
	value := L.CheckNumber(2)
	key := fmt.Sprintf("%d", len(data)+1)
	data[key] = float64(value)
	return 0
}

// statisticMean 计算平均值
func statisticMean(L *LState) int {
	values := getValues(L)
	if values == nil {
		return 1 // NaN 已经被 push 到栈上
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	L.Push(LNumber(sum / float64(len(values))))
	return 1
}

// statisticMedian 计算中位数
func statisticMedian(L *LState) int {
	values := getValues(L)
	if values == nil {
		return 1 // NaN 已经被 push 到栈上
	}
	quickSort(values, 0, len(values)-1)
	mid := len(values) / 2
	if len(values)%2 == 0 {
		L.Push(LNumber((values[mid-1] + values[mid]) / 2))
	} else {
		L.Push(LNumber(values[mid]))
	}
	return 1
}

// statisticMode 计算众数
func statisticMode(L *LState) int {
	values := getValues(L)
	if values == nil {
		return 1 // NaN 已经被 push 到栈上
	}
	quickSort(values, 0, len(values)-1)
	var mode float64
	var modeCount, currentCount int
	var current float64
	for i, v := range values {
		if i == 0 || v == current {
			currentCount++
		} else {
			if currentCount > modeCount {
				mode = current
				modeCount = currentCount
			}
			current = v
			currentCount = 1
		}
	}
	if currentCount > modeCount {
		mode = current
	}
	L.Push(LNumber(mode))
	return 1
}

// statisticStddev 计算标准差
func statisticStddev(L *LState) int {
	values := getValues(L)
	if values == nil {
		return 1 // NaN 已经被 push 到栈上
	}
	mean := calculateMean(values)
	variance := calculateVariance(values, mean)
	L.Push(LNumber(math.Sqrt(variance)))
	return 1
}

// statisticVariance 计算方差
func statisticVariance(L *LState) int {
	values := getValues(L)
	if values == nil {
		return 1 // NaN 已经被 push 到栈上
	}
	mean := calculateMean(values)
	variance := calculateVariance(values, mean)
	L.Push(LNumber(variance))
	return 1
}

// calculateMean 计算平均值
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateVariance 计算方差
func calculateVariance(values []float64, mean float64) float64 {
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	return variance / float64(len(values))
}

// quickSort 快速排序算法
func quickSort(values []float64, left, right int) {
	if left < right {
		pivot := partition(values, left, right)
		quickSort(values, left, pivot-1)
		quickSort(values, pivot+1, right)
	}
}

// partition 快速排序的分区函数
func partition(values []float64, left, right int) int {
	pivot := values[right]
	i := left
	for j := left; j < right; j++ {
		if values[j] < pivot {
			values[i], values[j] = values[j], values[i]
			i++
		}
	}
	values[i], values[right] = values[right], values[i]
	return i
}
