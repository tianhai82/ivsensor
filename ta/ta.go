package ta

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/piquette/finance-go"
	"github.com/shopspring/decimal"
	"github.com/tianhai82/ivsensor/model"
)

func ATRCandles(bars []model.Candle, period int) (float64, error) {
	if len(bars) <= period {
		return 0.0, errors.New("period should be smaller than length of bars")
	}
	trueRanges := make([]float64, 0, len(bars))
	for i, bar := range bars {
		if i == 0 {
			continue
		}
		tr := TRCandles(bar, bars[i-1])
		trueRanges = append(trueRanges, tr)
	}
	return EMA(trueRanges, period)
}
func TRCandles(currentBar, prevBar model.Candle) float64 {
	a := currentBar.High - currentBar.Low
	b := math.Abs(currentBar.High - prevBar.Close)
	c := math.Abs(currentBar.Low - prevBar.Close)
	f, _ := max(a, b, c)
	return f
}
func max(vars ...float64) (float64, error) {
	if len(vars) == 0 {
		return 0.0, fmt.Errorf("at least 1 number")
	}
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}
	return max, nil
}

func TrueRangePercentile(bars []finance.ChartBar, percentile float64) (float64, error) {
	trueRanges := make([]float64, 0, len(bars))
	for i, bar := range bars {
		if i == 0 {
			continue
		}
		tr := TR(bar, bars[i-1])
		trueRanges = append(trueRanges, tr)
	}
	return Percentile(trueRanges, percentile)
}

func ATR(bars []finance.ChartBar, period int) (float64, error) {
	if len(bars) <= period {
		return 0.0, errors.New("period should be smaller than length of bars")
	}
	trueRanges := make([]float64, 0, len(bars))
	for i, bar := range bars {
		if i == 0 {
			continue
		}
		tr := TR(bar, bars[i-1])
		trueRanges = append(trueRanges, tr)
	}
	return EMA(trueRanges, period)
}
func ATRP(bars []finance.ChartBar, period int) (float64, error) {
	if len(bars) <= period {
		return 0.0, errors.New("period should be smaller than length of bars")
	}
	trueRangePercents := make([]float64, 0, len(bars))
	for i, bar := range bars {
		if i == 0 {
			continue
		}
		tr, err := TRNormalized(bar, bars[i-1])
		if err != nil {
			continue
		}
		trueRangePercents = append(trueRangePercents, tr)
	}
	return EMA(trueRangePercents, period)
}

func TR(currentBar, prevBar finance.ChartBar) float64 {
	a := currentBar.High.Sub(currentBar.Low)
	b := currentBar.High.Sub(prevBar.Close).Abs()
	c := currentBar.Low.Sub(prevBar.Close).Abs()
	f, _ := decimal.Max(a, b, c).Float64()
	return f
}

func TRNormalized(currentBar, prevBar finance.ChartBar) (float64, error) {
	a := currentBar.High.Sub(currentBar.Low)
	b := currentBar.High.Sub(prevBar.Close).Abs()
	c := currentBar.Low.Sub(prevBar.Close).Abs()
	prev := prevBar.Close
	if prev.Equal(decimal.NewFromFloat(0.0)) {
		return 0.0, fmt.Errorf("prev bar closed at zero")
	}
	f, _ := decimal.Max(a, b, c).Div(prev).Float64()
	return f, nil
}

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

func Percentile(values []float64, percentile float64) (float64, error) {

	if len(values) == 0 {
		return 0.0, fmt.Errorf("no values provided")
	}
	if percentile == 0.5 && len(values) == 1 {
		return values[0], nil
	}

	sort.Float64s(values)

	index := percentile * float64(len(values)-1)

	if index < 0 || index >= float64(len(values)) {
		return 0.0, fmt.Errorf("percentile must be between 0 to 1")
	}

	lower := int(math.Floor(index))
	upper := lower + 1
	weight := index - math.Floor(index)

	return values[lower]*(1-weight) + values[upper]*weight, nil
}
