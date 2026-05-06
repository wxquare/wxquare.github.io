package ulid

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"time"
)

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Next(ctx context.Context, prefix string) (string, error) {
	var entropy [10]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return "", err
	}
	var data [16]byte
	ms := uint64(time.Now().UnixMilli())
	data[0] = byte(ms >> 40)
	data[1] = byte(ms >> 32)
	data[2] = byte(ms >> 24)
	data[3] = byte(ms >> 16)
	data[4] = byte(ms >> 8)
	data[5] = byte(ms)
	copy(data[6:], entropy[:])
	encoded := encodeCrockford(data)
	if prefix == "" {
		return encoded, nil
	}
	return prefix + "_" + encoded, nil
}

func encodeCrockford(data [16]byte) string {
	hi := binary.BigEndian.Uint64(data[0:8])
	lo := binary.BigEndian.Uint64(data[8:16])
	value := uint128{hi: hi, lo: lo}
	var out [26]byte
	for i := 25; i >= 0; i-- {
		rem := value.mod32()
		out[i] = crockford[rem]
		value.div32()
	}
	return string(out[:])
}

type uint128 struct {
	hi uint64
	lo uint64
}

func (u *uint128) mod32() byte {
	return byte(u.lo & 31)
}

func (u *uint128) div32() {
	u.lo = (u.lo >> 5) | (u.hi << 59)
	u.hi >>= 5
}
