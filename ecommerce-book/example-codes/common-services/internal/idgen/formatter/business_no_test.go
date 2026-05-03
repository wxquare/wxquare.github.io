package formatter

import (
	"strings"
	"testing"
	"time"
)

func TestBusinessNumberIncludesPrefixDateAndCheckDigit(t *testing.T) {
	f := NewBusinessNumberFormatter()
	got := f.Format("ORD", 1928475629384753152, time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC))
	if !strings.HasPrefix(got, "ORD20260429") {
		t.Fatalf("business number = %s", got)
	}
	if strings.Contains(got, "1928475629384753152") {
		t.Fatalf("business number exposes raw id: %s", got)
	}
	if len(got) < len("ORD20260429A0") {
		t.Fatalf("business number too short: %s", got)
	}
}
