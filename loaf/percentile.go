package loaf

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/montanaflynn/stats"
)

var (
	EmptyInputErr = stats.EmptyInputErr
	BoundsErr     = stats.BoundsErr
	NaNErr        = stats.NaNErr
)

type Float64Data []float64

// Percentile finds the relative standing in a slice of floats
func Percentile(input []float64, percent float64) (percentile float64, err error) {
	length := len(input)
	if length == 0 {
		return math.NaN(), EmptyInputErr
	}

	if length == 1 {
		return input[0], nil
	}

	if percent <= 0 || percent > 100 {
		fmt.Println(`percent out of range `, percent)
		return math.NaN(), BoundsErr
	}

	// Start by sorting a copy of the slice
	c := sortedCopy(input)

	// Multiply percent by length of input
	index := (percent/100)*float64(len(c)-1) + 1

	// Check if the index is a whole number
	if index == float64(int64(index)) {

		// Convert float to int
		i := int(index)

		// Find the value at the index
		percentile = c[i-1]

	} else if index > 1 {

		// Convert float to int via truncation
		i := int(index)

		// Find the average of the index and following values
		percentile = (c[i-1] + c[i]) / 2.

	} else {
		fmt.Println(` index out of range `, index)
		return math.NaN(), BoundsErr
	}

	return percentile, nil

}

// float64ToInt rounds a float64 to an int
func float64ToInt(input float64) (output int) {
	r, _ := Round(input, 0)
	return int(r)
}

// unixnano returns nanoseconds from UTC epoch
func unixnano() int64 {
	return time.Now().UTC().UnixNano()
}

// copyslice copies a slice of float64s
func copyslice(input Float64Data) Float64Data {
	s := make(Float64Data, len(input))
	copy(s, input)
	return s
}

// sortedCopy returns a sorted copy of float64s
func sortedCopy(input Float64Data) (copy Float64Data) {
	copy = copyslice(input)
	sort.Float64s(copy)
	return
}

// sortedCopyDif returns a sorted copy of float64s
// only if the original data isn't sorted.
// Only use this if returned slice won't be manipulated!
func sortedCopyDif(input Float64Data) (copy Float64Data) {
	if sort.Float64sAreSorted(input) {
		return input
	}
	copy = copyslice(input)
	sort.Float64s(copy)
	return
}

func Round(input float64, places int) (rounded float64, err error) {

	// If the float is not a number
	if math.IsNaN(input) {
		return math.NaN(), NaNErr
	}

	// Find out the actual sign and correct the input for later
	sign := 1.0
	if input < 0 {
		sign = -1
		input *= -1
	}

	// Use the places arg to get the amount of precision wanted
	precision := math.Pow(10, float64(places))

	// Find the decimal place we are looking to round
	digit := input * precision

	// Get the actual decimal number as a fraction to be compared
	_, decimal := math.Modf(digit)

	// If the decimal is less than .5 we round down otherwise up
	if decimal >= 0.5 {
		rounded = math.Ceil(digit)
	} else {
		rounded = math.Floor(digit)
	}

	// Finally we do the math to actually create a rounded number
	return rounded / precision * sign, nil
}
