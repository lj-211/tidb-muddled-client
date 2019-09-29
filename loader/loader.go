package loader

type StmtProcess func(string, bool) error

type Loader interface {
	Load(interface{}, StmtProcess) error
}
