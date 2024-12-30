package lua

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

var goCodeBuffer string // 用于存储传入的Go代码

func OpenFFI(L *LState) int {
	mod := L.RegisterModule(FFILibName, ffilibFuncs)
	L.Push(mod)
	return 1
}

var ffilibFuncs = map[string]LGFunction{
	"godef":  ffilibGoDef,
	"goexec": ffilibExec,
}

// 接收一段字符串形式的Go代码并存储到缓冲区
func ffilibGoDef(L *LState) int {
	goCode := L.CheckString(1)
	goCodeBuffer = goCode
	return 0
}

// 执行缓冲区中的Go代码
func ffilibExec(L *LState) int {
	if goCodeBuffer == "" {
		L.Push(LString("No Go code defined. Use godef() first."))
		return 1
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCodeBuffer, 0)
	if err != nil {
		L.Push(LString(fmt.Sprintf("Parse error: %v", err)))
		return 1
	}

	var mainFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			mainFunc = fn
			break
		}
	}

	if mainFunc == nil {
		L.Push(LString("No main function found in Go code"))
		return 1
	}

	output := &bytes.Buffer{}

	env := make(map[string]interface{}) // 模拟一个简单的变量环境

	for _, stmt := range mainFunc.Body.List {
		err := executeStmt(stmt, env, output)
		if err != nil {
			L.Push(LString(fmt.Sprintf("Execution error: %v", err)))
			return 1
		}
	}

	L.Push(LString(output.String()))
	return 1
}

