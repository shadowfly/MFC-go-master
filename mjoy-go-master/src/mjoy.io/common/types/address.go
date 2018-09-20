////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: address.go
// @Date: 2018/05/08 17:39:08
////////////////////////////////////////////////////////////////////////////////

package types

import (
	"encoding/hex"
	"fmt"
	"mjoy.io/common/types/util"
	mjoyhex "mjoy.io/common/types/util/hex"
	"mjoy.io/utils/crypto/sha3"
	"math/big"
	"reflect"
)

//go:generate msgp

const (
	AddressLength = 20
)

var (
	addressType int8
)

// Address represents the 20 byte address of an mjoy account.
type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func StringToAddress(s string) Address { return BytesToAddress([]byte(s)) }
func BigToAddress(b *big.Int) Address  { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address    { return BytesToAddress(util.FromHex(s)) }

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Mjoy address or not.
func IsHexAddress(s string) bool {
	if util.HasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && util.IsHex(s)
}

// Get the string representation of the underlying address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	sha := sha3.NewKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// Here, we'll pick an arbitrary number between
// 0 and 127 that isn't already in use
func (*Address) ExtensionType() int8 {
	return addressType
}

// We'll always use 20 bytes to encode the data
func (*Address) Len() int {
	return AddressLength
}

// MarshalBinaryTo simply copies the value
// of the bytes into 'b'
func (a *Address) MarshalBinaryTo(b []byte) error {
	copy(b, (*a)[:])
	return nil
}

// UnmarshalBinary copies the value of 'b'
// into the BYTE8 object. (We might want to add
// a sanity check here later that len(b) <= HashLength.)
func (a *Address) UnmarshalBinary(b []byte) error {
	// TODO: check b, only hex, len <= AddressLength
	if len(b) <= AddressLength {
		copy((*a)[:], b)
		return nil
	}

	return ErrBytesTooLong
}

// for json marshal
func (a Address) MarshalText() ([]byte, error) {
	// TODO:
	return mjoyhex.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return mjoyhex.UnmarshalFixedText("Address", input, a[:])
}

// for json unmarshal
func (a *Address) UnmarshalJSON(b []byte) error {
	// TODO:
	return mjoyhex.UnmarshalFixedJSON(reflect.TypeOf(Address{}), b, a[:])
}

// for format print
func (a Address) Format(s fmt.State, c rune) {
	switch c {
	case 'x' | 'X':
		fmt.Fprintf(s, "%#x", a[:])
	default:
		fmt.Fprintf(s, "%"+string(c), a[:])
	}
}

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}
