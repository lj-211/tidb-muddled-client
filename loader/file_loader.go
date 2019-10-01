package loader

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/common"
)

// sql loader: load sql stmt from file
type FileSqlLoader struct {
	Path string
}

func (this *FileSqlLoader) Load(path interface{}, prc StmtProcess) error {
	if path == nil {
		return common.NilInputErr
	}
	p, ok := path.(string)
	if !ok {
		return common.ParamInvalidErr
	}

	fd, err := os.Open(p)
	if err != nil {
		return errors.Wrap(err, "打开文件失败")
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	// 这里输入保证是按行分割，所以不做更多容错处理
	for {
		line, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			prc("", true)
			break
		}

		prc(line, false)
	}

	return nil
}
