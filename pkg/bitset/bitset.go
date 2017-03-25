// Package bitset provides a growable bitset (aka bitstring, bitmap, bitarray).
//
// Each element of the BitSet is a boolean (true or false).
//
// A bitset has a current size, which is the number of bits of space used
// in memory and a length, which is the index of the highest set bit + 1
//
package bitset

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	MaxSize = 200 * 1024 * 1024 * 8 // 200 MB in bits
)

// BitSet stores bits
type BitSet struct {
	bytes []byte
	len   int
	size  int
}

// New constructs a BitSet that can accomodate up to size bits without growing
func New(size int) *BitSet {
	if size < 1 {
		return nil
	}
	nbytes := size / 8
	if size%8 > 0 {
		nbytes++
	}
	bytes := make([]byte, nbytes)
	return &BitSet{bytes: bytes, len: 0, size: size}
}

// FromBytes return a BitSet initialized with the given byte slice
func FromBytes(bytes []byte) *BitSet {
	bs := New(1)
	bs.SetFromBytes(bytes)
	return bs
}

// FromUint64 returns a BitSet initialized with the 64-bits value
func FromUint64(v uint64) *BitSet {
	bs := New(64)
	bs.SetRangeFromUint64(0, 63, v)
	return bs
}

// FromUint32 returns a BitSet initialized with the 32-bits value
func FromUint32(v uint32) *BitSet {
	bs := New(32)
	bs.SetRangeFromUint32(0, 31, v)
	return bs
}

// FromUint16 returns a BitSet initialized with the 16-bits value
func FromUint16(v uint16) *BitSet {
	bs := New(16)
	bs.SetRangeFromUint16(0, 15, v)
	return bs
}

// FromUint8 returns a BitSet initialized with the 8-bits value
func FromUint8(v uint8) *BitSet {
	bs := New(8)
	bs.SetRangeFromUint8(0, 7, v)
	return bs
}

// FromString returns a BitSet using a string composed of 0 (meaning bit unset)
// and 1 (meaning bit set) characters
// The length of the resulting BitSet is that of the string
func FromString(s string) *BitSet {
	if !isValidString(s) {
		return nil
	}
	bs := New(len(s))
	bs.SetRangeFromString(0, bs.Size()-1, s)
	return bs
}

// Append adds a bit at the current position (Length()-1)
func (bs *BitSet) Append(val bool) error {
	return bs.Set(bs.Length(), val)
}

// Length returns the index of the highest bit used
func (bs *BitSet) Length() int {
	return bs.len
}

// Size returns the capacity of the BitSet
func (bs *BitSet) Size() int {
	return bs.size
}

// Get returns the bit value at index ix
func (bs *BitSet) Get(ix int) bool {
	if ix < 0 || ix > bs.Size() {
		panic(fmt.Sprintf("index %d is out of bounds", ix))
	}
	i := uint(ix)
	b := bs.bytes[i/8]
	i %= 8
	return ((b >> i) & 1) == 1
}

// SetNone sets all the bits to false (i.e. resets)
func (bs *BitSet) Clear() {
	for i, _ := range bs.bytes {
		bs.bytes[i] = 0
	}
	bs.len = 0
}

// grow grows the BitSet so it can accomodate at least n bits
func (bs *BitSet) grow(n int) error {
	if n < bs.size {
		return nil // already enough size
	}
	if n > MaxSize {
		return errors.New(fmt.Sprintf("BitSet.grow : trying to allocate a BitSet of more than %d MB", MaxSize))
	}
	nbytes := len(bs.bytes)
	for nbytes*8 < n {
		nbytes *= 2
	}
	extraBytes := make([]byte, nbytes-len(bs.bytes))
	bs.bytes = append(bs.bytes, extraBytes...)
	bs.size = len(bs.bytes) * 8
	return nil
}

func (bs *BitSet) SetFromBytes(bytes []byte) {
	// bs.bytes = make([]byte, len(bytes))
	copy(bs.bytes, bytes)
	bs.len = len(bytes) * 8
	bs.size = bs.len
}

// SetRangeFromString populates the bits at indices [a,b] (inclusive range)
// from the characters in the string : 0 to unset the bit (=false)
// or 1 to set the bit (=true).
//
// A string containing anything else than '0' and '1' is invalid and triggers
// an error.
func (bs *BitSet) SetRangeFromString(a, b int, s string) error {
	if !isValidString(s) {
		return errors.New(fmt.Sprintf("SetRangeFromString(a,b,s) : s=%s is not a valid bitstring", s))
	}
	if a > b {
		return errors.New(fmt.Sprintf("SetRangeFromString(a,b,s) : a=%d is not below b=%d", a, b))
	}
	if b-a+1 != len(s) {
		return errors.New(fmt.Sprintf("SetRangeFromString(a,b,s) : range a=%d,b=%d is not the same size as s=%s", a, b, s))
	}
	if a >= bs.Size() || b >= bs.Size() {
		err := bs.grow(b)
		if err != nil {
			return err
		}
	}
	var i int
	for i = 0; i < int(len(s)); i++ {
		bs.Set(i+a, s[i] == '1')
	}
	return nil
}

