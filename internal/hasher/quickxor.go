package hasher

import (
	"hash"

	"github.com/Virace/fast-hasher/pkg/quickxorhash"
)

type quickxorHasher struct{}

func (quickxorHasher) Name() string    { return "quickxor" }
func (quickxorHasher) New() hash.Hash  { return quickxorhash.New() }
func (quickxorHasher) OutputSize() int { return quickxorhash.Size }
func (quickxorHasher) IsBase64() bool  { return true } // OneDrive uses base64

func init() {
	Register(quickxorHasher{})
}
