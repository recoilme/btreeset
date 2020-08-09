package btreeset

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	seed := time.Now().UnixNano()
	fmt.Printf("seed: %d\n", seed)
	rand.Seed(seed)
}

func randKeys(N int) (keys []string) {
	format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", N-1)))
	for _, i := range rand.Perm(N) {
		keys = append(keys, fmt.Sprintf(format, i))
	}
	return
}

const flatLeaf = true

func (tr *BTreeSet) print() {
	tr.root.print(0, tr.height)
}

func (n *node) print(level, height int) {
	if n == nil {
		println("NIL")
		return
	}
	if height == 0 && flatLeaf {
		fmt.Printf("%s", strings.Repeat("  ", level))
	}
	for i := 0; i < n.numItems; i++ {
		if height > 0 {
			n.children[i].print(level+1, height-1)
		}
		if height > 0 || (height == 0 && !flatLeaf) {
			fmt.Printf("%s%v\n", strings.Repeat("  ", level), n.items[i].key)
		} else {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%s", n.items[i].key)
		}
	}
	if height == 0 && flatLeaf {
		fmt.Printf("\n")
	}
	if height > 0 {
		n.children[n.numItems].print(level+1, height-1)
	}
}

func (tr *BTreeSet) deepPrint() {
	fmt.Printf("%#v\n", tr)
	tr.root.deepPrint(0, tr.height)
}

func (n *node) deepPrint(level, height int) {
	if n == nil {
		fmt.Printf("%s %#v\n", strings.Repeat("  ", level), n)
		return
	}
	fmt.Printf("%s count: %v\n", strings.Repeat("  ", level), n.numItems)
	fmt.Printf("%s items: %v\n", strings.Repeat("  ", level), n.items)
	if height > 0 {
		fmt.Printf("%s child: %v\n", strings.Repeat("  ", level), n.children)
	}
	if height > 0 {
		for i := 0; i < n.numItems; i++ {
			n.children[i].deepPrint(level+1, height-1)
		}
		n.children[n.numItems].deepPrint(level+1, height-1)
	}
}

func stringsEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDescend(t *testing.T) {
	var tr BTreeSet
	var count int
	tr.Descend([]byte("1"), func(key []byte) bool {
		count++
		return true
	})
	if count > 0 {
		t.Fatalf("expected 0, got %v", count)
	}
	var keys []string
	for i := 0; i < 1000; i += 10 {
		keys = append(keys, fmt.Sprintf("%03d", i))
		tr.Set([]byte(keys[len(keys)-1]))
	}
	var exp []string
	tr.Reverse(func(key []byte) bool {
		exp = append(exp, string(key))
		return true
	})
	for i := 999; i >= 0; i-- {
		var key string
		key = fmt.Sprintf("%03d", i)
		var all []string
		tr.Descend([]byte(key), func(key []byte) bool {
			all = append(all, string(key))
			return true
		})
		for len(exp) > 0 && key < exp[0] {
			exp = exp[1:]
		}
		var count int
		tr.Descend([]byte(key), func(key []byte) bool {
			if count == (i+1)%maxItems {
				return false
			}
			count++
			return true
		})
		if count > len(exp) {
			t.Fatalf("expected 1, got %v", count)
		}

		if !stringsEquals(exp, all) {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("all: %v\n", all)
			t.Fatal("mismatch")
		}
	}
}

func TestAscend(t *testing.T) {
	var tr BTreeSet
	var count int
	tr.Ascend([]byte("1"), func(key []byte) bool {
		count++
		return true
	})
	if count > 0 {
		t.Fatalf("expected 0, got %v", count)
	}
	var keys []string
	for i := 0; i < 1000; i += 10 {
		keys = append(keys, fmt.Sprintf("%03d", i))
		tr.Set([]byte(keys[len(keys)-1]))
	}
	exp := keys
	for i := -1; i < 1000; i++ {
		var key string
		if i == -1 {
			key = ""
		} else {
			key = fmt.Sprintf("%03d", i)
		}
		var all []string
		tr.Ascend([]byte(key), func(key []byte) bool {
			all = append(all, string(key))
			return true
		})

		for len(exp) > 0 && key > exp[0] {
			exp = exp[1:]
		}
		var count int
		tr.Ascend([]byte(key), func(key []byte) bool {
			if count == (i+1)%maxItems {
				return false
			}
			count++
			return true
		})
		if count > len(exp) {
			t.Fatalf("expected 1, got %v", count)
		}
		if !stringsEquals(exp, all) {
			t.Fatal("mismatch")
		}
	}
}

