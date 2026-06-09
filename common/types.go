package common

import (
	rand2 "math/rand/v2"
	"strconv"
)

const (
	hashLength    = 32
	addressLength = 20
)

type (
	Hash    [hashLength]byte
	Address [addressLength]byte
)

var (
	INITIAL_VERSION = Hash{}
	SELF_VERSION    = BytesToHash([]byte("SELF"))
)

func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Hex() string   { return "0x" + Bytes2Hex(h[:]) }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// Sets the hash to the value of b. If b is larger than len(h) it will panic
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-hashLength:]
	}

	copy(h[hashLength-len(b):], b)
}

// Set string `s` to h. If s is larger than len(h) it will panic
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h to other
func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

func EmptyHash(h Hash) bool {
	return h == Hash{}
}

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func StringToAddress(s string) Address { return BytesToAddress([]byte(s)) }
func HexToAddress(s string) Address    { return BytesToAddress(FromHex(s)) }

// Get the string representation of the underlying address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }
func (a Address) Hex() string   { return "0x" + Bytes2Hex(a[:]) }

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-addressLength:]
	}
	copy(a[addressLength-len(b):], b)
}

// Set string `s` to a. If s is larger than len(a) it will panic
func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

func HashToAddress(hash Hash) Address {
	addr := Address{}
	copy(addr[:], hash[:])
	return addr
}

func AddressLess(addr1, addr Address) bool {
	for i := 0; i < addressLength; i++ {
		if addr1[i] < addr[i] {
			return true
		} else if addr1[i] == addr[i] {
			continue
		} else {
			break
		}
	}
	return false
}

func GenerateRandomHash() Hash {
	randomBytes := strconv.Itoa(rand2.Int())
	return BytesToHash([]byte(randomBytes))
}
