package skiplist

import (
	"bytes"
	"errors"
	"math"
	"math/rand"
)

type Node struct {
	key   []byte
	value []byte
	next  []*Node
}

type KeyValuePair struct {
	key   []byte
	value []byte
}

type RangeIterator struct {
	current *Node
	endKey  []byte
}

type Skiplist struct {
	head       *Node
	level      int
	maxLevel   int
	p          float64
	probTable  []float64
	randsource *rand.Rand
	update     []*Node
}

// Creates a new skiplist for the given max level and promotion probability
func New(maxLevel int, p float64) *Skiplist {
	// maxLevel 0 means a normal linked list
	if maxLevel < 0 {
		maxLevel = 0
	}
	// default probability should be 1/2 - coin flip
	if p <= 0.0 || p >= 1.0 {
		p = 1 / math.E
	}

	// keeping head's key as -1 does not matter as head's key is never compared
	// head node's key can be anything, we generally give it a negative infinity value for sentinel behavior
	head := &Node{
		key:  nil,
		next: make([]*Node, maxLevel+1),
	}

	return &Skiplist{
		head:       head,
		level:      0,
		maxLevel:   maxLevel,
		p:          p,
		probTable:  computeProbTable(maxLevel, p),
		randsource: rand.New(rand.NewSource(42)),
		update:     make([]*Node, maxLevel+1),
	}
}

// Search for node with given key in the skiplist
func (s *Skiplist) Search(key []byte) (interface{}, error) {
	current := s.head

	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && bytes.Compare(current.next[i].key, key) < 0 {
			current = current.next[i]
		}
	}

	current = current.next[0]
	if current != nil && bytes.Compare(current.key, key) == 0 {
		return current.value, nil
	}

	return 0, errors.New("key not found")
}

// Insert or update a key-value pair
func (s *Skiplist) Insert(key []byte, value []byte) {
	current := s.head

	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && bytes.Compare(current.next[i].key, key) < 0 {
			current = current.next[i]
		}
		s.update[i] = current
	}

	current = current.next[0]
	if current != nil && bytes.Compare(current.key, key) == 0 {
		current.value = value
		return
	}

	nodeLevel := s.randomLevel()
	if nodeLevel > s.level {
		for i := s.level + 1; i <= nodeLevel; i++ {
			s.update[i] = s.head
		}
		s.level = nodeLevel
	}

	newNode := &Node{
		key:   key,
		value: value,
		next:  make([]*Node, nodeLevel+1),
	}

	for i := 0; i <= nodeLevel; i++ {
		newNode.next[i] = s.update[i].next[i]
		s.update[i].next[i] = newNode
	}
}

// Delete node with given key in the skiplist
func (s *Skiplist) Delete(key []byte) bool {
	current := s.head

	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && bytes.Compare(current.next[i].key, key) < 0 {
			current = current.next[i]
		}
		s.update[i] = current
	}

	current = current.next[0]
	if current == nil || bytes.Compare(current.key, key) != 0 {
		return false
	}

	for i := 0; i <= s.level; i++ {
		if s.update[i].next[i] != current {
			break
		}
		s.update[i].next[i] = current.next[i]
	}

	for s.level > 0 && s.head.next[s.level] == nil {
		s.level--
	}

	return true
}

// Returns a list of key value pairs between [startKey, endKey] (both inclusive)
func (s *Skiplist) RangeQuery(startKey, endKey []byte) []KeyValuePair {
	// startKey > endKey returns +1
	if bytes.Compare(startKey, endKey) > 0 {
		return nil
	}

	results := make([]KeyValuePair, 0)
	current := s.head
	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && bytes.Compare(current.next[i].key, startKey) < 0 {
			current = current.next[i]
		}
	}

	current = current.next[0]
	if bytes.Compare(current.key, startKey) != 0 {
		return nil
	}

	for current != nil && bytes.Compare(current.key, endKey) <= 0 {
		results = append(results, KeyValuePair{
			key:   current.key,
			value: current.value,
		})
		current = current.next[0]
	}

	return results
}

func (s *Skiplist) RangeQueryIterator(startKey, endKey []byte) *RangeIterator {
	if bytes.Compare(startKey, endKey) > 0 {
		return &RangeIterator{current: nil, endKey: endKey}
	}

	current := s.head
	for i := s.level; i >= 0; i-- {
		for current.next[i] != nil && bytes.Compare(current.next[i].key, startKey) < 0 {
			current = current.next[i]
		}
	}

	// current.next[0] is now the first node with key >= startKey
	return &RangeIterator{
		current: current.next[0],
		endKey:  endKey,
	}
}

func (it *RangeIterator) Valid() bool {
	return it.current != nil && bytes.Compare(it.current.key, it.endKey) <= 0
}

func (it *RangeIterator) Next() {
	if it.current != nil {
		it.current = it.current.next[0]
	}
}

func (it *RangeIterator) Key() []byte {
	return it.current.key
}

func (it *RangeIterator) Value() []byte {
	return it.current.value
}

// Returns a random level based on a geometric probability distribution with prob p
// a fraction p of nodes from the current level will be promoted to next upper level
func (s *Skiplist) randomLevel() int {
	r := s.randsource.Float64()
	level := 0
	for level < s.maxLevel && r < s.probTable[level] {
		level++
	}
	return level
}

// compute the probabilities at which we will promote in advance
func computeProbTable(maxLevel int, p float64) []float64 {
	probTable := make([]float64, maxLevel+1)
	currentProb := 1.0
	for i := 0; i <= maxLevel; i++ {
		probTable[i] = currentProb
		currentProb *= p
	}
	return probTable
}