// SetRangeFromUint8 populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 8 obviously
func (bs *BitSet) SetRangeFromUint8(a, b int, v uint8) error {
	if a > b || b-a >= 8 {
		return errors.New(fmt.Sprintf("SetRangeFromUint8(a,b,v) : (a=%d,b=%d) out of range", a, b))
	}
	for i := 0; i <= b-a; i++ {
		if (v & (1 << byte(i))) > 0 {
			bs.Set(i+a, true)
		} else {
			bs.Set(i+a, false)
		}
	}
	return nil
}

// SetRangeFromUint8Fast populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 8 obviously. No bound check
func (bs *BitSet) SetRangeFromUint8Fast(a, b int, v uint8) {
	for i := 0; i <= b-a; i++ {
		bs.SetFast(i+a, (v&(1<<byte(i))) > 0)
	}
}

// SetRangeFromUint16 populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 16 obviously
func (bs *BitSet) SetRangeFromUint16(a, b int, v uint16) error {
	if b > bs.Size() {
		bs.grow(b)
	}
	if a > b || b-a >= 16 {
		return errors.New(fmt.Sprintf("SetRangeFromUint16(a,b,v) : (a=%d,b=%d) out of range", a, b))
	}
	for i := 0; i <= b-a; i++ {
		if (v & (1 << uint16(i))) > 0 {
			bs.Set(i+a, true)
		} else {
			bs.Set(i+a, false)
		}
	}
	return nil
}

// SetRangeFromUint32Fast populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 32 obviously
func (bs *BitSet) SetRangeFromUint32Fast(a, b int, v uint32) error {
	for i := 0; i <= b-a; i++ {
		bs.SetFast(i+a, (v&(1<<uint32(i))) > 0)
	}
	return nil
}

// SetRangeFromUint32 populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 32 obviously
func (bs *BitSet) SetRangeFromUint32(a, b int, v uint32) error {
	if b > bs.Size() {
		bs.grow(b)
	}
	if a > b || b-a >= 32 {
		return errors.New(fmt.Sprintf("SetRangeFromUint32(a,b,v) : (a=%d,b=%d) out of range", a, b))
	}
	for i := 0; i <= b-a; i++ {
		if (v & (1 << uint32(i))) > 0 {
			bs.SetFast(i+a, true)
		} else {
			bs.SetFast(i+a, false)
		}
	}
	return nil
}

// SetRangeFromUint64 populates the bits at indices [a,b] (inclusive range)
// with the bits of value v. b-a must be <= 64 obviously
func (bs *BitSet) SetRangeFromUint64(a, b int, v uint64) error {
	if a > b || b-a >= 64 {
		return errors.New(fmt.Sprintf("SetRangeFromUint64(a,b,v) : (a=%d,b=%d) out of range", a, b))
	}
	for i := 0; i <= b-a; i++ {
		if (v & (1 << uint64(i))) > 0 {
			bs.Set(i+a, true)
		} else {
			bs.Set(i+a, false)
		}
	}
	return nil
}

// Count returns the number of bits that are set
func (bs *BitSet) Count() int {
	var n int = 0
	var i int = 0
	for i = 0; i < bs.Length(); i++ {
		if bs.Get(i) {
			n++
		}
	}
	return n
}

// Any returns true if any of the bits is set
func (bs *BitSet) Any() bool {
	var i int = 0
	for i = 0; i < bs.Length(); i++ {
		if bs.Get(i) {
			return true
		}
	}
	return false
}

// SetFast sets the value of a given bit, w/o bound checking or growing
func (bs *BitSet) SetFast(pos int, val bool) {
	upos := uint(pos)
	b := &bs.bytes[upos/8]
	ix := upos % 8
	if val {
		*b |= (1 << ix)
	} else {
		*b &= ^(1 << ix)
	}
	if pos+1 > bs.len {
		bs.len = pos + 1
	}
}

// Set sets the value of a given bit
func (bs *BitSet) Set(pos int, val bool) error {
	// if pos < 0 {
	// 	panic(fmt.Sprintf("pos %d out of bounds", pos))
	// }
	if pos >= bs.Size() {
		err := bs.grow(pos)
		if err != nil {
			return err
		}
	}
	bs.SetFast(pos, val)
	return nil
}

