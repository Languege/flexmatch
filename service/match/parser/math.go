// parser
// @author LanguageY++2013 2023/4/3 18:00
// @company soulgame
package parser

import (
	"math"
)

//SumFloat64 float64求和
func SumFloat64(values []float64) (sum float64) {
	for _, v := range values {
		sum += v
	}

	return
}

//AvgFloat64 float64 切片求平均值
func AvgFloat64(values []float64) (avg float64) {
	totalValue := SumFloat64(values)
	return totalValue / float64(len(values))
}

//FlattenFloat2DArray   扁平化二维float数组
func FlattenFloat2DArray(array2d [][]float64) (out []float64) {
	for _, array := range array2d {
		out = append(out, array...)
	}

	return out
}

//FlattenFloat2DArrayWithPtr   扁平化二维float数组
func FlattenFloat2DArrayWithPtr(array2d []*[]float64) (out []float64) {
	for _, array := range array2d {
		out = append(out, *array...)
	}

	return out
}

//MinFloat64 获取float数组最小值
func MinFloat64(values []float64) float64 {
	min := math.MaxFloat64
	for _, value := range values {
		if value < min {
			min = value
		}
	}

	return min
}

//MaxFloat64 获取float数组最大值
func MaxFloat64(values []float64) float64 {
	max := -math.MaxFloat64
	for _, value := range values {
		if value > max {
			max = value
		}
	}

	return max
}