package shortning

import (
	"strings"
	"sync/atomic"
	"time"
)

var lastserial int64 = time.Now().UnixNano()

func NextUnique() (nxsrl int64) {
	for {
		if nxsrl = time.Now().UnixNano(); atomic.CompareAndSwapInt64(&lastserial, atomic.LoadInt64(&lastserial), nxsrl) {
			break
		}
		time.Sleep(1 * time.Nanosecond)
	}
	return
}

func EncodeUnique(number int64, getunique func(int64) (int64, bool)) (int64, string) {
	if getunique == nil {
		return number, Encode(number)
	}
	unqnf, isunq := getunique(number)
	for !isunq {
		unqnf, isunq = getunique(unqnf)
	}
	return unqnf, Encode(unqnf)
}

func Encode(number int64) string {
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(11)

	for ; number > 0; number = number / length {
		encodedBuilder.WriteByte(alphabet[(number % length)])
	}

	return encodedBuilder.String()
}
