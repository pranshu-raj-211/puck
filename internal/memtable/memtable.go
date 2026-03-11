package memtable

import (
	"puck/internal/skiplist"
)

type Memtable interface {
	Get(key []byte) (value []byte, found bool, tombstone []byte)
	Set(key []byte, value []byte)
	Delete(key []byte)
	Dump(dir string) error
	New(maxBytes int) *Memtable
	IsFull() bool
}

type InMem struct {
	s *skiplist.Skiplist
	maxBytes int
	currentBytes int
}

func New(maxBytes int) (*InMem){
	skiplist:=skiplist.New(
		18,
		0.0,	// becomes 1/math.E instead
	)
	m := InMem{
		s:skiplist,
		maxBytes: 64_000_000, // roughly 64 mb
		currentBytes: 0,
	}
	return &m
}

func (m *InMem) IsFull() (bool){
	return m.currentBytes>=m.maxBytes
}

func (m *InMem) Get(key []byte)([]byte, bool, bool){
	value, found, tombstone := m.s.Search(key)
	// TODO: may want to use a nice user message later, or do it in the engine
	return value, found, tombstone
}

func (m *InMem) Set(key, value []byte) {
	// TODO: update byte counter (check for isfull and dump are not done here)
	m.s.Insert(key, value, false)
}

func (m *InMem) Delete(key []byte) {
	m.s.Insert(key, nil, true)
}

func (m *InMem) Dump(dir string) error{
	// TODO: dump code (serialize, file, write, fsync dir, fsync file)
	return nil
}