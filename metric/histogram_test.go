package metric

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestFormatFloat(t *testing.T) {
	b := math.Inf(1)
	fmt.Println(strconv.FormatFloat(b, 'f', -1, 64))
}
