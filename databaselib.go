package lua

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDatabase(L *LState) int {
	mod := L.RegisterModule(DatabaseLibName, databaseFuncs)
	L.Push(mod)
	return 1
}

var databaseFuncs = map[string]LGFunction{
	"open":         dbOpen,
	"close":        dbClose,
	"query":        dbQuery,
	"exec":         dbExec,
	"lastInsertId": dbLastInsertId,
	"rowsAffected": dbRowsAffected,
	"createTable":  dbCreateTable,
	"insert":       dbInsert,
	"update":       dbUpdate,
	"delete":       dbDelete,
}

func dbOpen(L *LState) int {
	dbname := L.CheckString(1)
	port := L.CheckInt(2)
	usr := L.CheckString(3)
	pwd := L.CheckString(4)
	db, err := openDatabaseConnection(dbname, port, usr, pwd)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Database connection error: %v", err)))
		return 2
	}
	ud := L.NewUserData()
	ud.Value = db
	L.SetMetatable(ud, L.GetTypeMetatable("database"))
	L.Push(ud)
	return 1
}

func openDatabaseConnection(dbname string, port int, usr, pwd string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", usr, pwd, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return db, nil
}

func dbClose(L *LState) int {
	ud := L.CheckUserData(1)
	db := ud.Value.(*sql.DB)
	if err := db.Close(); err != nil {
		L.Push(LBool(false))
		L.Push(LString(fmt.Sprintf("Failed to close database: %v", err)))
		return 2
	}
	L.Push(LBool(true))
	return 1
}

func dbQuery(L *LState) int {
	ud := L.CheckUserData(1)
	db := ud.Value.(*sql.DB)
	query := L.CheckString(2)

	rows, err := db.Query(query)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Query execution error: %v", err)))
		return 2
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Failed to get columns: %v", err)))
		return 2
	}

	results := L.NewTable()
	for rows.Next() {
		row := make([]interface{}, len(columns))
		rowPtrs := make([]interface{}, len(columns))
		for i := range row {
			rowPtrs[i] = &row[i]
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			L.Push(LNil)
			L.Push(LString(fmt.Sprintf("Failed to scan row: %v", err)))
			return 2
		}

		rowTable := L.NewTable()
		for i, col := range columns {
			value := row[i]
			if byteArray, ok := value.([]byte); ok {
				value = string(byteArray)
			}
			rowTable.RawSetString(col, LString(fmt.Sprintf("%v", value)))
		}
		results.Append(rowTable)
	}

	L.Push(results)
	return 1
}

func dbExec(L *LState) int {
	ud := L.CheckUserData(1)
	db := ud.Value.(*sql.DB)
	query := L.CheckString(2)

	res, err := db.Exec(query)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("SQL Error: %v", err)))
		return 2
	}

	_, err = res.LastInsertId()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Failed to get last insert ID: %v", err)))
		return 2
	}

	rowsAffected, _ := res.RowsAffected()
	L.Push(LBool(true))
	L.Push(LNumber(rowsAffected))
	return 2
}

func dbLastInsertId(L *LState) int {
	ud := L.CheckUserData(1)
	res, ok := ud.Value.(sql.Result)
	if !ok {
		L.Push(LNil)
		L.Push(LString("Invalid result type"))
		return 2
	}
	id, err := res.LastInsertId()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Failed to get last insert ID: %v", err)))
		return 2
	}
	L.Push(LNumber(id))
	return 1
}

func dbRowsAffected(L *LState) int {
	ud := L.CheckUserData(1)
	res, ok := ud.Value.(sql.Result)
	if !ok {
		L.Push(LNil)
		L.Push(LString("Invalid result type"))
		return 2
	}
	ra, err := res.RowsAffected()
	if err != nil {
		L.Push(LNil)
		L.Push(LString(fmt.Sprintf("Failed to get rows affected: %v", err)))
		return 2
	}
	L.Push(LNumber(ra))
	return 1
}

func checkDatabase(L *LState, n int) *sql.DB {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*sql.DB); ok {
		return v
	}
	L.ArgError(n, "database expected")
	return nil
}

func dbCreateTable(L *LState) int {
	db := checkDatabase(L, 1)
	tableName := L.CheckString(2)
	columns := L.CheckTable(3)

	var columnDefs []string
	columns.ForEach(func(k, v LValue) {
		if ks, ok := k.(LString); ok {
			if vs, ok := v.(LString); ok {
				columnDefs = append(columnDefs, fmt.Sprintf("%s %s", ks, vs))
			}
		}
	})

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columnDefs, ", "))
	_, err := db.Exec(query)
	if err != nil {
		L.Push(LBool(false))
		L.Push(LString(fmt.Sprintf("Failed to create table: %v", err)))
		return 2
	}

	L.Push(LBool(true))
	return 1
}

func dbInsert(L *LState) int {
	db := checkDatabase(L, 1)
	tableName := L.CheckString(2)
	data := L.CheckTable(3)

	var columns []string
	var placeholders []string
	var values []interface{}

	data.ForEach(func(k, v LValue) {
		if ks, ok := k.(LString); ok {
			columns = append(columns, string(ks))
			placeholders = append(placeholders, "?")
			values = append(values, luaValueToInterface(v))
		}
	})

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	result, err := db.Exec(query, values...)
	if err != nil {
		L.Push(LBool(false))
		L.Push(LString(fmt.Sprintf("Failed to insert data: %v", err)))
		return 2
	}

	lastInsertId, _ := result.LastInsertId()
	L.Push(LBool(true))
	L.Push(LNumber(lastInsertId))
	return 2
}

func dbUpdate(L *LState) int {
	db := checkDatabase(L, 1)
	tableName := L.CheckString(2)
	data := L.CheckTable(3)
	condition := L.CheckString(4)

	var setStatements []string
	var values []interface{}

	data.ForEach(func(k, v LValue) {
		if ks, ok := k.(LString); ok {
			setStatements = append(setStatements, fmt.Sprintf("%s = ?", ks))
			values = append(values, luaValueToInterface(v))
		}
	})

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setStatements, ", "), condition)
	result, err := db.Exec(query, values...)
	if err != nil {
		L.Push(LBool(false))
		L.Push(LString(fmt.Sprintf("Failed to update data: %v", err)))
		return 2
	}

	rowsAffected, _ := result.RowsAffected()
	L.Push(LBool(true))
	L.Push(LNumber(rowsAffected))
	return 2
}

func dbDelete(L *LState) int {
	db := checkDatabase(L, 1)
	tableName := L.CheckString(2)
	condition := L.CheckString(3)

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, condition)
	result, err := db.Exec(query)
	if err != nil {
		L.Push(LBool(false))
		L.Push(LString(fmt.Sprintf("Failed to delete data: %v", err)))
		return 2
	}

	rowsAffected, _ := result.RowsAffected()
	L.Push(LBool(true))
	L.Push(LNumber(rowsAffected))
	return 2
}

func luaValueToInterface(v LValue) interface{} {
	switch v.Type() {
	case LTNil:
		return nil
	case LTBool:
		return bool(v.(LBool))
	case LTNumber:
		return float64(v.(LNumber))
	case LTString:
		return string(v.(LString))
	default:
		return v.String()
	}
}
