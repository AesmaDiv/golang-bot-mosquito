package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"golang-bot/sugar"

	_ "github.com/lib/pq"
)

const (
	NAME = "HelperPostgres"
)

var STATE = map[bool]string{
	true:  "Succes",
	false: "Error",
}

type HelperPostgres struct {
	Helper
	is_connected bool
	db           *sql.DB
}

func (h *HelperPostgres) Connect(params ConnectionParams) bool {
	conn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		params.Host, params.Port, params.User, params.Pswd, params.Dbname)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Connect", err.Error())
		return false
	}
	h.db = db

	err = h.db.Ping()
	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Connect", err.Error())
		return false
	}
	h.is_connected = true

	return h.is_connected
}

func (h *HelperPostgres) Disconnect() {
	if !h.checkConnection() {
		return
	}

	h.db.Close()
	h.is_connected = false
}

func (h *HelperPostgres) Query(sql string, columns []string) []map[string]any {
	if !h.checkConnection() {
		return nil
	}

	rows, err := h.db.Query(sql)
	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Query", err.Error())
		return nil
	}
	defer rows.Close()

	var vals = make([]any, len(columns))
	var ptrs []any
	for i := 0; i < len(vals); i++ {
		ptrs = append(ptrs, &vals[i])
	}

	var result []map[string]any
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			sugar.Log("ERROR", "HelperPostgres::Query", err.Error())
			return nil
		}
		dict := make(map[string]any)
		for i, v := range columns {
			dict[v] = vals[i]
		}
		result = append(result, dict)
	}
	return result
}

func (h *HelperPostgres) Select(table string, columns []string, where ...map[string]any) []map[string]any {
	if !h.checkConnection() {
		return nil
	}

	_columns := strings.Join(columns, ",")
	query := fmt.Sprintf("SELECT %s FROM %s", _columns, table)
	query += h.checkWhere(where)
	rows, err := h.db.Query(query)
	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Select", err.Error())
		return nil
	}
	defer rows.Close()

	if _columns == "*" {
		columns, _ = rows.Columns()
	}
	var vals = make([]any, len(columns))
	var ptrs []any
	for i := 0; i < len(vals); i++ {
		ptrs = append(ptrs, &vals[i])
	}

	var result []map[string]any
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			sugar.Log("ERROR", "HelperPostgres::Select", err.Error())
			return nil
		}

		dict := make(map[string]any)
		for i, v := range columns {
			dict[v] = vals[i]
		}
		result = append(result, dict)
	}
	return result
}

func (h *HelperPostgres) Insert(table string, rows []map[string]any, ret_cols ...string) int {
	// проверка подключения и длины массива параметров
	if !h.checkConnection() || len(rows) == 0 {
		return 0
	}
	// получение списка столбцов и первого радя параметров
	var cols, vals string = h.toPairOfStrings(rows[0], ",")
	// сброка всех рядов параметров в одну строку
	var arr_vals []string
	arr_vals = append(arr_vals, vals)
	for _, row := range rows[1:] {
		_, vals = h.toPairOfStrings(row, ",")
		arr_vals = append(arr_vals, vals)
	}
	vals = strings.Join(arr_vals, "),(")
	// вставка в БД
	var err error
	var result int = 1
	var returning string
	if len(ret_cols) > 0 {
		returning = fmt.Sprintf(" RETURNING %s", strings.Join(ret_cols, ","))
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)%s;", table, cols, vals, returning)
	if len(ret_cols) > 0 {
		err = h.db.QueryRow(query).Scan(&result)
	} else {
		_, err = h.db.Exec(query)
	}
	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Insert", err.Error())
		result = 0
	}

	return result
}

func (h *HelperPostgres) Update(table string, set map[string]any, where ...map[string]any) int {
	if !h.checkConnection() {
		return 0
	}

	var result int = 1
	_set := h.toStringOfPairs(set, ",")
	query := fmt.Sprintf("UPDATE %s SET %s", table, _set)
	query += h.checkWhere(where)
	_, err := h.db.Exec(query)

	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Update", err.Error())
		result = 0
	}

	return result
}

func (h *HelperPostgres) Delete(table string, where map[string]any) int {
	if !h.checkConnection() {
		return 0
	}

	var result int = -1
	_where := h.toStringOfPairs(where, " AND ")

	query := fmt.Sprintf("DELETE FROM %s WHERE %s;", table, _where)
	_, err := h.db.Exec(query)

	if err != nil {
		sugar.Log("ERROR", "HelperPostgres::Delete", err.Error())
	}

	return result
}

func (h *HelperPostgres) checkConnection() bool {
	var result = h.is_connected && h.db != nil
	if !result {
		sugar.Log("ERROR", "HelperPostgres::checkConnection", "DB is not connected.")
	}
	return result
}

func (h *HelperPostgres) checkWhere(where []map[string]any) string {
	if len(where) > 0 {
		_where := h.toStringOfPairs(where[0], " AND ")
		if _where != "" {
			return fmt.Sprintf(" WHERE %s;", _where)
		}
	}
	return ";"
}

func (h *HelperPostgres) toStringOfPairs(set map[string]any, sep string) string {
	if (set == nil) || len(set) == 0 {
		return ""
	}
	var key, val string
	var pairs []string
	for k, v := range set {
		key = k
		val = fmt.Sprintf(
			sugar.Iif(
				reflect.TypeOf(v).String() == "string",
				sugar.Iif(v == "", "NULL%s", "'%s'"),
				"%v"),
			v)
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, val))
	}

	return strings.Join(pairs, sep)
}

func (h *HelperPostgres) toPairOfStrings(set map[string]any, sep string) (string, string) {
	if (set == nil) || len(set) == 0 {
		return "", ""
	}
	var keys, vals []string
	for k, v := range set {
		keys = append(keys, k)
		vals = append(
			vals,
			fmt.Sprintf(
				sugar.Iif(
					reflect.TypeOf(v).String() == "string",
					sugar.Iif(v == "", "NULL%s", "'%s'"),
					"%v"),
				v),
		)
	}

	return strings.Join(keys, sep), strings.Join(vals, sep)
}
