package shortning

import (
	"errors"
	"math"
	"strings"
)

func Decode(encoded string) (int64, error) {
	var number int64

	for i, symbol := range encoded {
		alphabeticPosition := strings.IndexRune(alphabet, symbol)

		if alphabeticPosition == -1 {
			return int64(alphabeticPosition), errors.New("invalid character: " + string(symbol))
		}
		number += int64(alphabeticPosition) * int64(math.Pow(float64(length), float64(i)))
	}

	return number, nil
}
