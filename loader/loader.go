package loader

type StmtProcess func(string) error

type Loader interface {
	Load(interface{}, StmtProcess) error
}
