package lua

import "math"

func OpenStatistic(L *LState) int {
	mod := L.RegisterModule(StatisticLibName, statisticFuncs)
	L.Push(mod)
	return 1
}

var statisticFuncs = map[string]LGFunction{
	"mean":     statisticMean,
	"median":   statisticMedian,
	"mode":     statisticMode,
	"stddev":   statisticStddev,
	"variance": statisticVariance,
}

func statisticMean(L *LState) int {
	var sum float64
	var count float64
	for i := 1; i <= L.GetTop(); i++ {
		sum += float64(L.CheckNumber(i))
		count++
	}
	L.Push(LNumber(sum / count))
	return 1
}

func statisticMedian(L *LState) int {
	var values []float64
	for i := 1; i <= L.GetTop(); i++ {
		values = append(values, float64(L.CheckNumber(i)))
	}
	quickSort(values, 0, len(values)-1)
	if len(values)%2 == 0 {
		L.Push(LNumber((values[len(values)/2-1] + values[len(values)/2]) / 2))
	} else {
		L.Push(LNumber(values[len(values)/2]))
	}
	return 1
}

func statisticMode(L *LState) int {
	var values []float64
	for i := 1; i <= L.GetTop(); i++ {
		values = append(values, float64(L.CheckNumber(i)))
	}
	quickSort(values, 0, len(values)-1)
	var mode float64
	var modeCount int
	var current float64
	var currentCount int
	for i := 0; i < len(values); i++ {
		if values[i] == current {
			currentCount++
		} else {
			if currentCount > modeCount {
				mode = current
				modeCount = currentCount
			}
			current = values[i]
			currentCount = 1
		}
	}
	if currentCount > modeCount {
		mode = current
		modeCount = currentCount
	}
	L.Push(LNumber(mode))
	return 1
}

func statisticStddev(L *LState) int {
	var sum float64
	var count float64
	for i := 1; i <= L.GetTop(); i++ {
		sum += float64(L.CheckNumber(i))
		count++
	}
	mean := sum / count
	var variance float64
	for i := 1; i <= L.GetTop(); i++ {
		v := float64(L.CheckNumber(i))
		variance += (v - mean) * (v - mean)
	}
	L.Push(LNumber(math.Sqrt(variance / count)))
	return 1
}

func statisticVariance(L *LState) int {
	var sum float64
	var count float64
	for i := 1; i <= L.GetTop(); i++ {
		sum += float64(L.CheckNumber(i))
		count++
	}
	mean := sum / count
	var variance float64
	for i := 1; i <= L.GetTop(); i++ {
		v := float64(L.CheckNumber(i))
		variance += (v - mean) * (v - mean)
	}
	L.Push(LNumber(variance / count))
	return 1
}

func quickSort(values []float64, left, right int) {
	if left < right {
		pivot := partition(values, left, right)
		quickSort(values, left, pivot-1)
		quickSort(values, pivot+1, right)
	}
}

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