func TestBTreeSet(t *testing.T) {
	N := 10_000
	var tr BTreeSet
	keys := randKeys(N)

	// insert all items
	for _, key := range keys {
		replaced := tr.Set([]byte(key))
		if replaced {
			t.Fatal("expected false")
		}
	}

	// check length
	if tr.Len() != len(keys) {
		t.Fatalf("expected %v, got %v", len(keys), tr.Len())
	}

	// get each value
	for _, key := range keys {
		gotten := tr.Has([]byte(key))
		if !gotten {
			t.Fatal("expected true")
		}
	}

	// scan all items
	var last []byte
	all := make(map[string]interface{})
	tr.Scan(func(key []byte) bool {
		if Compare(key, last) < 0 { //key <= last {
			t.Fatal("out of order")
		}
		last = key
		all[string(key)] = key
		return true
	})
	if len(all) != len(keys) {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// reverse all items
	var prev []byte
	all = make(map[string]interface{})
	tr.Reverse(func(key []byte) bool {
		if prev != nil && Compare(key, prev) >= 0 { //key >= prev {
			t.Fatal("out of order")
		}
		prev = key
		all[string(key)] = prev
		return true
	})
	if len(all) != len(keys) {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// try to get an invalid item
	gotten := tr.Has([]byte("invalid"))
	if gotten {
		t.Fatal("expected false")
	}

	// scan and quit at various steps
	for i := 0; i < 100; i++ {
		var j int
		tr.Scan(func(key []byte) bool {
			if j == i {
				return false
			}
			j++
			return true
		})
	}

	// reverse and quit at various steps
	for i := 0; i < 100; i++ {
		var j int
		tr.Reverse(func(key []byte) bool {
			if j == i {
				return false
			}
			j++
			return true
		})
	}

	// delete half the items
	for _, key := range keys[:len(keys)/2] {
		deleted := tr.Delete([]byte(key))
		if !deleted {
			tr.deepPrint()
			t.Fatal("expected true")
		}
	}

	// check length
	if tr.Len() != len(keys)/2 {
		t.Fatalf("expected %v, got %v", len(keys)/2, tr.Len())
	}

	// try delete half again
	for _, key := range keys[:len(keys)/2] {
		deleted := tr.Delete([]byte(key))
		if deleted {
			t.Fatal("expected false")
		}
	}

	// try delete half again
	for _, key := range keys[:len(keys)/2] {
		deleted := tr.Delete([]byte(key))
		if deleted {
			t.Fatal("expected false")
		}
	}

	// check length
	if tr.Len() != len(keys)/2 {
		t.Fatalf("expected %v, got %v", len(keys)/2, tr.Len())
	}

	// scan items
	last = nil
	all = make(map[string]interface{})
	tr.Scan(func(key []byte) bool {
		if Compare(key, last) <= 0 {
			t.Fatal("out of order")
		}
		last = key
		all[string(key)] = last
		return true
	})
	if len(all) != len(keys)/2 {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// replace second half
	for _, key := range keys[len(keys)/2:] {
		replaced := tr.Set([]byte(key))
		if !replaced {
			t.Fatal("expected true")
		}
	}

	// delete next half the items
	for _, key := range keys[len(keys)/2:] {
		deleted := tr.Delete([]byte(key))
		if !deleted {
			t.Fatal("expected true")
		}
	}

	// check length
	if tr.Len() != 0 {
		t.Fatalf("expected %v, got %v", 0, tr.Len())
	}

	// do some stuff on an empty tree
	gotten = tr.Has([]byte(keys[0]))
	if gotten {
		t.Fatal("expected false")
	}
	tr.Scan(func(key []byte) bool {
		t.Fatal("should not be reached")
		return true
	})
	tr.Reverse(func(key []byte) bool {
		t.Fatal("should not be reached")
		return true
	})

	var deleted bool
	deleted = tr.Delete([]byte("invalid"))
	if deleted {
		t.Fatal("expected false")
	}
}

func BenchmarkTidwallSequentialSet(b *testing.B) {
	var tr BTreeSet
	keys := randKeys(b.N)
	sort.Strings(keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Set([]byte(keys[i]))
	}
}

func BenchmarkTidwallSequentialGet(b *testing.B) {
	var tr BTreeSet
	keys := randKeys(b.N)
	sort.Strings(keys)
	for i := 0; i < b.N; i++ {
		tr.Set([]byte(keys[i]))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Has([]byte(keys[i]))
	}
}

func BenchmarkTidwallRandomSet(b *testing.B) {
	var tr BTreeSet
	keys := nrandbin(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
}

func nrandbin(n int) [][]byte {
	i := make([][]byte, n)
	for ind := range i {
		bin, _ := KeyToBinary(rand.Int())
		i[ind] = bin
	}
	return i
}
func BenchmarkTidwallRandomGet(b *testing.B) {
	var tr BTreeSet
	keys := nrandbin(b.N)
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Has(keys[i])
	}
}

func TestBTreeSetOne(t *testing.T) {
	var tr BTreeSet
	tr.Set([]byte("1"))
	tr.Delete([]byte("1"))
	tr.Set([]byte("1"))
	tr.Delete([]byte("1"))
	tr.Set([]byte("1"))
	tr.Delete([]byte("1"))
}

func TestBTreeSet256(t *testing.T) {
	var tr BTreeSet
	var n int
	for j := 0; j < 2; j++ {
		for _, i := range rand.Perm(256) {
			tr.Set([]byte(fmt.Sprintf("%d", i)))
			n++
			if tr.Len() != n {
				t.Fatalf("expected 256, got %d", n)
			}
		}
		for _, i := range rand.Perm(256) {
			ok := tr.Has([]byte(fmt.Sprintf("%d", i)))
			if !ok {
				t.Fatal("expected true")
			}
		}
		for _, i := range rand.Perm(256) {
			tr.Delete([]byte(fmt.Sprintf("%d", i)))
			n--
			if tr.Len() != n {
				t.Fatalf("expected 256, got %d", n)
			}
		}
		for _, i := range rand.Perm(256) {
			ok := tr.Has([]byte(fmt.Sprintf("%d", i)))
			if ok {
				t.Fatal("expected false")
			}
		}
	}
}

func TestBTreeSetRandom(t *testing.T) {
	var count uint32
	T := runtime.NumCPU()
	D := time.Second
	N := 1000
	bkeys := make([]string, N)
	for i, key := range rand.Perm(N) {
		bkeys[i] = strconv.Itoa(key)
	}

	var wg sync.WaitGroup
	wg.Add(T)
	for i := 0; i < T; i++ {
		go func() {
			defer wg.Done()
			start := time.Now()
			for {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				keys := make([]string, len(bkeys))
				for i, key := range bkeys {
					keys[i] = key
				}
				testBTreeSetRandom(t, r, keys, &count)
				if time.Since(start) > D {
					break
				}
			}
		}()
	}
	wg.Wait()
	// println(count)
}

func shuffle(r *rand.Rand, keys []string) {
	for i := range keys {
		j := r.Intn(i + 1)
		keys[i], keys[j] = keys[j], keys[i]
	}
}

func testBTreeSetRandom(t *testing.T, r *rand.Rand, keys []string, count *uint32) {
	var tr BTreeSet
	keys = keys[:rand.Intn(len(keys))]
	shuffle(r, keys)
	for i := 0; i < len(keys); i++ {
		ok := tr.Set([]byte(keys[i]))
		if ok {
			t.Fatalf("expected nil")
		}
	}
	shuffle(r, keys)
	for i := 0; i < len(keys); i++ {
		ok := tr.Has([]byte(keys[i]))
		if !ok {
			t.Fatalf("expected '%v', got '%v'", keys[i], "")
		}
	}
	shuffle(r, keys)
	for i := 0; i < len(keys); i++ {
		ok := tr.Delete([]byte(keys[i]))
		if !ok {
			t.Fatalf("expected '%v', got '%v'", keys[i], "")
		}
		ok = tr.Has([]byte(keys[i]))
		if ok {
			//tr.deepPrint()
			t.Fatalf("expected nil %d %d %d %s", len(keys), i, tr.Len(), keys[i])
			tr.deepPrint()
			panic("")
		}
	}
	atomic.AddUint32(count, 1)
}

func TestBTreeFirstLast(t *testing.T) {
	bt := &BTreeSet{}
	bt.Set([]byte("hi"))
	//put 0-19
	for _, i := range rand.Perm(20) {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(i))
		bt.Set(b)
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, 0)
	assert.Equal(t, b, bt.First())

	assert.Equal(t, []byte("hi"), bt.Last())
}

func TestBTreePrefix(t *testing.T) {
	bt := &BTreeSet{}
	bt.Set([]byte("hi"))
	//put 0-19
	for _, i := range rand.Perm(20) {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(i))
		bt.Set(b)
	}
	var result []byte
	bt.Ascend([]byte("h"), func(key []byte) bool {
		result = key
		return false
	})
	assert.Equal(t, []byte("hi"), result)

	result = nil
	bt.Descend([]byte("h"), func(key []byte) bool {
		result = key
		return false
	})
	assert.Equal(t, []byte("hi"), result)

	bt.Descend(nil, func(key []byte) bool {
		fmt.Println(key)
		return true
	})
}
