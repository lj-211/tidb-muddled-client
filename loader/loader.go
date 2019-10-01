package loader

// sql处理函数(sql statement, is_last)
type StmtProcess func(string, bool) error

type Loader interface {
	Load(interface{}, StmtProcess) error
}