// 执行单个语句
func executeStmt(stmt ast.Stmt, env map[string]interface{}, output *bytes.Buffer) error {
	switch s := stmt.(type) {

	case *ast.DeclStmt:
		genDecl, ok := s.Decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			return fmt.Errorf("unsupported declaration: %#v", s)
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range valueSpec.Names {
				var value interface{}
				if len(valueSpec.Values) > i { // 如果有初始值
					val, err := evaluateExpr(valueSpec.Values[i], env)
					if err != nil {
						return err
					}
					value = val
				}
				env[name.Name] = value // 存储变量到环境中
			}
		}

	case *ast.ExprStmt: // 表达式语句，如函数调用：fmt.Println(...)
		err := executeExpr(s.X, env, output)
		if err != nil {
			return err
		}

	case *ast.AssignStmt:
		for i, lhs := range s.Lhs {
			nameIdent, ok := lhs.(*ast.Ident)
			if !ok {
				return fmt.Errorf("unsupported assignment target: %#v", lhs)
			}

			var rhsValue interface{}
			var err error

			if s.Tok == token.ADD_ASSIGN {
				leftValue, exists := env[nameIdent.Name]
				if !exists {
					return fmt.Errorf("undefined variable: %s", nameIdent.Name)
				}
				rightValue, err := evaluateExpr(s.Rhs[i], env)
				if err != nil {
					return err
				}
				rhsValue = leftValue.(int) + rightValue.(int)
			} else {
				rhsValue, err = evaluateExpr(s.Rhs[i], env)
				if err != nil {
					return err
				}
			}

			env[nameIdent.Name] = rhsValue
		}
	case *ast.ReturnStmt:
		for _, result := range s.Results {
			val, err := evaluateExpr(result, env)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, val)
		}

	case *ast.BlockStmt:
		for _, stmt := range s.List {
			err := executeStmt(stmt, env, output)
			if err != nil {
				return err
			}
		}

	case *ast.IfStmt:
		cond, err := evaluateExpr(s.Cond, env)
		if err != nil {
			return err
		}
		if cond.(bool) {
			err := executeStmt(s.Body, env, output)
			if err != nil {
				return err
			}
		} else if s.Else != nil {
			err := executeStmt(s.Else, env, output)
			if err != nil {
				return err
			}
		}

	case *ast.ForStmt:
		if s.Init != nil {
			err := executeStmt(s.Init, env, output)
			if err != nil {
				return err
			}
		}
		for {
			cond, err := evaluateExpr(s.Cond, env)
			if err != nil {
				return err
			}
			if !cond.(bool) {
				break
			}
			err = executeStmt(s.Body, env, output)
			if err != nil {
				return err
			}
			if s.Post != nil {
				err = executeStmt(s.Post, env, output)
				if err != nil {
					return err
				}
			}
		}

	case *ast.SwitchStmt:
		cond, err := evaluateExpr(s.Tag, env)
		if err != nil {
			return err
		}
		for _, stmt := range s.Body.List {
			cas, ok := stmt.(*ast.CaseClause)
			if !ok {
				return fmt.Errorf("unsupported statement in switch body: %#v", stmt)
			}
			for _, expr := range cas.List {
				val, err := evaluateExpr(expr, env)
				if err != nil {
					return err
				}
				if val == cond {
					for _, stmt := range cas.Body {
						err := executeStmt(stmt, env, output)
						if err != nil {
							return err
						}
					}
					break
				}
			}
		}
	case *ast.TypeSwitchStmt:
		assign, ok := s.Assign.(*ast.AssignStmt)
		if !ok {
			return fmt.Errorf("unsupported type switch assign: %#v", s.Assign)
		}
		if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return fmt.Errorf("unsupported assignment in type switch: %#v", assign)
		}
		nameIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok {
			return fmt.Errorf("unsupported assignment target in type switch: %#v", assign.Lhs[0])
		}
		value, err := evaluateExpr(assign.Rhs[0], env)
		if err != nil {
			return fmt.Errorf("error evaluating type switch: %v", err)
		}
		switch value.(type) {
		case int:
			env[nameIdent.Name] = value.(int)
		case string:
			env[nameIdent.Name] = value.(string)
		case bool:
			env[nameIdent.Name] = value.(bool)
		case []int:
			env[nameIdent.Name] = value.([]int)
		case []string:
			env[nameIdent.Name] = value.([]string)
		case []bool:
			env[nameIdent.Name] = value.([]bool)
		case []interface{}:
			env[nameIdent.Name] = value.([]interface{})
		default:
			return fmt.Errorf("unsupported type in type switch: %T", value)
		}
		for _, stmt := range s.Body.List {
			err := executeStmt(stmt, env, output)
			if err != nil {
				return fmt.Errorf("error executing type switch: %v", err)
			}
		}
	case *ast.CaseClause:
		if len(s.List) == 0 {
			return fmt.Errorf("nil case clause: %#v", s)
		}
		for _, expr := range s.List {
			val, err := evaluateExpr(expr, env)
			if err != nil {
				return fmt.Errorf("error evaluating case clause: %v", err)
			}
			if val.(bool) {
				for _, stmt := range s.Body {
					err := executeStmt(stmt, env, output)
					if err != nil {
						return fmt.Errorf("error executing case clause: %v", err)
					}
				}
				break
			}
		}
	case *ast.GoStmt:
		if s.Call == nil {
			return fmt.Errorf("nil call in go statement: %#v", s)
		}
		err := executeExpr(s.Call.Fun, env, output)
		if err != nil {
			return fmt.Errorf("error executing go statement: %v", err)
		}
	case *ast.DeferStmt:
		if s.Call == nil {
			return fmt.Errorf("nil call in defer statement: %#v", s)
		}
		err := executeExpr(s.Call.Fun, env, output)
		if err != nil {
			return fmt.Errorf("error executing defer statement: %v", err)
		}
		defer func() {
			err := executeExpr(s.Call.Fun, env, output)
			if err != nil {
				fmt.Fprintf(output, "defer error: %v\n", err)
			}
		}()
	case *ast.RangeStmt:
		value, err := evaluateExpr(s.X, env)
		if err != nil {
			return fmt.Errorf("error evaluating range statement: %v", err)
		}
		switch v := value.(type) {
		case []int:
			for i, val := range v {
				env[s.Key.(*ast.Ident).Name] = i
				env[s.Value.(*ast.Ident).Name] = val
				err := executeStmt(s.Body, env, output)
				if err != nil {
					return fmt.Errorf("error executing range statement: %v", err)
				}
			}
		case []string:
			for i, val := range v {
				env[s.Key.(*ast.Ident).Name] = i
				env[s.Value.(*ast.Ident).Name] = val
				err := executeStmt(s.Body, env, output)
				if err != nil {
					return fmt.Errorf("error executing range statement: %v", err)
				}
			}
		case []bool:
			for i, val := range v {
				env[s.Key.(*ast.Ident).Name] = i
				env[s.Value.(*ast.Ident).Name] = val
				err := executeStmt(s.Body, env, output)
				if err != nil {
					return fmt.Errorf("error executing range statement: %v", err)
				}
			}
		case []interface{}:
			for i, val := range v {
				env[s.Key.(*ast.Ident).Name] = i
				env[s.Value.(*ast.Ident).Name] = val
				err := executeStmt(s.Body, env, output)
				if err != nil {
					return fmt.Errorf("error executing range statement: %v", err)
				}
			}
		default:
			return fmt.Errorf("unsupported range type: %T", value)
		}
	case *ast.LabeledStmt:
		return executeStmt(s.Stmt, env, output)
	case *ast.SendStmt:
		if s.Chan == nil {
			return fmt.Errorf("nil channel in send statement: %#v", s)
		}
		if s.Value == nil {
			return fmt.Errorf("nil value in send statement: %#v", s)
		}
		err := executeExpr(s.Chan, env, output)
		if err != nil {
			return fmt.Errorf("error executing channel in send statement: %v", err)
		}
		err = executeExpr(s.Value, env, output)
		if err != nil {
			return fmt.Errorf("error executing value in send statement: %v", err)
		}
	case *ast.BranchStmt:
		switch s.Tok {
		case token.BREAK:
			return fmt.Errorf("break statement not supported")
		case token.CONTINUE:
			return fmt.Errorf("continue statement not supported")
		}
		return fmt.Errorf("unsupported branch statement: %#v", s)

	case *ast.IncDecStmt:
		nameIdent, ok := s.X.(*ast.Ident)
		if !ok {
			return fmt.Errorf("unsupported increment/decrement target: %#v", s.X)
		}
		value, exists := env[nameIdent.Name]
		if !exists {
			return fmt.Errorf("undefined variable: %s", nameIdent.Name)
		}
		switch s.Tok {
		case token.INC:
			env[nameIdent.Name] = value.(int) + 1
		case token.DEC:
			env[nameIdent.Name] = value.(int) - 1
		}

	case *ast.SelectStmt:
		if len(s.Body.List) != 1 {
			return fmt.Errorf("unsupported select statement: %#v", s)
		}
		comm, ok := s.Body.List[0].(*ast.CommClause)
		if !ok {
			return fmt.Errorf("unsupported statement in select body: %#v", s.Body.List[0])
		}
		if comm.Comm == nil {
			return fmt.Errorf("nil comm in select statement: %#v", s)
		}
		assign, ok := comm.Comm.(*ast.AssignStmt)
		if !ok {
			return fmt.Errorf("unsupported comm in select statement: %#v", comm.Comm)
		}
		if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return fmt.Errorf("unsupported assignment in select statement: %#v", assign)
		}
		nameIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok {
			return fmt.Errorf("unsupported assignment target in select statement: %#v", assign.Lhs[0])
		}
		value, err := evaluateExpr(assign.Rhs[0], env)
		if err != nil {
			return fmt.Errorf("error evaluating select statement: %v", err)
		}
		env[nameIdent.Name] = value
		for _, stmt := range comm.Body {
			err := executeStmt(stmt, env, output)
			if err != nil {
				return fmt.Errorf("error executing select statement: %v", err)
			}
		}
	default:
		return fmt.Errorf("unsupported statement type: %#v", stmt)
	}

	return nil
}

