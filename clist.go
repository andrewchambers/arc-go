package arc

import (
	"fmt"
	"strings"

	"github.com/andrewchambers/list-go"
)

type clist[K comparable] struct {
	l    *list.List[K]
	keys map[K]*list.Element[K]
}

func newClist[K comparable]() *clist[K] {
	return &clist[K]{
		l:    list.New[K](),
		keys: make(map[K]*list.Element[K]),
	}
}

func (c *clist[K]) Has(key K) bool {
	_, ok := c.keys[key]
	return ok
}

func (c *clist[K]) Lookup(key K) *list.Element[K] {
	elt := c.keys[key]
	return elt
}

func (c *clist[K]) MoveToFront(elt *list.Element[K]) {
	c.l.MoveToFront(elt)
}

func (c *clist[K]) PushFront(key K) {
	elt := c.l.PushFront(key)
	c.keys[key] = elt
}

func (c *clist[K]) Remove(key K, elt *list.Element[K]) {
	delete(c.keys, key)
	c.l.Remove(elt)
}

func (c *clist[K]) PushBack(key K) {
	elt := c.l.PushBack(key)
	c.keys[key] = elt
}

func (c *clist[K]) Pop() K {
	elt := c.l.Back()
	key := elt.Value
	c.Remove(key, elt)
	return key
}

func (c *clist[K]) Last() K {
	elt := c.l.Back()
	key := elt.Value
	return key
}

func (c *clist[K]) Len() int {
	return c.l.Len()
}

func (c *clist[K]) DebugDump() string {
	var sb strings.Builder

	sb.WriteString("{")
	for e := c.l.Front(); e != nil; e = e.Next() {
		if e != c.l.Front() {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", e.Value)
	}
	sb.WriteString("}")

	return sb.String()
}
