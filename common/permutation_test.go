package common

import (
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// [][]int compare
func equal(cmp [][]int, target [][]int) bool {
	csize := len(cmp)
	tsize := len(target)

	if csize != tsize {
		return false
	}

	for i := 0; i < csize; i++ {
		ciSize := len(cmp[i])
		tiSize := len(target[i])
		if ciSize != tiSize {
			return false
		}

		if !reflect.DeepEqual(cmp[i], target[i]) {
			return false
		}
	}

	return true
}

func Test_FullPermutation(t *testing.T) {
	log.Println("test FullPermutation")

	// normal
	in := []int{1, 2, 3}
	ret := FullPermutation(in)
	cret := [][]int{
		[]int{1, 2, 3},
		[]int{1, 3, 2},
		[]int{2, 1, 3},
		[]int{2, 3, 1},
		[]int{3, 1, 2},
		[]int{3, 2, 1},
	}

	assert.Equal(t, 6, len(cret), "123的全排列数为6")
	assert.Equal(t, true, equal(ret, cret), "检查返回值必须相等")

	// zero slice
	in = []int{}
	ret = FullPermutation(in)
	cret = [][]int{}

	assert.Equal(t, true, equal(ret, cret), "返回值为空slice")
}
