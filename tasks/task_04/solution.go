package main

import "math"

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}

	n := len(nums)

	stats := Stats{
		Min: math.MaxInt64,
		Max: math.MinInt64,
	}

	for i := 1; i < n; i++ {
		stats.Count++

		diffValue := nums[i] - nums[i-1]

		stats.Sum += diffValue

		stats.Max = max(stats.Max, diffValue)
		stats.Min = min(stats.Min, diffValue)
	}

	return stats
}
