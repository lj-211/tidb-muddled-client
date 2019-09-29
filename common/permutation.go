package common

func backtracking(nums []int, arr []int, depth int, mask map[int]bool, output *[][]int) {
	size := len(nums)
	if depth == size {
		newArr := make([]int, depth)
		copy(newArr, arr)
		*output = append(*output, newArr)
		return
	}

	for i := 0; i < size; i++ {
		v := nums[i]
		_, ok := mask[v]
		if !ok {
			arr[depth] = v
			mask[v] = true
			backtracking(nums, arr, depth+1, mask, output)
			delete(mask, v)
		}
	}
}

func FullPermutation(in []int) [][]int {
	size := len(in)
	if size == 0 {
		return [][]int{}
	}
	ret := make([][]int, 0)
	arr := make([]int, size)
	mask := make(map[int]bool)

	backtracking(in, arr, 0, mask, &ret)

	return ret
}
