package main

import (
	"fmt"
	"slices"
)

type Pair struct {
	key   string
	value int
}

func main() {

	map1 := map[string]int{
		"first": 1,
		"sec":   3,
		"thres": 5,
	}
	map1["forth"] = 7
	map1["q"] = 8
	map1["six"] = 9

	map2 := map[string]int{
		"ten":    10,
		"eleven": 11,
		"twel":   12,
	}
	fmt.Println(mergeMap(map1, map2))
	
	map3 := mergeMap(map1, map2)
	fmt.Println(sortShit(map3))
}

func mergeMap(map1, map2 map[string]int) map[string]int {
	map3 := make(map[string]int)

	for key, value := range map1 {
		map3[key] = value
	}
	for key, value := range map2 {
		map3[key] = value
	}
	return map3
}

func sortShit(map3 map[string]int) []Pair {
	sortedList := make([]Pair, 0, len(map3))
	
	for key, value := range map3 {
		pair := Pair{key: key, value: value}
		sortedList = append(sortedList, pair)
	} 
	slices.SortFunc(sortedList, func(a, b Pair) int {
		return a.value - b.value
	})
	return sortedList

}
