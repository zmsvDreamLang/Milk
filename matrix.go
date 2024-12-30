package lua

import (
	"math"
	"math/rand"
	"time"
)

func OpenMatrix(L *LState) int {
	mod := L.RegisterModule(MatrixLibName, matrixFuncs)
	L.Push(mod)
	return 1
}

var matrixFuncs = map[string]LGFunction{
	"add":            matrixAdd,
	"add_bias":       matrixAddBias,
	"apply":          matrixApply,
	"sub":            matrixSub,
	"sum":            matrixSum,
	"mul":            matrixMul,
	"div":            matrixDiv,
	"dot":            matrixDot,
	"cross":          matrixCross,
	"hadamard":       matrixHadamard,
	"transpose":      matrixTranspose,
	"det":            matrixDet,
	"inv":            matrixInv,
	"cofactor":       matrixCofactor,
	"scale":          matrixScale,
	"random":         matrixRandom,
	"trace":          matrixTrace,
	"eigenvalues":    matrixEigenvalues,
	"lu":             matrixLU,
	"lup":            matrixLUP,
	"qr":             matrixQR,
	"solve":          matrixSolve,
	"rref":           matrixRREF,
	"nullspace":      matrixNullspace,
	"rowspace":       matrixRowSpace,
	"orthogonalize":  matrixOrthogonalize,
	"orthonormalize": matrixOrthonormalize,
	"gram_schmidt":   matrixGramSchmidt,
	"householder":    matrixHouseholder,
	"givens":         matrixGivens,
	"jacobi":         matrixJacobi,
	"lanczos":        matrixLanczos,
	"power":          matrixPower,
}

