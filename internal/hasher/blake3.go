package hasher

import (
	"hash"

	"github.com/zeebo/blake3"
)

type blake3Hasher struct{}

func (blake3Hasher) Name() string    { return "blake3" }
func (blake3Hasher) New() hash.Hash  { return blake3.New() }
func (blake3Hasher) OutputSize() int { return 32 }
func (blake3Hasher) IsBase64() bool  { return false }

func init() {
	Register(blake3Hasher{})
}
