package btreeset

import (
	"bytes"
)

const maxItems = 255
const minItems = maxItems * 40 / 100

type item struct {
	key []byte
}

type node struct {
	numItems int
	items    [maxItems]item
	children [maxItems + 1]*node
}

// BTreeSet is an ordered set of keys
type BTreeSet struct {
	//sync.RWMutex
	height int
	root   *node
	length int
}

// Compare - binary comparator
func Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func (n *node) find(key []byte) (index int, found bool) {
	low := 0
	high := n.numItems - 1
	for low <= high {
		mid := low + ((high+1)-low)/2
		if Compare(key, n.items[mid].key) >= 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if low > 0 && Compare(n.items[low-1].key, key) == 0 {
		index = low - 1
		found = true
	} else {
		index = low
		found = false
	}
	return index, found
}

// Set or replace a key
func (tr *BTreeSet) Set(key []byte) (replaced bool) {
	//tr.Lock()
	//defer tr.Unlock()
	if tr.root == nil {
		tr.root = new(node)
		tr.root.items[0] = item{key}
		tr.root.numItems = 1
		tr.length = 1
		return
	}
	replaced = tr.root.set(key, tr.height)
	if replaced {
		return
	}
	if tr.root.numItems == maxItems {
		n := tr.root
		right, median := n.split(tr.height)
		tr.root = new(node)
		tr.root.children[0] = n
		tr.root.items[0] = median
		tr.root.children[1] = right
		tr.root.numItems = 1
		tr.height++
	}
	tr.length++
	return
}

func (n *node) split(height int) (right *node, median item) {
	right = new(node)
	median = n.items[maxItems/2]
	copy(right.items[:maxItems/2], n.items[maxItems/2+1:])
	if height > 0 {
		copy(right.children[:maxItems/2+1], n.children[maxItems/2+1:])
	}
	right.numItems = maxItems / 2
	if height > 0 {
		for i := maxItems/2 + 1; i < maxItems+1; i++ {
			n.children[i] = nil
		}
	}
	for i := maxItems / 2; i < maxItems; i++ {
		n.items[i] = item{}
	}
	n.numItems = maxItems / 2
	return
}

func (n *node) set(key []byte, height int) (replaced bool) {
	i, found := n.find(key)
	if found {
		return true
	}
	if height == 0 {
		for j := n.numItems; j > i; j-- {
			n.items[j] = n.items[j-1]
		}
		n.items[i] = item{key}
		n.numItems++
		return false
	}
	replaced = n.children[i].set(key, height-1)
	if replaced {
		return
	}
	if n.children[i].numItems == maxItems {
		right, median := n.children[i].split(height - 1)
		copy(n.children[i+1:], n.children[i:])
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = median
		n.children[i+1] = right
		n.numItems++
	}
	return
}

// Scan all items in tree
func (tr *BTreeSet) Scan(iter func(key []byte) bool) {
	if tr.root != nil {
		tr.root.scan(iter, tr.height)
	}
}

func (n *node) scan(iter func(key []byte) bool, height int) bool {
	if height == 0 {
		for i := 0; i < n.numItems; i++ {
			if !iter(n.items[i].key) {
				return false
			}
		}
		return true
	}
	for i := 0; i < n.numItems; i++ {
		if !n.children[i].scan(iter, height-1) {
			return false
		}
		if !iter(n.items[i].key) {
			return false
		}
	}
	return n.children[n.numItems].scan(iter, height-1)
}

// Has return tue if key exists
func (tr *BTreeSet) Has(key []byte) (gotten bool) {
	//tr.RLock()
	//defer tr.RUnlock()
	if tr.root == nil {
		return
	}
	return tr.root.has(key, tr.height)
}

func (n *node) has(key []byte, height int) (gotten bool) {
	i, found := n.find(key)
	if found {
		return true
	}
	if height == 0 {
		return false
	}
	return n.children[i].has(key, height-1)
}

// Len returns the number of items in the tree
func (tr *BTreeSet) Len() int {
	return tr.length
}

// Delete a value for a key
func (tr *BTreeSet) Delete(key []byte) (deleted bool) {
	if tr.root == nil {
		return
	}
	_, deleted = tr.root.delete(false, key, tr.height)
	if !deleted {
		return
	}

	if tr.root.numItems == 0 {
		tr.root = tr.root.children[0]
		tr.height--
	}
	tr.length--
	if tr.length == 0 {
		tr.root = nil
		tr.height = 0
	}
	return
}

func (n *node) delete(max bool, key []byte, height int) (prev item, deleted bool) {
	i, found := 0, false
	if max {
		i, found = n.numItems-1, true
	} else {
		i, found = n.find(key)
	}
	if height == 0 {
		if found {
			prev = n.items[i]
			// found the items at the leaf, remove it and return.
			copy(n.items[i:], n.items[i+1:n.numItems])
			n.items[n.numItems-1] = item{}
			n.children[n.numItems] = nil
			n.numItems--
			return prev, true
		}
		return item{}, false
	}

	if found {
		if max {
			i++
			prev, deleted = n.children[i].delete(true, nil, height-1)
		} else {
			prev = n.items[i]
			maxItem, _ := n.children[i].delete(true, nil, height-1)
			n.items[i] = maxItem
			deleted = true
		}
	} else {
		prev, deleted = n.children[i].delete(max, key, height-1)
	}
	if !deleted {
		return
	}
	if n.children[i].numItems < minItems {
		if i == n.numItems {
			i--
		}
		if n.children[i].numItems+n.children[i+1].numItems+1 < maxItems {
			// merge left + item + right
			n.children[i].items[n.children[i].numItems] = n.items[i]
			copy(n.children[i].items[n.children[i].numItems+1:],
				n.children[i+1].items[:n.children[i+1].numItems])
			if height > 1 {
				copy(n.children[i].children[n.children[i].numItems+1:],
					n.children[i+1].children[:n.children[i+1].numItems+1])
			}
			n.children[i].numItems += n.children[i+1].numItems + 1
			copy(n.items[i:], n.items[i+1:n.numItems])
			copy(n.children[i+1:], n.children[i+2:n.numItems+1])
			n.items[n.numItems] = item{}
			n.children[n.numItems+1] = nil
			n.numItems--
		} else if n.children[i].numItems > n.children[i+1].numItems {
			// move left -> right
			copy(n.children[i+1].items[1:],
				n.children[i+1].items[:n.children[i+1].numItems])
			if height > 1 {
				copy(n.children[i+1].children[1:],
					n.children[i+1].children[:n.children[i+1].numItems+1])
			}
			n.children[i+1].items[0] = n.items[i]
			if height > 1 {
				n.children[i+1].children[0] =
					n.children[i].children[n.children[i].numItems]
			}
			n.children[i+1].numItems++
			n.items[i] = n.children[i].items[n.children[i].numItems-1]
			n.children[i].items[n.children[i].numItems-1] = item{}
			if height > 1 {
				n.children[i].children[n.children[i].numItems] = nil
			}
			n.children[i].numItems--
		} else {
			// move right -> left
			n.children[i].items[n.children[i].numItems] = n.items[i]
			if height > 1 {
				n.children[i].children[n.children[i].numItems+1] =
					n.children[i+1].children[0]
			}
			n.children[i].numItems++
			n.items[i] = n.children[i+1].items[0]
			copy(n.children[i+1].items[:],
				n.children[i+1].items[1:n.children[i+1].numItems])
			if height > 1 {
				copy(n.children[i+1].children[:],
					n.children[i+1].children[1:n.children[i+1].numItems+1])
			}
			n.children[i+1].numItems--
		}
	}
	return
}

// Ascend the tree within the range [pivot, last]
func (tr *BTreeSet) Ascend(pivot []byte, iter func(key []byte) bool) {
	if tr.root != nil {
		tr.root.ascend(pivot, iter, tr.height)
	}
}

func (n *node) ascend(pivot []byte, iter func(key []byte) bool, height int) bool {
	i, found := n.find(pivot)
	if !found {
		if height > 0 {
			if !n.children[i].ascend(pivot, iter, height-1) {
				return false
			}
		}
	}
	for ; i < n.numItems; i++ {
		if !iter(n.items[i].key) {
			return false
		}
		if height > 0 {
			if !n.children[i+1].scan(iter, height-1) {
				return false
			}
		}
	}
	return true
}

// Reverse all items in tree
func (tr *BTreeSet) Reverse(iter func(key []byte) bool) {
	if tr.root != nil {
		tr.root.reverse(iter, tr.height)
	}
}

func (n *node) reverse(iter func(key []byte) bool, height int) bool {
	if height == 0 {
		for i := n.numItems - 1; i >= 0; i-- {
			if !iter(n.items[i].key) {
				return false
			}
		}
		return true
	}
	if !n.children[n.numItems].reverse(iter, height-1) {
		return false
	}
	for i := n.numItems - 1; i >= 0; i-- {
		if !iter(n.items[i].key) {
			return false
		}
		if !n.children[i].reverse(iter, height-1) {
			return false
		}
	}
	return true
}

// Descend the tree within the range [pivot, first]
func (tr *BTreeSet) Descend(pivot []byte, iter func(key []byte) bool) {
	if tr.root != nil {
		tr.root.descend(pivot, iter, tr.height)
	}
}

func (n *node) descend(pivot []byte, iter func(key []byte) bool, height int) bool {
	i, found := n.find(pivot)
	if !found {
		if height > 0 {
			if !n.children[i].descend(pivot, iter, height-1) {
				return false
			}
		}
		i--
	}
	for ; i >= 0; i-- {
		if !iter(n.items[i].key) {
			return false
		}
		if height > 0 {
			if !n.children[i].reverse(iter, height-1) {
				return false
			}
		}
	}
	return true
}
