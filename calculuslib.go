package lua

import (
	"math"
)

func OpenCalculus(L *LState) int {
	mod := L.RegisterModule(CalculusLibName, calculusFuncs)

	L.Push(mod)
	return 1
}

var calculusFuncs = map[string]LGFunction{
	"derivative": calculusDerivative,
	"integral":   calculusIntegral,
	"autodiff":   calculusAutoDiff,

	"Dual": newDual,
	"sin":  dualSin,
	"cos":  dualCos,
	"exp":  dualExp,
	"log":  dualLog,
	"pow":  dualPow,
}

func calculusDerivative(L *LState) int {
	fun := L.CheckFunction(1)
	x := L.CheckNumber(2)
	delta := L.OptNumber(3, 1e-8)
	precision := L.OptNumber(4, 1e-8)
	maxIterations := L.OptInt(5, 10)

	if delta == 0 {
		L.RaiseError("delta cannot be zero")
		return 0
	}

	derivative := func(x, h float64) float64 {
		L.Push(fun)
		L.Push(LNumber(x + h))
		L.Call(1, 1)
		fx_plus_h := float64(L.CheckNumber(-1))

		L.Push(fun)
		L.Push(LNumber(x - h))
		L.Call(1, 1)
		fx_minus_h := float64(L.CheckNumber(-1))

		return (fx_plus_h - fx_minus_h) / (2 * h)
	}

	// Adaptive step size method
	h := float64(delta)
	prevResult := derivative(float64(x), h)
	for i := 0; i < maxIterations; i++ {
		h /= 2
		result := derivative(float64(x), h)
		if math.Abs(result-prevResult) < float64(precision) {
			L.Push(LNumber(result))
			return 1 // 成功时返回 1
		}
		prevResult = result
	}

	L.Push(LNumber(prevResult)) // 即使没有达到期望精度，也返回最后的结果
	L.Push(LString("derivative calculation did not converge to desired precision"))
	return 2 // 返回结果和错误信息
}

func calculusIntegral(L *LState) int {
	fun := L.CheckFunction(1)
	a := L.CheckNumber(2)
	b := L.CheckNumber(3)
	n := L.OptInt(4, 1000)
	precision := L.OptNumber(5, 1e-8)
	maxIterations := L.OptInt(6, 10)

	if a >= b {
		L.Push(LNil)
		L.Push(LString("lower bound must be less than upper bound"))
		return 2
	}

	if n <= 0 {
		L.Push(LNil)
		L.Push(LString("number of intervals must be positive"))
		return 2
	}

	f := func(x float64) float64 {
		L.Push(fun)
		L.Push(LNumber(x))
		L.Call(1, 1)
		return float64(L.CheckNumber(-1))
	}

	result := adaptiveSimpson(f, float64(a), float64(b), maxIterations, float64(precision))
	L.Push(LNumber(result))
	L.Push(LNil) // 没有错误，所以第二个返回值为 nil
	return 2     // 返回两个值
}

func adaptiveSimpson(f func(float64) float64, a, b float64, maxDepth int, epsilon float64) float64 {
	var asr func(float64, float64, float64, float64, int) float64
	asr = func(a, b, fa, fb float64, depth int) float64 {
		m := (a + b) / 2
		fm := f(m)
		left := (b - a) / 6 * (fa + 4*fm + fb)
		flm := f((a + m) / 2)
		frm := f((m + b) / 2)
		right := (b - a) / 12 * (fa + 4*flm + 2*fm + 4*frm + fb)
		if depth >= maxDepth {
			return right
		}
		if math.Abs(right-left) <= 15*epsilon {
			return right
		}
		return asr(a, m, fa, fm, depth+1) + asr(m, b, fm, fb, depth+1)
	}

	fa, fb := f(a), f(b)
	return asr(a, b, fa, fb, 0)
}

type Dual struct {
	Value, Derivative float64
}

func newDual(L *LState) int {
	value := float64(L.CheckNumber(1))
	derivative := float64(L.OptNumber(2, 1))
	ud := L.NewUserData()
	ud.Value = Dual{Value: value, Derivative: derivative}
	L.Push(ud)
	return 1
}

func getDual(L *LState, index int) Dual {
	ud := L.CheckUserData(index)
	return ud.Value.(Dual)
}

func pushDual(L *LState, d Dual) {
	ud := L.NewUserData()
	ud.Value = d
	L.Push(ud)
}

func dualSin(L *LState) int {
	d := getDual(L, 1)
	result := Sin(d)
	pushDual(L, result)
	return 1
}

func dualCos(L *LState) int {
	d := getDual(L, 1)
	result := Cos(d)
	pushDual(L, result)
	return 1
}

func dualExp(L *LState) int {
	d := getDual(L, 1)
	result := Exp(d)
	pushDual(L, result)
	return 1
}

func dualLog(L *LState) int {
	d := getDual(L, 1)
	result := Log(d)
	pushDual(L, result)
	return 1
}

func dualPow(L *LState) int {
	d := getDual(L, 1)
	n := L.CheckNumber(2)
	result := Pow(d, float64(n))
	pushDual(L, result)
	return 1
}

func calculusAutoDiff(L *LState) int {
	fun := L.CheckFunction(1)
	x := L.CheckNumber(2)
	ud := L.NewUserData()
	ud.Value = Dual{Value: float64(x), Derivative: 1}

	L.Push(fun)
	L.Push(ud)
	L.Call(1, 1)

	result := getDual(L, -1)
	L.Push(LNumber(result.Value))
	L.Push(LNumber(result.Derivative))
	return 2
}

func autoDiff(f func(Dual) Dual, x float64) Dual {
	return f(Dual{Value: x, Derivative: 1})
}

// 基本运算
func (d Dual) Add(other Dual) Dual {
	return Dual{d.Value + other.Value, d.Derivative + other.Derivative}
}

func (d Dual) Mul(other Dual) Dual {
	return Dual{
		d.Value * other.Value,
		d.Value*other.Derivative + d.Derivative*other.Value,
	}
}

func (d Dual) Div(other Dual) Dual {
	return Dual{
		d.Value / other.Value,
		(d.Derivative*other.Value - d.Value*other.Derivative) / (other.Value * other.Value),
	}
}

// 初等函数
func Sin(d Dual) Dual {
	return Dual{
		math.Sin(d.Value),
		d.Derivative * math.Cos(d.Value),
	}
}

func Cos(d Dual) Dual {
	return Dual{
		math.Cos(d.Value),
		-d.Derivative * math.Sin(d.Value),
	}
}

func Exp(d Dual) Dual {
	ex := math.Exp(d.Value)
	return Dual{ex, d.Derivative * ex}
}

func Log(d Dual) Dual {
	return Dual{
		math.Log(d.Value),
		d.Derivative / d.Value,
	}
}

func Pow(d Dual, n float64) Dual {
	return Dual{
		math.Pow(d.Value, n),
		n * math.Pow(d.Value, n-1) * d.Derivative,
	}
}