func matrixAdd(L *LState) int {
	matrix1 := L.CheckTable(1)
	matrix2 := L.CheckTable(2)

	if matrix1 == nil || matrix2 == nil {
		L.RaiseError("One of the input matrices is nil")
		return 0
	}

	result := L.NewTable()
	m1Rows := matrix1.Len()

	isMatrix2Vector := matrix2.RawGetInt(1).Type() != LTTable

	for i := 1; i <= m1Rows; i++ {
		row1 := matrix1.RawGetInt(i).(*LTable)
		resultRow := L.NewTable()

		for j := 1; j <= row1.Len(); j++ {
			var value2 LNumber
			if isMatrix2Vector {
				value2 = matrix2.RawGetInt(j).(LNumber)
			} else {
				row2 := matrix2.RawGetInt((i-1)%matrix2.Len() + 1).(*LTable)
				value2 = row2.RawGetInt(j).(LNumber)
			}

			sum := LNumber(row1.RawGetInt(j).(LNumber) + value2)
			resultRow.RawSetInt(j, sum)
		}
		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixAddBias(L *LState) int {
	matrix := L.CheckTable(1)
	bias := L.CheckTable(2)

	if matrix == nil || bias == nil {
		L.RaiseError("Matrix or bias vector is nil")
		return 0
	}

	result := L.NewTable()
	mRows := matrix.Len()

	for i := 1; i <= mRows; i++ {
		row := matrix.RawGetInt(i).(*LTable)
		resultRow := L.NewTable()

		for j := 1; j <= row.Len(); j++ {
			matrixValue := row.RawGetInt(j).(LNumber)
			biasValue := bias.RawGetInt(j).(LNumber)

			sum := LNumber(float64(matrixValue) + float64(biasValue))
			resultRow.RawSetInt(j, sum)
		}
		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixApply(L *LState) int {
	matrix := L.CheckTable(1)
	fn := L.CheckFunction(2)

	if matrix == nil {
		L.RaiseError("Input matrix is nil")
		return 0
	}

	result := L.NewTable()

	for i := 1; i <= matrix.Len(); i++ {
		row := matrix.RawGetInt(i).(*LTable)
		resultRow := L.NewTable()

		for j := 1; j <= row.Len(); j++ {
			L.Push(fn)
			L.Push(row.RawGetInt(j))
			L.Call(1, 1)
			resultRow.RawSetInt(j, L.Get(-1))
			L.Pop(1)
		}

		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixSub(L *LState) int {
	a := L.CheckAny(1)
	b := L.CheckAny(2)

	switch av := a.(type) {
	case *LTable:
		switch bv := b.(type) {
		case *LTable:
			return subtractTables(L, av, bv)
		case LNumber:
			return subtractTableNumber(L, av, bv)
		default:
			L.RaiseError("Unsupported type for subtraction: %T and %T", a, b)
		}
	case LNumber:
		switch bv := b.(type) {
		case *LTable:
			return subtractNumberTable(L, av, bv)
		case LNumber:
			L.Push(LNumber(float64(av) - float64(bv)))
			return 1
		default:
			L.RaiseError("Unsupported type for subtraction: %T and %T", a, b)
		}
	default:
		L.RaiseError("Unsupported type for subtraction: %T", a)
	}

	return 0
}

func subtractTables(L *LState, a, b *LTable) int {
	if a.Len() != b.Len() {
		L.RaiseError("Matrix dimensions do not match for subtraction")
		return 0
	}

	result := L.NewTable()
	for i := 1; i <= a.Len(); i++ {
		aRow := a.RawGetInt(i)
		bRow := b.RawGetInt(i)

		switch ar := aRow.(type) {
		case *LTable:
			switch br := bRow.(type) {
			case *LTable:
				if ar.Len() != br.Len() {
					L.RaiseError("Matrix dimensions do not match for subtraction")
					return 0
				}
				resultRow := L.NewTable()
				for j := 1; j <= ar.Len(); j++ {
					aVal := ar.RawGetInt(j).(LNumber)
					bVal := br.RawGetInt(j).(LNumber)
					resultRow.RawSetInt(j, LNumber(float64(aVal)-float64(bVal)))
				}
				result.RawSetInt(i, resultRow)
			default:
				L.RaiseError("Inconsistent matrix structure")
				return 0
			}
		case LNumber:
			switch br := bRow.(type) {
			case LNumber:
				result.RawSetInt(i, LNumber(float64(ar)-float64(br)))
			default:
				L.RaiseError("Inconsistent matrix structure")
				return 0
			}
		default:
			L.RaiseError("Unsupported type in matrix: %T", ar)
			return 0
		}
	}

	L.Push(result)
	return 1
}

func subtractTableNumber(L *LState, table *LTable, number LNumber) int {
	result := L.NewTable()
	for i := 1; i <= table.Len(); i++ {
		row := table.RawGetInt(i)
		switch r := row.(type) {
		case *LTable:
			resultRow := L.NewTable()
			for j := 1; j <= r.Len(); j++ {
				val := r.RawGetInt(j).(LNumber)
				resultRow.RawSetInt(j, LNumber(float64(val)-float64(number)))
			}
			result.RawSetInt(i, resultRow)
		case LNumber:
			result.RawSetInt(i, LNumber(float64(r)-float64(number)))
		default:
			L.RaiseError("Unsupported type in matrix: %T", r)
			return 0
		}
	}
	L.Push(result)
	return 1
}

func subtractNumberTable(L *LState, number LNumber, table *LTable) int {
	result := L.NewTable()
	for i := 1; i <= table.Len(); i++ {
		row := table.RawGetInt(i)
		switch r := row.(type) {
		case *LTable:
			resultRow := L.NewTable()
			for j := 1; j <= r.Len(); j++ {
				val := r.RawGetInt(j).(LNumber)
				resultRow.RawSetInt(j, LNumber(float64(number)-float64(val)))
			}
			result.RawSetInt(i, resultRow)
		case LNumber:
			result.RawSetInt(i, LNumber(float64(number)-float64(r)))
		default:
			L.RaiseError("Unsupported type in matrix: %T", r)
			return 0
		}
	}
	L.Push(result)
	return 1
}

func matrixSum(L *LState) int {
	matrix := L.CheckTable(1)
	axis := L.OptInt(2, 0) // 默认沿着所有轴求和

	if matrix == nil {
		L.RaiseError("Input matrix is nil")
		return 0
	}

	if axis != 0 && axis != 1 {
		L.RaiseError("Invalid axis. Must be 0 (all axes) or 1 (column-wise)")
		return 0
	}

	rows := matrix.Len()
	if rows == 0 {
		L.Push(L.NewTable())
		return 1
	}

	cols := matrix.RawGetInt(1).(*LTable).Len()

	result := L.NewTable()

	if axis == 0 {
		// 沿所有轴求和
		sum := LNumber(0)
		for i := 1; i <= rows; i++ {
			row := matrix.RawGetInt(i).(*LTable)
			for j := 1; j <= cols; j++ {
				sum += row.RawGetInt(j).(LNumber)
			}
		}
		result.RawSetInt(1, L.NewTable())
		result.RawGetInt(1).(*LTable).RawSetInt(1, sum)
	} else {
		// 沿列求和
		for j := 1; j <= cols; j++ {
			colSum := LNumber(0)
			for i := 1; i <= rows; i++ {
				row := matrix.RawGetInt(i).(*LTable)
				colSum += row.RawGetInt(j).(LNumber)
			}
			result.RawSetInt(j, colSum)
		}
	}

	L.Push(result)
	return 1
}

func matrixHadamard(L *LState) int {
	matrix1 := L.CheckTable(1)
	matrix2 := L.CheckTable(2)

	if matrix1 == nil || matrix2 == nil {
		L.RaiseError("One of the input matrices is nil")
		return 0
	}

	rows1 := matrix1.Len()
	cols1 := matrix1.RawGetInt(1).(*LTable).Len()
	rows2 := matrix2.Len()
	cols2 := matrix2.RawGetInt(1).(*LTable).Len()

	if rows1 != rows2 || cols1 != cols2 {
		L.RaiseError("Matrix dimensions do not match for Hadamard product: %dx%d and %dx%d", rows1, cols1, rows2, cols2)
		return 0
	}

	result := L.NewTable()
	for i := 1; i <= rows1; i++ {
		row1 := matrix1.RawGetInt(i).(*LTable)
		row2 := matrix2.RawGetInt(i).(*LTable)
		resultRow := L.NewTable()
		for j := 1; j <= cols1; j++ {
			value := LNumber(row1.RawGetInt(j).(LNumber) * row2.RawGetInt(j).(LNumber))
			resultRow.RawSetInt(j, value)
		}
		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixMul(L *LState) int {
	matrix1 := L.CheckTable(1)
	matrix2 := L.CheckTable(2)

	if matrix1 == nil || matrix2 == nil {
		L.RaiseError("One of the input matrices is nil")
		return 0
	}

	rows1 := matrix1.Len()
	cols1 := matrix1.RawGetInt(1).(*LTable).Len()
	rows2 := matrix2.Len()
	cols2 := matrix2.RawGetInt(1).(*LTable).Len()

	if cols1 != rows2 {
		L.RaiseError("Matrix dimensions do not match for multiplication: %dx%d and %dx%d", rows1, cols1, rows2, cols2)
		return 0
	}

	result := L.NewTable()
	for i := 1; i <= rows1; i++ {
		resultRow := L.NewTable()
		for j := 1; j <= cols2; j++ {
			sum := LNumber(0)
			for k := 1; k <= cols1; k++ {
				a := matrix1.RawGetInt(i).(*LTable).RawGetInt(k).(LNumber)
				b := matrix2.RawGetInt(k).(*LTable).RawGetInt(j).(LNumber)
				sum += a * b
			}
			resultRow.RawSetInt(j, sum)
		}
		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixDiv(L *LState) int {
	matrix1 := L.CheckTable(1)
	matrix2 := L.CheckTable(2)

	result := L.NewTable()

	for i := 1; i <= matrix1.Len(); i++ {
		row1 := matrix1.RawGetInt(i).(*LTable)
		row2 := matrix2.RawGetInt(i).(*LTable)

		if row1.Len() != row2.Len() {
			L.RaiseError("Matrix dimensions do not match")
		}

		resultRow := L.NewTable()
		for j := 1; j <= row1.Len(); j++ {
			divisor := row2.RawGetInt(j).(LNumber)
			if divisor == 0 {
				L.RaiseError("Division by zero")
			}
			div := LNumber(row1.RawGetInt(j).(LNumber) / divisor)
			resultRow.RawSetInt(j, div)
		}
		result.RawSetInt(i, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixDot(L *LState) int {
	matrix1 := L.CheckTable(1)
	matrix2 := L.CheckTable(2)

	result := LNumber(0)

	for i := 1; i <= matrix1.Len(); i++ {
		row1 := matrix1.RawGetInt(i).(*LTable)
		row2 := matrix2.RawGetInt(i).(*LTable)

		for j := 1; j <= row1.Len(); j++ {
			result += row1.RawGetInt(j).(LNumber) * row2.RawGetInt(j).(LNumber)
		}
	}

	L.Push(result)
	return 1
}

func matrixCross(L *LState) int {
	vec1 := L.CheckTable(1)
	vec2 := L.CheckTable(2)

	if vec1.Len() != 3 || vec2.Len() != 3 {
		L.RaiseError("Cross product is only defined for 3D vectors")
	}

	result := L.NewTable()

	x1, y1, z1 := vec1.RawGetInt(1).(LNumber), vec1.RawGetInt(2).(LNumber), vec1.RawGetInt(3).(LNumber)
	x2, y2, z2 := vec2.RawGetInt(1).(LNumber), vec2.RawGetInt(2).(LNumber), vec2.RawGetInt(3).(LNumber)

	result.RawSetInt(1, y1*z2-z1*y2)
	result.RawSetInt(2, z1*x2-x1*z2)
	result.RawSetInt(3, x1*y2-y1*x2)

	L.Push(result)
	return 1
}

func matrixTranspose(L *LState) int {
	matrix := L.CheckTable(1)

	if matrix == nil {
		L.RaiseError("Input matrix is nil")
		return 0
	}

	rows := matrix.Len()
	if rows == 0 {
		L.Push(L.NewTable())
		return 1
	}

	cols := matrix.RawGetInt(1).(*LTable).Len()

	result := L.NewTable()
	for j := 1; j <= cols; j++ {
		resultRow := L.NewTable()
		for i := 1; i <= rows; i++ {
			row := matrix.RawGetInt(i).(*LTable)
			resultRow.RawSetInt(i, row.RawGetInt(j))
		}
		result.RawSetInt(j, resultRow)
	}

	L.Push(result)
	return 1
}

func matrixDet(L *LState) int {
	matrix := L.CheckTable(1)
	LU, P := lupDecomposition(L, matrix)
	det := LNumber(1)
	n := matrix.Len()
	for i := 1; i <= n; i++ {
		det *= LU.RawGetInt(i).(*LTable).RawGetInt(i).(LNumber)
	}
	// 考虑置换矩阵的符号
	if countSwaps(P)%2 != 0 {
		det = -det
	}
	L.Push(LNumber(det))
	return 1
}

func countSwaps(P *LTable) int {
	swaps := 0
	n := P.Len()
	for i := 1; i <= n; i++ {
		if P.RawGetInt(i).(LNumber) != LNumber(i) {
			swaps++
		}
	}
	return swaps
}

func matrixInv(L *LState) int {
	matrix := L.CheckTable(1)
	n := matrix.Len()

	// 使用LU分解求逆矩阵
	LU, P := lupDecomposition(L, matrix)
	inv := L.NewTable()

	// 检查矩阵是否可逆
	for i := 1; i <= n; i++ {
		if LU.RawGetInt(i).(*LTable).RawGetInt(i).(LNumber) == 0 {
			L.RaiseError("Matrix is not invertible")
			return 0
		}
	}

	for j := 1; j <= n; j++ {
		b := L.NewTable()
		for i := 1; i <= n; i++ {
			if P.RawGetInt(i).(LNumber) == LNumber(j) {
				b.RawSetInt(i, LNumber(1))
			} else {
				b.RawSetInt(i, LNumber(0))
			}
		}
		x := luSolve(L, LU, b)
		for i := 1; i <= n; i++ {
			if inv.RawGetInt(i) == LNil {
				inv.RawSetInt(i, L.NewTable())
			}
			inv.RawGetInt(i).(*LTable).RawSetInt(j, x.RawGetInt(i))
		}
	}

	L.Push(inv)
	return 1
}

func luDecomposition(L *LState, matrix *LTable) (*LTable, *LTable) {
	n := matrix.Len()

	// 初始化LU矩阵
	LU := copyMatrix(L, matrix)

	for k := 1; k <= n-1; k++ {
		for i := k + 1; i <= n; i++ {
			LU.RawGetInt(i).(*LTable).RawSetInt(k, LU.RawGetInt(i).(*LTable).RawGetInt(k).(LNumber)/LU.RawGetInt(k).(*LTable).RawGetInt(k).(LNumber))
			for j := k + 1; j <= n; j++ {
				LU.RawGetInt(i).(*LTable).RawSetInt(j, LU.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber)-LU.RawGetInt(i).(*LTable).RawGetInt(k).(LNumber)*LU.RawGetInt(k).(*LTable).RawGetInt(j).(LNumber))
			}
		}
	}

	return LU, nil
}

func luSolve(L *LState, LU, b *LTable) *LTable {
	n := LU.Len()

	// 使用LU分解求解线性方程组
	y := L.NewTable()
	for i := 1; i <= n; i++ {
		sum := LNumber(0)
		for j := 1; j < i; j++ {
			sum += LU.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber) * y.RawGetInt(j).(LNumber)
		}
		y.RawSetInt(i, (b.RawGetInt(i).(LNumber) - sum))
	}

	x := L.NewTable()
	for i := n; i >= 1; i-- {
		sum := LNumber(0)
		for j := i + 1; j <= n; j++ {
			sum += LU.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber) * x.RawGetInt(j).(LNumber)
		}
		x.RawSetInt(i, (y.RawGetInt(i).(LNumber)-sum)/LU.RawGetInt(i).(*LTable).RawGetInt(i).(LNumber))
	}

	return x
}

func lupDecomposition(L *LState, matrix *LTable) (*LTable, *LTable) {
	n := matrix.Len()

	// 初始化LU矩阵
	LU := copyMatrix(L, matrix)
	P := L.NewTable()
	for i := 1; i <= n; i++ {
		P.RawSetInt(i, LNumber(i))
	}

	for k := 1; k <= n-1; k++ {
		// 部分主元法选取主元
		maxVal := LNumber(0)
		maxRow := 0
		for i := k; i <= n; i++ {
			val := LNumber(math.Abs(float64(LU.RawGetInt(i).(*LTable).RawGetInt(k).(LNumber))))
			if val > maxVal {
				maxVal = val
				maxRow = i
			}
		}

		if maxVal == 0 {
			L.RaiseError("Matrix is singular")
		}

		// 交换行
		if maxRow != k {
			tempRow := LU.RawGetInt(k)
			LU.RawSetInt(k, LU.RawGetInt(maxRow))
			LU.RawSetInt(maxRow, tempRow)
			temp := P.RawGetInt(k)
			P.RawSetInt(k, P.RawGetInt(maxRow))
			P.RawSetInt(maxRow, temp)
		}

		// 计算消元因子
		for i := k + 1; i <= n; i++ {
			factor := LU.RawGetInt(i).(*LTable).RawGetInt(k).(LNumber) / LU.RawGetInt(k).(*LTable).RawGetInt(k).(LNumber)
			LU.RawGetInt(i).(*LTable).RawSetInt(k, factor)
			for j := k + 1; j <= n; j++ {
				LU.RawGetInt(i).(*LTable).RawSetInt(j, LU.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber)-factor*LU.RawGetInt(k).(*LTable).RawGetInt(j).(LNumber))
			}
		}
	}

	return LU, P
}

func matrixCofactor(L *LState) int {
	matrix := L.CheckTable(1)
	n := matrix.Len()

	cofactor := L.NewTable()
	for i := 1; i <= n; i++ {
		cofactor.RawSetInt(i, L.NewTable())
		for j := 1; j <= n; j++ {
			minor := calculateMinor(L, matrix, i, j)
			sign := LNumber(math.Pow(-1, float64(i+j)))
			cofactor.RawGetInt(i).(*LTable).RawSetInt(j, sign*minor)
		}
	}

	L.Push(cofactor)
	return 1
}

func calculateMinor(L *LState, matrix *LTable, row, col int) LNumber {
	minorMatrix := L.NewTable()
	n := matrix.Len()

	for i := 1; i <= n; i++ {
		if i == row {
			continue
		}

		minorRow := L.NewTable()
		for j := 1; j <= n; j++ {
			if j == col {
				continue
			}
			minorRow.RawSetInt(j, matrix.RawGetInt(i).(*LTable).RawGetInt(j))
		}
		minorMatrix.RawSetInt(minorMatrix.Len()+1, minorRow)
	}

	minor := matrixDet(L)
	return LNumber(minor)
}

func matrixScale(L *LState) int {
	input := L.CheckAny(1)
	scalar := L.OptNumber(2, 1)

	switch v := input.(type) {
	case LNumber:
		L.Push(LNumber(float64(v) * float64(scalar)))
	case *LTable:
		result := L.NewTable()
		for i := 1; i <= v.Len(); i++ {
			row := v.RawGetInt(i)
			switch r := row.(type) {
			case LNumber:
				result.RawSetInt(i, LNumber(float64(r)*float64(scalar)))
			case *LTable:
				resultRow := L.NewTable()
				for j := 1; j <= r.Len(); j++ {
					if value, ok := r.RawGetInt(j).(LNumber); ok {
						resultRow.RawSetInt(j, LNumber(float64(value)*float64(scalar)))
					} else {
						L.RaiseError("Unsupported type in matrix at position [%d][%d]", i, j)
						return 0
					}
				}
				result.RawSetInt(i, resultRow)
			default:
				L.RaiseError("Unsupported type in matrix at row %d", i)
				return 0
			}
		}
		L.Push(result)
	default:
		L.RaiseError("Unsupported input type for scale: %T", input)
		return 0
	}

	return 1
}

func matrixRandom(L *LState) int {
	rows := L.CheckInt(1)
	cols := L.CheckInt(2)

	// 初始化随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建新的Lua表作为矩阵
	matrix := L.NewTable()

	for i := 1; i <= rows; i++ {
		row := L.NewTable()
		for j := 1; j <= cols; j++ {
			// 生成-1到1之间的随机浮点数
			value := r.Float64()*2 - 1
			row.RawSetInt(j, LNumber(value))
		}
		matrix.RawSetInt(i, row)
	}

	L.Push(matrix)
	return 1
}

func matrixTrace(L *LState) int {
	matrix := L.CheckTable(1)

	trace := LNumber(0)
	for i := 1; i <= matrix.Len(); i++ {
		trace += matrix.RawGetInt(i).(*LTable).RawGetInt(i).(LNumber)
	}

	L.Push(trace)
	return 1
}

func matrixEigenvalues(L *LState) int {
	matrix := L.CheckTable(1)
	n := matrix.Len()

	// 先进行 Hessenberg 变换
	H := hessenbergReduction(L, matrix)

	// 使用 QR 算法计算特征值
	eigenvalues := L.NewTable()
	for k := 1; k <= 100; k++ { // 最多迭代100次
		Q, R := qrDecomposition(L, H)
		H = matrixMultiply(L, R, Q)

		// 检查对角线元素是否收敛
		if isConverged(H) {
			break
		}
	}

	// 提取对角线元素作为特征值
	for i := 1; i <= n; i++ {
		eigenvalues.RawSetInt(i, H.RawGetInt(i).(*LTable).RawGetInt(i))
	}

	L.Push(eigenvalues)
	return 1
}

func hessenbergReduction(L *LState, matrix *LTable) *LTable {
	n := matrix.Len()

	H := copyMatrix(L, matrix)
	for k := 1; k <= n-2; k++ {
		// 计算Householder变换矩阵
		vector := L.NewTable()
		for i := 1; i <= n; i++ {
			if i < k+2 {
				vector.RawSetInt(i, LNumber(0))
			} else {
				vector.RawSetInt(i, H.RawGetInt(i).(*LTable).RawGetInt(k+1))
			}
		}
		H := calculateHouseholderMatrix(L, vector)

		// 计算Hessenberg矩阵
		H = matrixMultiply(L, H, H)
	}

	return H
}

func qrDecomposition(L *LState, matrix *LTable) (*LTable, *LTable) {
	n := matrix.Len()

	Q := L.NewTable()
	for i := 1; i <= n; i++ {
		Q.RawSetInt(i, L.NewTable())
		for j := 1; j <= n; j++ {
			Q.RawGetInt(i).(*LTable).RawSetInt(j, LNumber(0))
		}
		Q.RawGetInt(i).(*LTable).RawSetInt(i, LNumber(1))
	}

	for k := 1; k <= n-1; k++ {
		// 计算Householder变换矩阵
		vector := L.NewTable()
		for i := 1; i <= n; i++ {
			if i < k {
				vector.RawSetInt(i, LNumber(0))
			} else {
				vector.RawSetInt(i, matrix.RawGetInt(i).(*LTable).RawGetInt(k))
			}
		}
		H := calculateHouseholderMatrix(L, vector)

		// 计算QR分解
		matrix = matrixMultiply(L, H, matrix)
		Q = matrixMultiply(L, Q, H)
	}

	return Q, matrix
}

func isConverged(matrix *LTable) bool {
	n := matrix.Len()

	for i := 1; i <= n; i++ {
		for j := 1; j <= n; j++ {
			if i != j && math.Abs(float64(matrix.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber))) > 1e-6 {
				return false
			}
		}
	}

	return true
}

func calculateHouseholderMatrix(L *LState, vector *LTable) *LTable {
	n := vector.Len()
	magnitude := calculateVectorMagnitude(vector)

	v := L.NewTable()
	for i := 1; i <= n; i++ {
		value := vector.RawGetInt(i).(LNumber)
		if i == 1 {
			sign := LNumber(1)
			if value < 0 {
				sign = -1
			}
			value += sign * magnitude
		}
		v.RawSetInt(i, value)
	}

	vMagnitude := calculateVectorMagnitude(v)
	H := L.NewTable()
	for i := 1; i <= n; i++ {
		H.RawSetInt(i, L.NewTable())
		for j := 1; j <= n; j++ {
			if i == j {
				H.RawGetInt(i).(*LTable).RawSetInt(j, LNumber(1)-2*v.RawGetInt(i).(LNumber)*v.RawGetInt(j).(LNumber)/(vMagnitude*vMagnitude))
			} else {
				H.RawGetInt(i).(*LTable).RawSetInt(j, -2*v.RawGetInt(i).(LNumber)*v.RawGetInt(j).(LNumber)/(vMagnitude*vMagnitude))
			}
		}
	}

	return H
}

func calculateVectorMagnitude(vector *LTable) LNumber {
	magnitude := LNumber(0)
	for i := 1; i <= vector.Len(); i++ {
		value := vector.RawGetInt(i).(LNumber)
		magnitude += value * value
	}
	return LNumber(math.Sqrt(float64(magnitude)))
}

func matrixMultiply(L *LState, matrix1, matrix2 *LTable) *LTable {
	result := L.NewTable()

	for i := 1; i <= matrix1.Len(); i++ {
		row1 := matrix1.RawGetInt(i).(*LTable)

		resultRow := L.NewTable()
		for j := 1; j <= matrix2.RawGetInt(1).(*LTable).Len(); j++ {
			var sum LNumber
			for k := 1; k <= row1.Len(); k++ {
				col2 := matrix2.RawGetInt(k).(*LTable)
				sum += row1.RawGetInt(k).(LNumber) * col2.RawGetInt(j).(LNumber)
			}
			resultRow.RawSetInt(j, sum)
		}
		result.RawSetInt(i, resultRow)
	}

	return result
}

func copyMatrix(L *LState, matrix *LTable) *LTable {
	result := L.NewTable()
	for i := 1; i <= matrix.Len(); i++ {
		row := matrix.RawGetInt(i).(*LTable)
		resultRow := L.NewTable()
		for j := 1; j <= row.Len(); j++ {
			value := row.RawGetInt(j).(LNumber)
			resultRow.RawSetInt(j, value)
		}
		result.RawSetInt(i, resultRow)
	}
	return result
}

func matrixLU(L *LState) int {
	matrix := L.CheckTable(1)

	_, U := luDecomposition(L, matrix)

	result := L.NewTable()
	result.RawSetString("L", L)
	result.RawSetString("U", U)

	L.Push(result)
	return 1
}

func matrixLUP(L *LState) int {
	matrix := L.CheckTable(1)

	LU, P := lupDecomposition(L, matrix)

	result := L.NewTable()
	result.RawSetString("LU", LU)
	result.RawSetString("P", P)

	L.Push(result)
	return 1
}

func matrixQR(L *LState) int {
	matrix := L.CheckTable(1)

	Q, R := qrDecomposition(L, matrix)

	result := L.NewTable()
	result.RawSetString("Q", Q)
	result.RawSetString("R", R)

	L.Push(result)
	return 1
}

func matrixSolve(L *LState) int {
	A := L.CheckTable(1)
	b := L.CheckTable(2)

	// 使用LU分解求解线性方程组 Ax = b
	_, P := lupDecomposition(L, A)
	x := luSolve(L, P, b)

	L.Push(x)
	return 1
}

func matrixRREF(L *LState) int {
	matrix := L.CheckTable(1)

	rref := gaussJordanElimination(L, matrix)

	L.Push(rref)
	return 1
}

func gaussJordanElimination(L *LState, matrix *LTable) *LTable {
	n := matrix.Len()
	m := matrix.RawGetInt(1).(*LTable).Len()

	rref := copyMatrix(L, matrix)

	row := 1
	for col := 1; col <= m; col++ {
		// 选取主元
		pivotRow := row
		for i := row; i <= n; i++ {
			if rref.RawGetInt(i).(*LTable).RawGetInt(col).(LNumber) != 0 {
				pivotRow = i
				break
			}
		}

		if rref.RawGetInt(pivotRow).(*LTable).RawGetInt(col).(LNumber) == 0 {
			continue
		}

		// 交换行
		tempRow := rref.RawGetInt(row)
		rref.RawSetInt(row, rref.RawGetInt(pivotRow))
		rref.RawSetInt(pivotRow, tempRow)

		// 主元归一
		pivot := rref.RawGetInt(row).(*LTable).RawGetInt(col).(LNumber)
		for j := col; j <= m; j++ {
			rref.RawGetInt(row).(*LTable).RawSetInt(j, rref.RawGetInt(row).(*LTable).RawGetInt(j).(LNumber)/pivot)
		}

		// 消元
		for i := 1; i <= n; i++ {
			if i == row {
				continue
			}

			factor := rref.RawGetInt(i).(*LTable).RawGetInt(col).(LNumber)
			for j := col; j <= m; j++ {
				rref.RawGetInt(i).(*LTable).RawSetInt(j, rref.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber)-factor*rref.RawGetInt(row).(*LTable).RawGetInt(j).(LNumber))
			}
		}

		row++
	}

	return rref
}

func matrixNullspace(L *LState) int {
	matrix := L.CheckTable(1)

	// 计算RREF
	rref := gaussJordanElimination(L, matrix)

	// 找出自由变量
	freeVariables := findFreeVariables(rref)

	// 构造零空间的基向量
	nullspace := constructNullspaceBasis(L, rref, freeVariables)

	L.Push(nullspace)
	return 1
}

func findFreeVariables(matrix *LTable) []int {
	freeVariables := []int{}
	m := matrix.RawGetInt(1).(*LTable).Len()

	for j := 1; j <= m; j++ {
		isPivot := false
		for i := 1; i <= matrix.Len(); i++ {
			if matrix.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber) != 0 {
				isPivot = true
				break
			}
		}
		if !isPivot {
			freeVariables = append(freeVariables, j)
		}
	}

	return freeVariables
}

func constructNullspaceBasis(L *LState, rref *LTable, freeVariables []int) *LTable {
	m := rref.RawGetInt(1).(*LTable).Len()
	n := rref.Len()
	nullspace := L.NewTable()

	for _, j := range freeVariables {
		vector := L.NewTable()
		for i := 1; i <= m; i++ {
			if i == j {
				vector.RawSetInt(i, LNumber(1))
			} else if i <= n {
				vector.RawSetInt(i, -rref.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber))
			} else {
				vector.RawSetInt(i, LNumber(0))
			}
		}
		nullspace.RawSetInt(nullspace.Len()+1, vector)
	}

	return nullspace
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func matrixRowSpace(L *LState) int {
	matrix := L.CheckTable(1)

	// 计算RREF
	rref := gaussJordanElimination(L, matrix)

	// 提取非零行作为行空间的基
	rowSpace := extractNonZeroRows(L, rref)

	L.Push(rowSpace)
	return 1
}

func extractNonZeroRows(L *LState, matrix *LTable) *LTable {
	rowSpace := L.NewTable()
	for i := 1; i <= matrix.Len(); i++ {
		row := matrix.RawGetInt(i).(*LTable)
		nonZero := false
		for j := 1; j <= row.Len(); j++ {
			if row.RawGetInt(j).(LNumber) != 0 {
				nonZero = true
				break
			}
		}
		if nonZero {
			rowSpace.RawSetInt(rowSpace.Len()+1, row)
		}
	}
	return rowSpace
}

func matrixOrthogonalize(L *LState) int {
	matrix := L.CheckTable(1)

	// 使用Gram-Schmidt正交化过程
	orthogonalized := gramSchmidtProcess(L, matrix, false)

	L.Push(orthogonalized)
	return 1
}

func gramSchmidtProcess(L *LState, matrix *LTable, normalize bool) *LTable {
	n := matrix.Len()

	orthogonalized := L.NewTable()
	for i := 1; i <= n; i++ {
		orthogonalized.RawSetInt(i, L.NewTable())
	}

	for i := 1; i <= n; i++ {
		v := copyVector(L, matrix.RawGetInt(i).(*LTable))
		for j := 1; j < i; j++ {
			u := copyVector(L, orthogonalized.RawGetInt(j).(*LTable))
			projection := calculateProjection(L, v, u)
			v = subtract(L, v, projection)
		}
		if normalize {
			v = normalizeVector(L, v)
		}
		orthogonalized.RawSetInt(i, v)
	}

	return orthogonalized
}

func normalizeVector(L *LState, vector *LTable) *LTable {
	magnitude := calculateVectorMagnitude(vector)
	normalized := L.NewTable()
	for i := 1; i <= vector.Len(); i++ {
		normalized.RawSetInt(i, vector.RawGetInt(i).(LNumber)/magnitude)
	}
	return normalized
}

func copyVector(L *LState, vector *LTable) *LTable {
	v := L.NewTable()
	for i := 1; i <= vector.Len(); i++ {
		v.RawSetInt(i, vector.RawGetInt(i))
	}
	return v
}

func calculateProjection(L *LState, v, u *LTable) *LTable {
	projection := copyVector(L, u)
	dot := calculateDotProduct(v, u)
	for i := 1; i <= projection.Len(); i++ {
		projection.RawSetInt(i, projection.RawGetInt(i).(LNumber)*dot)
	}
	return projection
}

func calculateDotProduct(v, u *LTable) LNumber {
	dot := LNumber(0)
	for i := 1; i <= v.Len(); i++ {
		dot += v.RawGetInt(i).(LNumber) * u.RawGetInt(i).(LNumber)
	}
	return dot
}

func subtract(L *LState, v, u *LTable) *LTable {
	result := L.NewTable()
	for i := 1; i <= v.Len(); i++ {
		result.RawSetInt(i, v.RawGetInt(i).(LNumber)-u.RawGetInt(i).(LNumber))
	}
	return result
}

func matrixOrthonormalize(L *LState) int {
	matrix := L.CheckTable(1)

	// 使用Gram-Schmidt正交化过程，并进行归一化
	orthonormalized := gramSchmidtProcess(L, matrix, true)

	L.Push(orthonormalized)
	return 1
}

func matrixGramSchmidt(L *LState) int {
	return matrixOrthonormalize(L)
}

func matrixHouseholder(L *LState) int {
	vector := L.CheckTable(1)

	// 计算Householder变换矩阵
	H := calculateHouseholderMatrix(L, vector)

	L.Push(H)
	return 1
}

func matrixGivens(L *LState) int {
	a := L.CheckNumber(1)
	b := L.CheckNumber(2)

	// 计算Givens旋转矩阵
	G := calculateGivensMatrix(L, a, b)

	L.Push(G)
	return 1
}

func calculateGivensMatrix(L *LState, a, b LNumber) *LTable {
	G := NewMatrix(L, 2, 2)
	c := float64(a) / math.Sqrt(float64(a*a+b*b))
	s := float64(b) / math.Sqrt(float64(a*a+b*b))

	G.Set(1, 1, LNumber(c))
	G.Set(1, 2, LNumber(-s))
	G.Set(2, 1, LNumber(s))
	G.Set(2, 2, LNumber(c))

	return G
}

func (G *LTable) Set(i, j int, value LNumber) {
	G.RawGetInt(i).(*LTable).RawSetInt(j, value)
}

func NewMatrix(L *LState, rows, cols int) *LTable {
	matrix := L.NewTable()
	for i := 1; i <= rows; i++ {
		matrix.RawSetInt(i, L.NewTable())
		for j := 1; j <= cols; j++ {
			matrix.RawGetInt(i).(*LTable).RawSetInt(j, LNumber(0))
		}
	}
	return matrix
}

func matrixJacobi(L *LState) int {
	A := L.CheckTable(1)
	b := L.CheckTable(2)
	maxIterations := L.OptInt(3, 100)
	tolerance := L.OptNumber(4, 1e-6)

	// 使用Jacobi迭代法求解线性方程组
	x := jacobiIteration(L, A, b, maxIterations, LNumber(tolerance))

	L.Push(x)
	return 1
}

func jacobiIteration(L *LState, A, b *LTable, maxIterations int, tolerance LNumber) *LTable {
	n := A.Len()

	// 初始化解向量
	x := L.NewTable()
	for i := 1; i <= n; i++ {
		x.RawSetInt(i, LNumber(0))
	}

	// 迭代求解
	for k := 1; k <= maxIterations; k++ {
		xNew := L.NewTable()
		for i := 1; i <= n; i++ {
			sum := LNumber(0)
			for j := 1; j <= n; j++ {
				if i != j {
					sum += A.RawGetInt(i).(*LTable).RawGetInt(j).(LNumber) * x.RawGetInt(j).(LNumber)
				}
			}
			xNew.RawSetInt(i, (b.RawGetInt(i).(LNumber)-sum)/A.RawGetInt(i).(*LTable).RawGetInt(i).(LNumber))
		}

		// 检查收敛
		if isConvergedJacobi(x, xNew, tolerance) {
			break
		}

		x = xNew
	}

	return x
}

func isConvergedJacobi(x, xNew *LTable, tolerance LNumber) bool {
	n := x.Len()

	for i := 1; i <= n; i++ {
		if math.Abs(float64(x.RawGetInt(i).(LNumber)-xNew.RawGetInt(i).(LNumber))) > float64(tolerance) {
			return false
		}
	}

	return true
}

func matrixLanczos(L *LState) int {
	A := L.CheckTable(1)
	v := L.CheckTable(2)
	k := L.CheckInt(3)

	// 执行Lanczos算法
	T, Q := lanczosAlgorithm(L, A, v, k)

	result := L.NewTable()
	result.RawSetString("T", T)
	result.RawSetString("Q", Q)

	L.Push(result)
	return 1
}

func lanczosAlgorithm(L *LState, A, v *LTable, k int) (*LTable, *LTable) {
	n := A.Len()

	// 初始化矩阵T和Q
	T := NewMatrix(L, k, k)
	Q := NewMatrix(L, n, k)

	// 初始化向量q1
	q := copyVector(L, v)
	Q.SetColumn(1, q)

	// 初始化向量r0
	r := multiplyMatrixVector(L, A, q)
	alpha := calculateDotProduct(q, r)
	r = subtract(L, r, multiplyVectorScalar(L, q, alpha))
	beta := calculateVectorMagnitude(r)

	// 迭代计算T和Q
	for i := 1; i <= k; i++ {
		T.Set(i, i, LNumber(alpha))
		if i < k {
			T.Set(i, i+1, LNumber(beta))
			T.Set(i+1, i, LNumber(beta))
		}

		if i == k {
			break
		}

		q = normalizeVector(L, r)
		Q.SetColumn(i+1, q)

		r = multiplyMatrixVector(L, A, q)
		alpha = calculateDotProduct(q, r)
		r = subtract(L, r, multiplyVectorScalar(L, q, alpha))
		beta = calculateVectorMagnitude(r)
	}

	return T, Q
}

func (Q *LTable) SetColumn(j int, vector *LTable) {
	for i := 1; i <= Q.Len(); i++ {
		Q.RawGetInt(i).(*LTable).RawSetInt(j, vector.RawGetInt(i))
	}
}

func multiplyMatrixVector(L *LState, matrix, vector *LTable) *LTable {
	result := L.NewTable()
	for i := 1; i <= matrix.Len(); i++ {
		row := matrix.RawGetInt(i).(*LTable)
		sum := LNumber(0)
		for j := 1; j <= row.Len(); j++ {
			sum += row.RawGetInt(j).(LNumber) * vector.RawGetInt(j).(LNumber)
		}
		result.RawSetInt(i, sum)
	}
	return result
}

func multiplyVectorScalar(L *LState, vector *LTable, scalar LNumber) *LTable {
	result := L.NewTable()
	for i := 1; i <= vector.Len(); i++ {
		result.RawSetInt(i, vector.RawGetInt(i).(LNumber)*scalar)
	}
	return result
}

func matrixPowerMethod(L *LState) int {
	A := L.CheckTable(1)
	v := L.CheckTable(2)
	maxIterations := L.OptInt(3, 100)
	tolerance := L.OptNumber(4, 1e-6)

	// 使用幂法求解最大特征值
	lambda := powerMethod(L, A, v, maxIterations, LNumber(tolerance))

	L.Push(lambda)
	return 1
}

func powerMethod(L *LState, A, v *LTable, maxIterations int, tolerance LNumber) LNumber {
	// n := A.Len()

	// 初始化向量v
	x := copyVector(L, v)

	// 迭代求解
	for k := 1; k <= maxIterations; k++ {
		y := multiplyMatrixVector(L, A, x)
		lambda := calculateDotProduct(x, y)
		y = multiplyVectorScalar(L, y, 1/lambda)

		// 检查收敛
		if isConvergedPowerMethod(x, y, tolerance) {
			return lambda
		}

		x = y
	}

	L.RaiseError("Power method did not converge after %d iterations", maxIterations)
	return 0
}

func isConvergedPowerMethod(x, y *LTable, tolerance LNumber) bool {
	n := x.Len()

	for i := 1; i <= n; i++ {
		if math.Abs(float64(x.RawGetInt(i).(LNumber)-y.RawGetInt(i).(LNumber))) > float64(tolerance) {
			return false
		}
	}

	return true
}

func matrixPower(L *LState) int {
	matrix := L.CheckTable(1)
	power := L.CheckInt(2)

	// 计算矩阵的幂
	result := matrixExponentiation(L, matrix, power)

	L.Push(result)
	return 1
}

func matrixExponentiation(L *LState, matrix *LTable, power int) *LTable {
	result := copyMatrix(L, matrix)

	for i := 1; i < power; i++ {
		result = matrixMultiply(L, result, matrix)
	}

	return result
}
