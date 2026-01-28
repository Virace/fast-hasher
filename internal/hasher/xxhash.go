package hasher

import (
	"encoding/binary"
	"hash"

	"github.com/zeebo/xxh3"
)

// xxh3Hasher implements 64-bit XXH3
type xxh3Hasher struct{}

func (xxh3Hasher) Name() string    { return "xxh3" }
func (xxh3Hasher) New() hash.Hash  { return &xxh3Hash64{h: xxh3.New()} }
func (xxh3Hasher) OutputSize() int { return 8 }
func (xxh3Hasher) IsBase64() bool  { return false }

// xxh3Hash64 wraps xxh3.Hasher to implement hash.Hash
type xxh3Hash64 struct {
	h *xxh3.Hasher
}

func (x *xxh3Hash64) Write(p []byte) (n int, err error) {
	return x.h.Write(p)
}

func (x *xxh3Hash64) Sum(b []byte) []byte {
	sum := x.h.Sum64()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], sum)
	return append(b, buf[:]...)
}

func (x *xxh3Hash64) Reset() {
	x.h.Reset()
}

func (x *xxh3Hash64) Size() int {
	return 8
}

func (x *xxh3Hash64) BlockSize() int {
	return 64
}

// xxh128Hasher implements 128-bit XXH3
type xxh128Hasher struct{}

func (xxh128Hasher) Name() string    { return "xxh128" }
func (xxh128Hasher) New() hash.Hash  { return &xxh3Hash128{h: xxh3.New()} }
func (xxh128Hasher) OutputSize() int { return 16 }
func (xxh128Hasher) IsBase64() bool  { return false }

// xxh3Hash128 wraps xxh3.Hasher to implement hash.Hash for 128-bit output
type xxh3Hash128 struct {
	h *xxh3.Hasher
}

func (x *xxh3Hash128) Write(p []byte) (n int, err error) {
	return x.h.Write(p)
}

func (x *xxh3Hash128) Sum(b []byte) []byte {
	sum := x.h.Sum128()
	var buf [16]byte
	binary.BigEndian.PutUint64(buf[:8], sum.Hi)
	binary.BigEndian.PutUint64(buf[8:], sum.Lo)
	return append(b, buf[:]...)
}

func (x *xxh3Hash128) Reset() {
	x.h.Reset()
}

func (x *xxh3Hash128) Size() int {
	return 16
}

func (x *xxh3Hash128) BlockSize() int {
	return 64
}

func init() {
	Register(xxh3Hasher{})
	Register(xxh128Hasher{})
}
