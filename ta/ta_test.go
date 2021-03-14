package ta

import (
	"math"
	"testing"
)

func TestEMA(t *testing.T) {
	arr := []float64{
		2, 4, 5, 4, 6, 3, 5,
	}
	_, err := EMA(arr, 8)
	if err == nil {
		t.Errorf("should give error for period greater than array size")
	}
	out, err := EMA(arr, 7)
	if err != nil {
		t.Errorf("should give error")
	}
	if math.Round(out*10000)/10000 != math.Round(3.799927*10000)/10000 {
		t.Errorf("expected %f, but got %f", 3.799927, out)
	}
	out, err = EMA(arr, 6)
	if err != nil {
		t.Errorf("should give error")
	}
	if math.Round(out*10000)/10000 != 3.99220 {
		t.Errorf("expected %f, but got %f", 3.99220, out)
	}
}

func TestPercentRank(t *testing.T) {
	arr := []float64{2,
		4,
		4,
		9,
		34,
		12,
		6,
		9,
		20,
		7,
	}
	a := math.Round(PercentRank(2, arr)*1000) / 1000
	if a != 0.0 {
		t.Errorf("pr 2 should be 0.0 but received %f", a)
	}
	b := math.Round(PercentRank(4, arr)*1000) / 1000
	if b != 0.111 {
		t.Errorf("pr 4 should be 0.111 but received %f", b)
	}
	c := math.Round(PercentRank(10, arr)*1000) / 1000
	if c != 0.704 {
		t.Errorf("pr 10 should be 0.704 but received %f", c)
	}
	d := math.Round(PercentRank(12, arr)*1000) / 1000
	if d != 0.778 {
		t.Errorf("pr 12 should be 0.778 but received %f", d)
	}
	e := math.Round(PercentRank(18, arr)*1000) / 1000
	if e != 0.861 {
		t.Errorf("pr 18 should be 0.861 but received %f", e)
	}
	f := math.Round(PercentRank(6, arr)*1000) / 1000
	if f != 0.333 {
		t.Errorf("pr 6 should be 0.333 but received %f", f)
	}
	g := math.Round(PercentRank(9, arr)*1000) / 1000
	if g != 0.556 {
		t.Errorf("pr 9 should be 0.556 but received %f", g)
	}
}
