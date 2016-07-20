package util

func QuickSort(array []int, start, end int) {
	if start >= end {
		return
	}
	//以最后一个数据为基准点
	s := end
	//i为从头开始遍历时小于基准的游标
	i := start - 1
	//j为从头开始遍历时大于基准的游标
	j := start
	for j <= s {
		//添加=的意思是将最后的基准点调整到小于
		//和大于它的中间
		if array[j] <= array[s] {
			array[i+1], array[j] = array[j], array[i+1]
			i++
		}
		j++
	}
	QuickSort(array, start, i-1)
	QuickSort(array, i+1, end)
}

func QuickSortS(array []string, start, end int) {
	if start >= end {
		return
	}
	//以最后一个数据为基准点
	s := end
	//i为从头开始遍历时小于基准的游标
	i := start - 1
	//j为从头开始遍历时大于基准的游标
	j := start
	for j <= s {
		//添加=的意思是将最后的基准点调整到小于
		//和大于它的中间
		if array[j] <= array[s] {
			array[i+1], array[j] = array[j], array[i+1]
			i++
		}
		j++
	}
	QuickSortS(array, start, i-1)
	QuickSortS(array, i+1, end)
}
