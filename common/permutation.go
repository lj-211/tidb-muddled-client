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

func backtrackingNew(numList [][]int, mask []int, depth int, arr []int, output *[][]int) {
	if output == nil {
		return
	}

	if len(numList) != len(mask) {
		return
	}

	if depth == len(arr) {
		newArr := make([]int, len(arr))
		copy(newArr, arr)
		*output = append(*output, newArr)
		return
	}

	for i := 0; i < len(numList); i++ {
		v := numList[i]
		cursor := mask[i]
		if cursor < len(v) {
			mask[i]++
			arr[depth] = v[cursor]
			backtrackingNew(numList, mask, depth+1, arr, output)
			mask[i]--
		}
	}
}

func FullPermutationNew(numList [][]int) [][]int {
	size := len(numList)
	if size == 0 {
		return [][]int{}
	}

	allSize := 0
	for i := 0; i < size; i++ {
		v := numList[i]
		allSize += len(v)
	}

	cursorList := make([]int, size)
	output := make([][]int, 0)
	arr := make([]int, allSize)

	backtrackingNew(numList, cursorList, 0, arr, &output)

	return output
}
