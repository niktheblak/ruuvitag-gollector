package dewpoint

import (
	"fmt"
	"math"
)

const (
	MaxCount = 10
)

func Solve(f func(float64) float64, y, x0 float64) (float64, error) {
	x := x0
	var xNew float64
	count := 0
	for {
		if count > MaxCount {
			return 0, fmt.Errorf("solver does not converge")
		}
		dx := x / 1000.0
		z := f(x)
		xNew = x + dx*(y-z)/(f(x+dx)-z)
		if math.Abs((xNew-x)/xNew) < 0.0001 {
			return xNew, nil
		}
		x = xNew
		count++
	}
}
