package loader

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"
)

// sql loader: load sql stmt from file
type FileSqlLoader struct {
	Path string
}

func (this *FileSqlLoader) Load(path interface{}, prc StmtProcess) error {
	p, ok := path.(string)
	if !ok {
		return errors.New("文件sql加载器输入路径为非字符串")
	}

	fd, err := os.Open(p)
	if err != nil {
		return errors.Wrap(err, "打开文件失败")
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	for {
		line, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		prc(line)
	}

	return nil
}
