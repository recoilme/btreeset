package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	"github.com/recoilme/btreeset"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	hi := []byte("hi")
	//create
	bt := &btreeset.BTreeSet{}
	//set
	replaced := bt.Set(hi)
	fmt.Println("replaced:", replaced)
	//replaced: false

	//check
	gotten := bt.Has(hi)
	fmt.Println("gotten:", gotten)
	//gotten: true

	//put 0-19
	for _, i := range rand.Perm(20) {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(i))
		bt.Set(b)
	}

	//read 3 keys from 7 descending
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(7))
	buf := bytes.Buffer{}

	limit := 0
	bt.Descend(b, func(key []byte) bool {
		k := binary.BigEndian.Uint64(key)
		buf.WriteString(fmt.Sprintf("%d ", k))
		limit++
		if limit == 3 {
			return false
		}
		return true
	})
	fmt.Println(buf.String())
	//7 6 5

	bt.Delete(b)
	buf.Reset()
	bt.Scan(func(key []byte) bool {
		if len(key) == 8 {
			k := binary.BigEndian.Uint64(key)
			buf.WriteString(fmt.Sprintf("%d ", k))
		} else {
			buf.WriteString(fmt.Sprintf("%s ", key))
		}
		return true
	})
	fmt.Println(buf.String())
	//0 1 2 3 4 5 6 8 9 10 11 12 13 14 15 16 17 18 19 hi
}
