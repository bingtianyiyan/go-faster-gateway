package utils

// 过滤函数示例
func Filter[T any](slice []T, f func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

func UnionSlicesUnique(slice1, slice2 []string) []string {
	set := make(map[string]bool)
	var result []string

	for _, item := range append(slice1, slice2...) {
		if !set[item] {
			set[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Set 泛型集合类型
type Set[T comparable] map[T]struct{}

// NewSet 创建新集合
func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T])
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// Union 返回两个集合的并集
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := make(Set[T])
	for k := range s {
		result[k] = struct{}{}
	}
	for k := range other {
		result[k] = struct{}{}
	}
	return result
}
