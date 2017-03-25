package bitset

import (
	"fmt"
	"log"
	"testing"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestCount(t *testing.T) {
	bs := New(10)

	err := bs.Set(1, true)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.Set(2, true)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.Set(3, false)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.Set(9, true)
	if err != nil {
		t.Fatal(err)
	}
	n := bs.Count()
	if n != 3 {
		t.Errorf("Expected %d bit at 1, got %d : bs=%s", 3, n, bs.String())
	}
}

func TestAppend(t *testing.T) {
	bs := New(10)
	bs.Append(true)
	bs.Append(true)
	bs.Append(false)
	bs.Append(false)
	bs.Append(true)
	if bs.Uint8(0, 5) != 0x13 {
		t.Errorf("Expected 0x13, got 0x%X", bs.Uint8(0, 5))
	}
}

func TestPruneFirst(t *testing.T) {
	bs := New(9)
	err := bs.SetRangeFromString(0, 6, "1101011")
	if err != nil {
		t.Fatal(err)
	}
	bs.PruneFirst(2)
	if bs.String() != "01011" {
		t.Fatal("pruned bitset not as expected")
	}
}

func TestAny(t *testing.T) {
	bs := New(10)
	if bs.Any() {
		t.Error("Any is true while all bits are zero")
	}
	bs.Set(2, true)

	if !bs.Any() {
		t.Error("Any is false while it's expected to be true")
	}
}

func TestNew(t *testing.T) {
	bs := New(-1)
	if bs != nil {
		t.Fatal(fmt.Sprintf("was expecting nil as size is negative"))
	}
	bs = New(0)
	if bs != nil {
		t.Fatal(fmt.Sprintf("was expecting nil as size is < 1"))
	}
	bs = New(100)
	if bs == nil {
		t.Fatal(fmt.Sprintf("was not expecting a nil here, value is sane"))
	}
}

func TestSet(t *testing.T) {
	bs := New(10)
	err := bs.Set(0, true)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.Set(2, true)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.Set(20, true)
	if err != nil {
		t.Errorf("could not set a bit past the current size of the bitset...")
	}
	if bs.Size() != 32 {
		t.Errorf("Found size of %d while expecting 32 : has the grow method changed ?", bs.Size())
	}
	if bs.Length() != 21 {
		t.Errorf("Length of %d while expecting 21", bs.Length())
	}
	assertPanic(t, func() {
		bs.Set(-1, true)
	})
}

func TestGet(t *testing.T) {
	bs := New(10)
	err := bs.Set(0, true)
	if err != nil {
		t.Fatal(err)
	}
	bs.Set(2, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bs.Get(2) {
		t.Errorf("did not get back true for pos=2")
	}
	if !bs.Get(0) {
		t.Errorf("did not get back true for pos=0")
	}
	if bs.Get(1) {
		t.Errorf("did not get back false for pos=1")
	}
	assertPanic(t, func() {
		bs := New(10)
		bs.Get(100)
	})
}

func TestClear(t *testing.T) {
	bs := New(1) // FromUint32(0x1008000)
	bs.Set(24, true)
	if bs.Length() != 25 {
		t.Errorf("Expected len of 25, got %d", bs.Length())
	}
	bs.Clear()
	if bs.Length() != 0 {
		t.Errorf("Got a non zero length = %d", bs.Length())
	}
}

func TestString(t *testing.T) {
	bs := New(8)
	bs.Set(1, true)
	bs.Set(3, true)
	bs.Set(5, true)
	s := bs.String()
	expected := "010101"
	if s != expected {
		t.Errorf("got %s while expecting %s", s, expected)
	}
}

func TestFromString(t *testing.T) {
	bs := FromString("00011x")
	if bs != nil {
		t.Fatal("bs should be nil")
	}
	expected := "01011011"
	bs = FromString(expected)
	if bs.String() != expected {
		t.Fatal(fmt.Sprintf("got %s, expected %s", bs.String(), expected))
	}
}

func TestRangeFromString(t *testing.T) {
	bs := FromString("110011")
	err := bs.SetRangeFromString(2, 3, "11")
	if err != nil {
		t.Fatal(err)
	}
	if bs.String() != "111111" {
		t.Fatal("not as expected")
	}
	err = bs.SetRangeFromString(2, 3, "x-")
	if err == nil {
		t.Errorf("invalid string should not produce a valid bitset")
	}
	err = bs.SetRangeFromString(4, 1, "101")
	if err == nil {
		t.Errorf("inverted range should not produce a valid bitset")
	}
	err = bs.SetRangeFromString(32, 38, "1100")
	if err == nil {
		t.Errorf("bit range smaller than string len should trigger an error")
	}
	err = bs.SetRangeFromString(32, 38, "1100110")
	if err != nil {
		t.Fatal(err)
	}
	bs = FromString("abcd")
	if bs != nil {
		t.Errorf("invalid string should not produce a valid bitset")
	}
}

func TestFromIntegers(t *testing.T) {
	var v uint64 = 0xF0F8FCFEFF3F3F1F
	bs := FromUint64(v)
	s := "1111100011111100111111001111111101111111001111110001111100001111"
	if bs.String() != s {
		t.Error("incorrect textual representation")
		t.Errorf("expected: %s", s)
		t.Errorf("got     : %s", bs.String())
	}
	x := bs.Uint64(0, 63)
	if x != v {
		t.Error("incorrect value")
	}
	bs = FromUint8(0x13)
	if bs.Uint8(0, 7) != 0x13 {
		t.Errorf("expected 0x13, got %X", bs.Uint8(0, 7))
	}
	bs = FromUint16(0x8000)
	if bs.Uint16(0, 15) != 0x8000 {
		t.Errorf("expected 0x8000, got %X", bs.Uint16(0, 15))
	}
	bs = FromUint32(0xF0008000)
	if bs.Uint32(0, 31) != 0xF0008000 {
		t.Errorf("expected 0xF0008000, got %X", bs.Uint32(0, 31))
	}

	bs = New(150)

	err := bs.SetRangeFromUint8(0, 9, 0)
	if err == nil {
		t.Errorf("too big a range should trigger an error")
	}
	err = bs.SetRangeFromUint16(9, 32, 0)
	if err == nil {
		t.Errorf("too big a range should trigger an error")
	}
	err = bs.SetRangeFromUint32(24, 57, 0)
	if err == nil {
		t.Errorf("too big a range should trigger an error")
	}
	err = bs.SetRangeFromUint64(56, 122, 0)
	if err == nil {
		t.Errorf("too big a range should trigger an error")
	}
	err = bs.SetRangeFromUint8(0, 7, 0xFF)
	if err != nil {
		t.Errorf("should have been able to assign 255 to a 8-bit bitset !", err)
	}
	if bs.Length() != 8 {
		t.Errorf("bs should be of length 8 and is %d", bs.Length())
	}

}

func TestRangeFromUint64(t *testing.T) {
	var v uint64 = 0xF0F8FCFEFF3F3F1F
	bs := FromUint64(v)
	bs.SetRangeFromUint64(60, 62, 0)
	if bs.Uint64(0, 63) != 0x80F8FCFEFF3F3F1F {
		t.Error("incorrect value after SetRangeFromUint64")
	}
}

func TestRangeFromIntegers(t *testing.T) {
	bs := New(64)
	err := bs.SetRangeFromUint8(0, 5, 0x13)
	if err != nil {
		t.Fatal(err)
	}
	bs.Set(8, true)
	err = bs.SetRangeFromUint8(20, 23, 0xF)
	if err != nil {
		t.Fatal(err)
	}
	err = bs.SetRangeFromUint32(29, 48, 0xAAAAA)
	if err != nil {
		t.Fatal(err)
	}
	if bs.Uint64(0, 63) != 0x1555540F00113 {
		t.Error("incorrect value after SetRangeFromIntegers")
	}
}

func TestFromBytes(t *testing.T) {
	bs := New(150) // any number would work, the size will be set by SetFromBytes
	bs.SetFromBytes([]byte{0xfe, 0x5a, 0x1e, 0xda})
	if bs.Uint32(0, 31) != 0xDA1E5AFE {
		t.Errorf("incorrect value after SetFromBytes")
	}
	if bs.Size() != 32 {
		t.Errorf("incorrect size of %d after SetFromBytes", bs.Size())
	}
	if bs.Length() != 32 {
		t.Errorf("incorrect Length of %d after SetFromBytes", bs.Length())
	}
}

func TestIsEqual(t *testing.T) {
	b1 := FromString("110011")
	b2 := FromString("110011")
	if !b1.IsEqual(*b2) {
		t.Fatal("b1 should be equal to b2")
	}
	b2 = FromString("1010")
	if b1.IsEqual(*b2) {
		t.Fatal("b2 should not be equal to b1")
	}
}

func TestSearch(t *testing.T) {
	s := FromString("001100101101011001")
	p := FromString("11001")
	pos, err := s.Search(*p)
	if err != nil {
		log.Fatal(err)
	}
	if pos != 2 {
		log.Fatal("Did not get expected position 2")
	}
}

func TestSub(t *testing.T) {
	str := "110011"
	bs := FromString(str)
	b := bs.Subset(2, 4)
	if b.String() != "001" {
		t.Fatal("Slice not correct")
	}
	b.Set(1, true)
	if b.String() != "011" {
		t.Fatal("Slice not correct")
	}
}

func TestUint64(t *testing.T) {
	bs := FromString("110011")
	v := bs.Uint64(0, 5)
	if v != 51 {
		t.Fatal(fmt.Sprintf("v should be 51 and is %d", v))
	}
	v = bs.Uint64(0, 1)
	if v != 3 {
		t.Fatal("v should be 3")
	}
}

func TestGrow(t *testing.T) {
	bs := New(32)
	if bs.grow(16) != nil {
		t.Errorf("Should not grow if size is already enough !")
	}
	if bs.grow(MaxSize+1) == nil {
		t.Errorf("Requesting too big a size should have failed")
		t.Errorf("the package limit is at %d MB", MaxSize/8/1024/1024)
	}

}

func BenchmarkSetRangeFromUint8(b *testing.B) {

	bs := New(8)
	var value uint8 = 1
	// we test the worst case, i.e. all bits to be set
	b.Run("regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs.SetRangeFromUint8(0, 7, value)
		}
	})
	b.Run("fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs.SetRangeFromUint8Fast(0, 7, value)
		}
	})
	b.Run("FromBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs.SetFromBytes([]byte{value})
		}
	})
}

func BenchmarkSetRangeFromUint32(b *testing.B) {
	bs := New(32)
	var value uint32 = 1
	b.Run("regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs.SetRangeFromUint32(0, 31, value)
		}
	})
	b.Run("fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs.SetRangeFromUint32Fast(0, 31, value)
		}
	})
}
