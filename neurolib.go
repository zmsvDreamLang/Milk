package lua

import (
	"math"
	"math/rand"
)

func OpenNeurolib(L *LState) int {
	mod := L.RegisterModule(NeurolibName, neurolibFuncs)
	L.Push(mod)
	return 1
}

var neurolibFuncs = map[string]LGFunction{
	"relu":       mathRelu,
	"leaky_relu": mathLeakyRelu,
	"sigmoid":    mathSigmoid,
	"elu":        mathElu,

	"mse":           mathMSE,
	"cross_entropy": mathCrossEntropy,
	"hinge_loss":    mathHingeLoss,
	"huber_loss":    mathHuberLoss,

	"l1_regularization": l1Regularization,
	"l2_regularization": l2Regularization,

	"dropout": dropout,
}

func mathRelu(L *LState) int {
	x := float64(L.CheckNumber(1))
	if x > 0 {
		L.Push(LNumber(x))
	} else {
		L.Push(LNumber(0))
	}
	return 1
}

func mathLeakyRelu(L *LState) int {
	x := float64(L.CheckNumber(1))
	alpha := L.OptNumber(2, 0.01)
	if x > 0 {
		L.Push(LNumber(x))
	} else {
		L.Push(LNumber(float64(alpha) * float64(x)))
	}
	return 1
}

func mathSigmoid(L *LState) int {
	x := float64(L.CheckNumber(1))
	L.Push(LNumber(1 / (1 + math.Exp(-x))))
	return 1
}

func mathElu(L *LState) int {
	x := float64(L.CheckNumber(1))
	alpha := L.OptNumber(2, 1.0)
	if x > 0 {
		L.Push(LNumber(x))
	} else {
		L.Push(LNumber(float64(alpha) * (math.Exp(float64(x)) - 1)))
	}
	return 1
}

func mathMSE(L *LState) int {
	y := float64(L.CheckNumber(1))
	yHat := float64(L.CheckNumber(2))
	L.Push(LNumber(math.Pow(y-yHat, 2)))
	return 1
}

func mathCrossEntropy(L *LState) int {
	y := L.CheckTable(1)
	yHat := L.CheckTable(2)

	if y.Len() != yHat.Len() {
		L.RaiseError("Dimensions mismatch in cross_entropy")
		return 0
	}

	total := 0.0
	count := 0

	for i := 1; i <= y.Len(); i++ {
		yRow := y.RawGetInt(i).(*LTable)
		yHatRow := yHat.RawGetInt(i).(*LTable)

		for j := 1; j <= yRow.Len(); j++ {
			yVal := float64(yRow.RawGetInt(j).(LNumber))
			yHatVal := float64(yHatRow.RawGetInt(j).(LNumber))

			// 避免 log(0) 的情况
			yHatVal = math.Max(math.Min(yHatVal, 1-1e-15), 1e-15)

			total += -yVal*math.Log(yHatVal) - (1-yVal)*math.Log(1-yHatVal)
			count++
		}
	}

	L.Push(LNumber(total / float64(count)))
	return 1
}

func mathHingeLoss(L *LState) int {
	y := float64(L.CheckNumber(1))
	yHat := float64(L.CheckNumber(2))
	L.Push(LNumber(math.Max(0, 1-y*yHat)))
	return 1
}

func mathHuberLoss(L *LState) int {
	y := float64(L.CheckNumber(1))
	yHat := float64(L.CheckNumber(2))
	delta := float64(L.OptNumber(3, 1.0))
	absDiff := math.Abs(y - yHat)
	if absDiff <= delta {
		L.Push(LNumber(0.5 * math.Pow(absDiff, 2)))
	} else {
		L.Push(LNumber(delta*absDiff - 0.5*math.Pow(delta, 2)))
	}
	return 1
}

func l1Regularization(L *LState) int {
	weights := float64(L.CheckNumber(1))
	L.Push(LNumber(math.Abs(weights)))
	return 1
}

func l2Regularization(L *LState) int {
	weights := float64(L.CheckNumber(1))
	L.Push(LNumber(0.5 * math.Pow(weights, 2)))
	return 1
}

func dropout(L *LState) int {
	x := L.CheckTable(1)
	keepProb := float64(L.CheckNumber(2))
	seed := L.OptInt64(3, 0)
	r := rand.New(rand.NewSource(seed))
	for i := 1; i <= x.Len(); i++ {
		if r.Float64() < keepProb {
			L.Push(x.RawGetInt(i))
		} else {
			L.Push(LNumber(0))
		}
	}
	return 1
}
