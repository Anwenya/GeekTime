package week01

import "fmt"

// 要求一：能够实现删除操作就可以。
// 要求二：考虑使用比较高性能的实现。
// 要求三：改造为泛型方法
// 要求四：支持缩容，并旦设计缩容机制。

// DeleteByIndex:根据下标删除元素
func DeleteByIndex[T any](src []T, index int) ([]T, error) {
	length := len(src)
	if index < 0 || index >= length {
		return nil, fmt.Errorf("Illegal index, length:%d, index:%d", length, index)
	}

	for i := index; i+1 < length; i++ {
		src[i] = src[i+1]
	}

	return src[:length-1], nil
}

func DeleteByIndexOfCopy[T any](src []T, index int) ([]T, error) {
	length := len(src)
	if index < 0 || index >= length {
		return nil, fmt.Errorf("Illegal index, length:%d, index:%d", length, index)
	}
	copy(src[index:], src[index+1:])
	return src[:length-1], nil
}

// Shrink:缩容
func Shrink[T any](src []T) []T {
	capacity, length := cap(src), len(src)
	dstCapacity, changed := calCapacity(capacity, length)
	if !changed {
		return src
	}
	newer := make([]T, 0, dstCapacity)
	copy(newer, src[:length])
	return newer
}

func calCapacity(capacity, length int) (int, bool) {
	// fast path
	if capacity <= 64 {
		return capacity, false
	}

	// 数组长度大于2048 利用率不足0.5 缩容为0.625
	if capacity > 2048 && (capacity/length >= 2) {
		return int(0.625 * float32(capacity)), true
	}

	// 数组长度小于2048 利用率不足0.25 缩容为0.5
	if capacity <= 2048 && (capacity/length >= 4) {
		return capacity / 2, true
	}

	return capacity, false
}