// 执行表达式（主要用于函数调用）
func executeExpr(expr ast.Expr, env map[string]interface{}, output *bytes.Buffer) error {
	switch e := expr.(type) {

	case *ast.CallExpr: // 函数调用，如 fmt.Println(...)
		funcIdent, ok := e.Fun.(*ast.Ident)
		if !ok || funcIdent.Name != "println" { // 简单支持 println 函数
			return fmt.Errorf("unsupported function call: %#v", e.Fun)
		}

		switch funcIdent.Name {
		case "println":
			args := []interface{}{}
			for _, arg := range e.Args {
				val, err := evaluateExpr(arg, env)
				if err != nil {
					return err
				}
				args = append(args, val)
			}
			fmt.Fprintln(output, args...)

		case "printf":
			args := []interface{}{}
			for _, arg := range e.Args {
				val, err := evaluateExpr(arg, env)
				if err != nil {
					return err
				}
				args = append(args, val)
			}
			fmt.Fprintf(output, args[0].(string), args[1:]...)

		case "print":
			args := []interface{}{}
			for _, arg := range e.Args {
				val, err := evaluateExpr(arg, env)
				if err != nil {
					return err
				}
				args = append(args, val)
			}
			fmt.Fprint(output, args...)
		case "panic":
			if len(e.Args) != 1 {
				return fmt.Errorf("panic function must have exactly one argument")
			}
			val, err := evaluateExpr(e.Args[0], env)
			if err != nil {
				return err
			}
			panic(val)
		case "len":
			if len(e.Args) != 1 {
				return fmt.Errorf("len function must have exactly one argument")
			}
			val, err := evaluateExpr(e.Args[0], env)
			if err != nil {
				return err
			}
			switch v := val.(type) {
			case []int:
				fmt.Fprintln(output, len(v))
			case []string:
				fmt.Fprintln(output, len(v))
			case []bool:
				fmt.Fprintln(output, len(v))
			case []interface{}:
				fmt.Fprintln(output, len(v))
			default:
				return fmt.Errorf("unsupported type in len function: %T", val)
			}

		default:
			return fmt.Errorf("unsupported function call: %#v", e.Fun)
		}

	default:
		return fmt.Errorf("unsupported expression type: %#v", expr)
	}

	return nil
}

