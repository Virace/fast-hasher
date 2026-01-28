// Package quickxorhash provides the quickXorHash algorithm which is a
// quick, simple non-cryptographic hash algorithm that works by XORing
// the bytes in a circular-shifting fashion.
//
// It is used by Microsoft Onedrive for Business to hash data.
//
// See: https://docs.microsoft.com/en-us/onedrive/developer/code-snippets/quickxorhash
//
// This code was ported from rclone:
// https://github.com/rclone/rclone/blob/master/backend/onedrive/quickxorhash/quickxorhash.go
//
// Original C implementation from https://github.com/namazso/QuickXorHash
// Licensed under BSD Zero Clause License.
package quickxorhash

import (
	"crypto/subtle"
	"hash"
)

const (
	// BlockSize is the preferred size for hashing
	BlockSize = 64
	// Size of the output checksum
	Size        = 20
	shift       = 11
	widthInBits = 8 * Size
	dataSize    = shift * widthInBits
)

type quickXorHash struct {
	data [dataSize]byte
	size uint64
}

// New returns a new hash.Hash computing the quickXorHash checksum.
func New() hash.Hash {
	return &quickXorHash{}
}

// xor dst with src
func xorBytes(dst, src []byte) int {
	return subtle.XORBytes(dst, src, dst)
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (q *quickXorHash) Write(p []byte) (n int, err error) {
	var i int
	// fill last remain
	lastRemain := q.size % dataSize
	if lastRemain != 0 {
		i += xorBytes(q.data[lastRemain:], p)
	}

	if i != len(p) {
		for len(p)-i >= dataSize {
			i += xorBytes(q.data[:], p[i:])
		}
		xorBytes(q.data[:], p[i:])
	}
	q.size += uint64(len(p))
	return len(p), nil
}

// Calculate the current checksum
func (q *quickXorHash) checkSum() (h [Size + 1]byte) {
	for i := range dataSize {
		shift := (i * 11) % 160
		shiftBytes := shift / 8
		shiftBits := shift % 8
		shifted := int(q.data[i]) << shiftBits
		h[shiftBytes] ^= byte(shifted)
		h[shiftBytes+1] ^= byte(shifted >> 8)
	}
	h[0] ^= h[20]

	// XOR the file length with the least significant bits in little endian format
	d := q.size
	h[Size-8] ^= byte(d >> (8 * 0))
	h[Size-7] ^= byte(d >> (8 * 1))
	h[Size-6] ^= byte(d >> (8 * 2))
	h[Size-5] ^= byte(d >> (8 * 3))
	h[Size-4] ^= byte(d >> (8 * 4))
	h[Size-3] ^= byte(d >> (8 * 5))
	h[Size-2] ^= byte(d >> (8 * 6))
	h[Size-1] ^= byte(d >> (8 * 7))

	return h
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (q *quickXorHash) Sum(b []byte) []byte {
	hash := q.checkSum()
	return append(b, hash[:Size]...)
}

// Reset resets the Hash to its initial state.
func (q *quickXorHash) Reset() {
	*q = quickXorHash{}
}

// Size returns the number of bytes Sum will return.
func (q *quickXorHash) Size() int {
	return Size
}

// BlockSize returns the hash's underlying block size.
func (q *quickXorHash) BlockSize() int {
	return BlockSize
}

// Sum returns the quickXorHash checksum of the data.
func Sum(data []byte) (h [Size]byte) {
	var d quickXorHash
	_, _ = d.Write(data)
	s := d.checkSum()
	copy(h[:], s[:])
	return h
}
