package responsetime

import (
	"math"
	"math/rand"
	"time"
)

//GetResponseTime : Exported function for getting response time
func GetResponseTime(meanValue float64) float64 {
	// ExpFloat64 returns an exponentially distributed float64 in the range (0, +math.MaxFloat64]
	// with an exponential distribution whose rate parameter (lambda) is 1 and whose mean is 1/lambda (1) from the default Source
	// parameter (lambda) := 1 / meanValue
	// Reference: https://stackoverflow.com/questions/2106503/pseudorandom-number-generator-exponential-distribution

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	u := r1.Float64()
	rateParameter := 1 / meanValue
	x := -1 * math.Log(1-u) / rateParameter
	return x
}
