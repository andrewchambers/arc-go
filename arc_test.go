package arc

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"testing"

	"github.com/andrewchambers/list-go"
)

func TestARCBlackBox(t *testing.T) {

	cacheCallbacks := Callbacks[int, int]{
		GetValue: func(k int) (int, error) {
			if rand.Float64() < 0.5 {
				return 0, errors.New("GetValue failed")
			}
			return k, nil
		},
		OnEvict: func(k, v int) error {
			if rand.Float64() < 0.5 {
				return errors.New("Evict failed")
			}
			return nil
		},
	}

	cacheSize := int(5)

	cache := New[int, int](cacheSize, cacheCallbacks)

	for _, vBound := range []int{1, cacheSize, cacheSize * 2, cacheSize * 10} {
		for i := 0; i < 25000; i += 1 {
			x := rand.Int() % vBound
			for {
				state1 := cache.DebugDump()
				v, err := cache.Get(x)
				if err != nil {
					endState2 := cache.DebugDump()
					if state1 != endState2 {
						t.Logf("state1:\n%s", state1)
						t.Logf("state2:\n%s", endState2)
						t.Fatalf("error rollback failed after error - %s", err)
					}
					continue
				}
				if v != x {
					t.Fatal("bad value")
				}
				break
			}
		}
	}
}

func TestARCInternal(t *testing.T) {

	tst := []uint32{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		11, 12, 13, 14,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 39,
		38, 37, 36, 35, 34, 33, 32, 16, 17, 11, 41,
	}

	cache := New[string, []byte](10, Callbacks[string, []byte]{
		GetValue: func(k string) ([]byte, error) { return []byte(k), nil },
	})

	for _, v := range tst {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		cache.Get(string(b[:]))
	}

	checkList(t, "t1", cache.t1.l, []byte{41})
	checkList(t, "t2", cache.t2.l, []byte{11, 17, 16, 32, 33, 34, 35, 36, 37})
	checkList(t, "b1", cache.b1.l, []byte{31, 30})
	checkList(t, "b2", cache.b2.l, []byte{38, 39, 19, 18, 15, 14, 13, 12})

	if cache.part != 5 {
		t.Errorf("bad p: got=%v want=5", cache.part)
	}
}

func checkList(t *testing.T, name string, l *list.List[string], expected []byte) {

	idx := 0

	for e := l.Front(); e != nil; e = e.Next() {
		b := []byte(e.Value)
		if b[0] != expected[idx] {
			t.Errorf("list %s failed idx %d: got=%d want=%d\n", name, idx, b[0], expected[idx])
		}
		idx++
	}
}

func BenchmarkEviction(b *testing.B) {

	cacheCallbacks := Callbacks[int, int]{
		GetValue: func(k int) (int, error) {
			return k, nil
		},
	}

	cacheSize := 10
	cache := New[int, int](cacheSize, cacheCallbacks)

	// Prepopulate cache.
	for i := 0; i < cacheSize; i += 1 {
		cache.Get(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i += 1 {
		cache.Get(i + cacheSize)
	}
}
