package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"hash/crc32"
)

// Standard library hashers

type md5Hasher struct{}

func (md5Hasher) Name() string    { return "md5" }
func (md5Hasher) New() hash.Hash  { return md5.New() }
func (md5Hasher) OutputSize() int { return md5.Size }
func (md5Hasher) IsBase64() bool  { return false }

type sha1Hasher struct{}

func (sha1Hasher) Name() string    { return "sha1" }
func (sha1Hasher) New() hash.Hash  { return sha1.New() }
func (sha1Hasher) OutputSize() int { return sha1.Size }
func (sha1Hasher) IsBase64() bool  { return false }

type sha256Hasher struct{}

func (sha256Hasher) Name() string    { return "sha256" }
func (sha256Hasher) New() hash.Hash  { return sha256.New() }
func (sha256Hasher) OutputSize() int { return sha256.Size }
func (sha256Hasher) IsBase64() bool  { return false }

type sha512Hasher struct{}

func (sha512Hasher) Name() string    { return "sha512" }
func (sha512Hasher) New() hash.Hash  { return sha512.New() }
func (sha512Hasher) OutputSize() int { return sha512.Size }
func (sha512Hasher) IsBase64() bool  { return false }

type crc32Hasher struct{}

func (crc32Hasher) Name() string    { return "crc32" }
func (crc32Hasher) New() hash.Hash  { return crc32.NewIEEE() }
func (crc32Hasher) OutputSize() int { return crc32.Size }
func (crc32Hasher) IsBase64() bool  { return false }

func init() {
	Register(md5Hasher{})
	Register(sha1Hasher{})
	Register(sha256Hasher{})
	Register(sha512Hasher{})
	Register(crc32Hasher{})
}
