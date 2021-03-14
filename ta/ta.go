package ta

import (
	"errors"
	"sort"
)

func TR()

func EMA(array []float64, period int) (float64, error) {
	if len(array) < period {
		return 0.0, errors.New("array size is less than period")
	}
	mult := 2.0 / float64(period+1)
	prev := 0.0
	for _, a := range array {
		prev = ema(prev, a, mult)
	}
	return prev, nil
}
func ema(prev, current, multiplier float64) float64 {
	return current*multiplier + prev*(1.0-multiplier)
}

func PercentRank(val float64, array []float64) float64 {
	internal := make([]float64, len(array))
	copy(internal, array)

	floatslice := sort.Float64Slice(internal)
	floatslice.Sort()
	arrLen := floatslice.Len()
	idx := floatslice.Search(val)
	if floatslice[idx] == val {
		return float64(idx) / float64(arrLen-1)
	}
	if idx == 0 {
		return 0.0
	}
	if idx == arrLen-1 {
		return 1.0
	}
	prev := floatslice[idx-1]
	next := floatslice[idx]

	d := (val - prev) / (next - prev)
	return (float64(idx) - 1.0 + d) / float64(arrLen-1)
}
