package formatter

import (
	"strconv"
	"strings"
	"time"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type BusinessNumberFormatter struct{}

func NewBusinessNumberFormatter() *BusinessNumberFormatter {
	return &BusinessNumberFormatter{}
}

func (f *BusinessNumberFormatter) Format(prefix string, rawID int64, now time.Time) string {
	body := strings.ToUpper(strconv.FormatInt(rawID, 36))
	base := prefix + now.Format("20060102") + body
	return base + string(alphabet[checksum(base)%len(alphabet)])
}

func checksum(s string) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += int(s[i]) * (i + 1)
	}
	return sum
}