// 计算表达式的值（支持常量和标识符）
func evaluateExpr(expr ast.Expr, env map[string]interface{}) (interface{}, error) {
	switch e := expr.(type) {

	case *ast.BasicLit: // 字面量，如数字或字符串常量
		switch e.Kind {
		case token.INT:
			return strconv.Atoi(e.Value)
		case token.STRING:
			return e.Value[1 : len(e.Value)-1], nil // 去掉引号
		default:
			return nil, fmt.Errorf("unsupported literal type: %v", e.Kind)
		}

	case *ast.Ident: // 标识符，如变量名 x
		value, exists := env[e.Name]
		if !exists {
			return nil, fmt.Errorf("undefined variable: %s", e.Name)
		}
		return value, nil
	case *ast.BinaryExpr:
		left, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		right, err := evaluateExpr(e.Y, env)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case token.ADD:
			return left.(int) + right.(int), nil
		case token.SUB:
			return left.(int) - right.(int), nil
		case token.MUL:
			return left.(int) * right.(int), nil
		case token.QUO:
			return left.(int) / right.(int), nil
		case token.REM:
			return left.(int) % right.(int), nil
		case token.LSS:
			return left.(int) < right.(int), nil
		case token.LEQ:
			return left.(int) <= right.(int), nil
		case token.GTR:
			return left.(int) > right.(int), nil
		case token.GEQ:
			return left.(int) >= right.(int), nil
		case token.EQL:
			return left.(int) == right.(int), nil
		case token.NEQ:
			return left.(int) != right.(int), nil
		case token.LAND:
			return left.(bool) && right.(bool), nil
		case token.LOR:
			return left.(bool) || right.(bool), nil
		case token.SHL:
			return left.(int) << right.(uint), nil
		case token.SHR:
			return left.(int) >> right.(uint), nil
		case token.AND:
			return left.(int) & right.(int), nil
		case token.OR:
			return left.(int) | right.(int), nil
		case token.XOR:
			return left.(int) ^ right.(int), nil
		case token.AND_NOT:
			return left.(int) &^ right.(int), nil
		case token.ADD_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for +=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) + rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.SUB_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for -=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) - rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.MUL_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for *=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) * rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.QUO_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for /=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) / rightValue.(int)

			env[leftIdent.Name] = result
			return result, nil
		case token.REM_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("%s", "unsupported left-hand side for %=")
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) % rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.AND_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for &=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) & rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.OR_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for |=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) | rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.XOR_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for ^=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) ^ rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.SHL_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for <<=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) << rightValue.(uint)
			env[leftIdent.Name] = result
			return result, nil
		case token.SHR_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for >>=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) >> rightValue.(uint)
			env[leftIdent.Name] = result
			return result, nil
		case token.AND_NOT_ASSIGN:
			leftIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("unsupported left-hand side for &^=: %#v", e.X)
			}
			leftValue, exists := env[leftIdent.Name]
			if !exists {
				return nil, fmt.Errorf("undefined variable: %s", leftIdent.Name)
			}
			rightValue, err := evaluateExpr(e.Y, env)
			if err != nil {
				return nil, err
			}
			result := leftValue.(int) &^ rightValue.(int)
			env[leftIdent.Name] = result
			return result, nil
		case token.ASSIGN:
			env[e.X.(*ast.Ident).Name] = right
			return right, nil

		default:
			return nil, fmt.Errorf("unsupported binary operator: %v", e.Op)
		}
	case *ast.UnaryExpr:
		x, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case token.ADD:
			return x, nil
		case token.SUB:
			return -x.(int), nil
		case token.NOT:
			return !x.(bool), nil
		case token.XOR:
			return ^x.(int), nil
		case token.AND:
			return &x, nil
		default:
			return nil, fmt.Errorf("unsupported unary operator: %v", e.Op)
		}
	case *ast.ParenExpr:
		return evaluateExpr(e.X, env)
	case *ast.IndexExpr:
		x, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		i, err := evaluateExpr(e.Index, env)
		if err != nil {
			return nil, err
		}
		return x.([]int)[i.(int)], nil
	case *ast.SliceExpr:
		x, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		low, err := evaluateExpr(e.Low, env)
		if err != nil {
			return nil, fmt.Errorf("error evaluating slice low: %v", err)
		}
		high, err := evaluateExpr(e.High, env)
		if err != nil {
			return nil, fmt.Errorf("error evaluating slice high: %v", err)
		}
		return x.([]int)[low.(int):high.(int)], nil
	case *ast.SelectorExpr:
		x, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		return x.(map[string]int)[e.Sel.Name], nil
	case *ast.CompositeLit:
		switch e.Type.(*ast.Ident).Name {
		case "int":
			vals := []int{}
			for _, elt := range e.Elts {
				val, err := evaluateExpr(elt, env)
				if err != nil {
					return nil, err
				}
				vals = append(vals, val.(int))
			}
			return vals, nil
		case "string":
			vals := []string{}
			for _, elt := range e.Elts {
				val, err := evaluateExpr(elt, env)
				if err != nil {
					return nil, err
				}
				vals = append(vals, val.(string))
			}
			return vals, nil
		case "bool":
			vals := []bool{}
			for _, elt := range e.Elts {
				val, err := evaluateExpr(elt, env)
				if err != nil {
					return nil, err
				}
				vals = append(vals, val.(bool))
			}
			return vals, nil
		}
		return nil, fmt.Errorf("unsupported composite literal type: %s", e.Type.(*ast.Ident).Name)
	case *ast.KeyValueExpr:
		key, err := evaluateExpr(e.Key, env)
		if err != nil {
			return nil, err
		}
		val, err := evaluateExpr(e.Value, env)
		if err != nil {
			return nil, err
		}
		return map[string]int{key.(string): val.(int)}, nil
	case *ast.TypeAssertExpr:
		x, err := evaluateExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		switch x.(type) {
		case int:
			return x.(int), nil
		case string:
			return x.(string), nil
		case bool:
			return x.(bool), nil
		case []int:
			return x.([]int), nil
		case []string:
			return x.([]string), nil
		case []bool:
			return x.([]bool), nil
		case []interface{}:
			return x.([]interface{}), nil
		}
		return nil, fmt.Errorf("unsupported type assertion: %T", x)
	case *ast.StructType:
		if len(e.Fields.List) != 1 {
			return nil, fmt.Errorf("unsupported struct type: %#v", e)
		}
		field := e.Fields.List[0]
		if len(field.Names) != 1 {
			return nil, fmt.Errorf("unsupported struct field: %#v", field)
		}
		return map[string]int{field.Names[0].Name: 0}, nil
	case *ast.CallExpr:
		fun, err := evaluateExpr(e.Fun, env)
		if err != nil {
			return nil, err
		}
		args := []interface{}{}
		for _, arg := range e.Args {
			val, err := evaluateExpr(arg, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		switch f := fun.(type) {
		case func(...interface{}) interface{}:
			return f(args...), nil
		}
		return nil, fmt.Errorf("unsupported function type: %T", fun)

	default:
		return nil, fmt.Errorf("Unsupported ")
	}
}