// String returns a text representation of the BitSet
// where the LSB is on the left
func (bs *BitSet) String() string {
	s := ""
	for i := int(0); i < bs.Length(); i++ {
		if bs.Get(i) {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

// StringLSBRight returns a text representation of the BitSet
// where the LSB is on the right
func (bs *BitSet) StringLSBRight() string {
	s := ""
	for i := bs.Length() - 1; i >= 0; i-- {
		if bs.Get(i) {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

// Subset returns a subset of the BitSet.
// Subset is not a slice, i.e. it's a copy of
// the internals, not a reference
// a,b are inclusive
func (bs *BitSet) Subset(a, b int) *BitSet {
	if a >= bs.Size() || b > bs.Size() || b < a {
		return nil
	}
	sub := New(b - a + 1)
	var i int
	for i = a; i <= b; i++ {
		sub.Set(i-a, bs.Get(i))
	}
	return sub
}

func (bs *BitSet) IsEqual(bo BitSet) bool {
	if bs.Length() != bo.Length() {
		return false
	}
	var i int
	for i = 0; i < bs.Length(); i++ {
		if bs.Get(i) != bo.Get(i) {
			return false
		}
	}
	return true
}

// Search returns the first position of the given pattern,
// or an error if not found
// FIXME: this method is highly inefficient, is it necessary at all ?
func (bs *BitSet) Search(pattern BitSet) (int, error) {
	if bs.Length() < pattern.Length() {
		return 0, errors.New("BitSet too short for pattern")
	}
	var i, j int
	for i = 0; i <= bs.Length()-pattern.Length(); i++ {
		found := true
		for j = 0; j < pattern.Length(); j++ {
			vi := bs.Get(i + j)
			vj := pattern.Get(j)
			if vi != vj {
				found = false
				continue
			}
		}
		if found {
			return i, nil
		}
	}
	return 0, errors.New("not found")
}

// Uint64 converts the BitSet(a,b) into a 64-bits value, if possible
// if b is negative, it is set to the bitset length
func (bs *BitSet) Uint64(a, b int) uint64 {
	if b < 0 {
		b = bs.Length() - 1
	}
	if a < 0 || b < 0 || b-a > 64 || a >= bs.Size() || b >= bs.Size() {
		log.Fatalf("Range [%d,%d] out of range", a, b)
		return 0
	}
	var value uint64 = 0
	for i := a; i <= b; i++ {
		if bs.Get(i) {
			value += 1 << uint64(i-a)
		}
	}
	return value
}

// Uint32 converts the BitSet into a 32-bits value, if possible
// if b is negative, it is set to the bitset length
func (bs *BitSet) Uint32(a, b int) uint32 {
	if b < 0 {
		b = bs.Length() - 1
	}
	if a < 0 || b < 0 || b-a > 32 || a >= bs.Size() || b >= bs.Size() {
		log.Fatalf("Range [%d,%d] out of range", a, b)
		return 0
	}
	var value uint32 = 0
	for i := a; i <= b; i++ {
		if bs.Get(i) {
			value += 1 << uint32(i-a)
		}
	}
	return value
}

// Uint16 converts the BitSet into a 16-bits value, if possible
// if b is negative, it is set to the bitset length
func (bs *BitSet) Uint16(a, b int) uint16 {
	if b < 0 {
		b = bs.Length() - 1
	}
	if a < 0 || b < 0 || b-a > 16 || a >= bs.Size() || b >= bs.Size() {
		log.Fatalf("Range [%d,%d] out of range", a, b)
		return 0
	}
	var value uint16 = 0
	for i := a; i <= b; i++ {
		if bs.Get(i) {
			value += 1 << uint16(i-a)
		}
	}
	return value
}

// Uint8 converts the BitSet into a 8-bits value, if possible
// if b is negative, it is set to the bitset length
func (bs *BitSet) Uint8(a, b int) uint8 {
	if b < 0 {
		b = bs.Length() - 1
	}
	if a < 0 || b < 0 || b-a > 8 || a >= bs.Size() || b >= bs.Size() {
		log.Fatalf("Range [%d,%d] out of range", a, b)
		return 0
	}
	var value uint8 = 0
	for i := a; i <= b; i++ {
		if bs.Get(i) {
			value += 1 << uint8(i-a)
		}
	}
	return value
}

// Last returns a BitSet containing the last n bits of the bitset
// if bitset does not contain enough bits, nil is returned
func (bs *BitSet) Last(n int) *BitSet {
	if bs.Length() < n {
		return nil
	}
	subbs := bs.Subset(bs.Length()-n, bs.Length()-1)
	if subbs.Length() != n {
		log.Fatal("subset not of the expected len")
	}
	return subbs
}

// PruneFirst removes the first n bits from the bitset
func (bs *BitSet) PruneFirst(n int) error {
	if bs.Length() < n {
		return errors.New(fmt.Sprintf("cannot prune %d bits", n))
	}
	for i := 0; i < bs.Length()-n; i++ {
		bs.Set(i, bs.Get(i+n))
	}
	for i := bs.Length() - n; i < bs.Length(); i++ {
		bs.Set(i, bs.Get(i+n))
	}
	bs.len -= n
	return nil
}

func isValidString(s string) bool {
	if strings.Count(s, "0")+strings.Count(s, "1") != len(s) {
		return false
	}
	return true
}
