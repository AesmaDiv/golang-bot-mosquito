package db

type ConnectionParams struct {
	Host   string
	Port   int
	User   string
	Pswd   string
	Dbname string
}

type Helper interface {
	Connect(ConnectionParams) bool
	Disconnect()
	Query(sql string, columns []string) []map[string]any
	Select(tabla string, columns []string, where ...map[string]any) []map[string]any
	Insert(table string, rows []map[string]any, return_cols ...string) int
	Update(table string, set map[string]any, where ...map[string]any) int
	Delete(table string, where map[string]any) int
}
