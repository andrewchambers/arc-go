// Package arc implements the Adaptive Replacement Cache
/*

https://www.usenix.org/legacy/events/fast03/tech/full_papers/megiddo/megiddo.pdf

This code is a straight-forward translation of the Python implementation at
http://code.activestate.com/recipes/576532-adaptive-replacement-cache-in-python/
modified to make the O(n) list operations O(1).

*/
package arc

import (
	"fmt"
	"strings"
)

// Callbacks used by the cache to fill the cache.
type Callbacks[K comparable, V any] struct {
	// GetValue is called to retrieve a value from the cache.
	// If it returns an error, the Get operation fails with an error.
	GetValue func(K) (V, error)
	// OnEvict is called when a key is evicted from the cache.
	// If it returns an error, the Get operation fails with an error.
	OnEvict func(K, V) error
}

// Cache is a type implementing an Adaptive Replacement Cache,
// it is NOT threadsafe without additional synchronization.
type Cache[K comparable, V any] struct {
	Callbacks Callbacks[K, V]

	data map[K]V

	cap  int
	part int

	t1 *clist[K]
	t2 *clist[K]
	b1 *clist[K]
	b2 *clist[K]
}

func New[K comparable, V any](size int, callbacks Callbacks[K, V]) *Cache[K, V] {
	if callbacks.GetValue == nil {
		panic("expected a GetValue callback")
	}
	if callbacks.OnEvict == nil {
		callbacks.OnEvict = func(K, V) error { return nil }
	}
	return &Cache[K, V]{
		Callbacks: callbacks,
		data:      make(map[K]V),
		cap:       size,
		t1:        newClist[K](),
		t2:        newClist[K](),
		b1:        newClist[K](),
		b2:        newClist[K](),
	}
}

func (c *Cache[K, V]) replace(key K, part int) error {
	var t, b *clist[K]
	if (c.t1.Len() > 0 && c.b2.Has(key) && c.t1.Len() == part) || (c.t1.Len() > part) {
		t = c.t1
		b = c.b1
	} else {
		t = c.t2
		b = c.b2
	}
	old := t.Last()
	err := c.Callbacks.OnEvict(old, c.data[old])
	if err != nil {
		return err
	}
	t.Pop()
	b.PushFront(old)
	delete(c.data, old)
	return nil
}

func (c *Cache[K, V]) Get(key K) (V, error) {

	if elt := c.t1.Lookup(key); elt != nil {
		c.t1.Remove(key, elt)
		c.t2.PushFront(key)
		return c.data[key], nil
	}

	if elt := c.t2.Lookup(key); elt != nil {
		c.t2.MoveToFront(elt)
		return c.data[key], nil
	}

	result, err := c.Callbacks.GetValue(key)
	if err != nil {
		return result, err
	}

	if elt := c.b1.Lookup(key); elt != nil {
		part := min(c.cap, c.part+max(c.b2.Len()/c.b1.Len(), 1))
		err := c.replace(key, part)
		if err != nil {
			return result, err
		}
		c.part = part
		c.b1.Remove(key, elt)
		c.t2.PushFront(key)
		c.data[key] = result
		return result, nil
	}

	if elt := c.b2.Lookup(key); elt != nil {
		part := max(0, c.part-max(c.b1.Len()/c.b2.Len(), 1))
		err := c.replace(key, part)
		if err != nil {
			return result, err
		}
		c.part = part
		c.b2.Remove(key, elt)
		c.t2.PushFront(key)
		c.data[key] = result
		return result, nil
	}

	if c.t1.Len()+c.b1.Len() == c.cap {
		if c.t1.Len() < c.cap {
			err := c.replace(key, c.part)
			if err != nil {
				return result, err
			}
			c.b1.Pop()
		} else {
			pop := c.t1.Last()
			err := c.Callbacks.OnEvict(pop, c.data[pop])
			if err != nil {
				return result, err
			}
			c.t1.Pop()
			delete(c.data, pop)
		}
	} else {
		total := c.t1.Len() + c.b1.Len() + c.t2.Len() + c.b2.Len()
		if total >= c.cap {
			if total == (2 * c.cap) {
				removed := c.b2.Pop()
				err := c.replace(key, c.part)
				if err != nil {
					// Rollback removal.
					c.b2.PushBack(removed)
					return result, err
				}
			} else {
				err := c.replace(key, c.part)
				if err != nil {
					return result, err
				}
			}
		}
	}

	c.t1.PushFront(key)
	c.data[key] = result

	return result, nil
}

func (c *Cache[K, V]) DebugDump() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Cache DebugDump:\n")
	fmt.Fprintf(&sb, "  data: %v\n", c.data)
	fmt.Fprintf(&sb, "  cap: %d\n", c.cap)
	fmt.Fprintf(&sb, "  part: %d\n", c.part)

	fmt.Fprintf(&sb, "  t1:\n")
	sb.WriteString(c.t1.DebugDump())
	fmt.Fprintf(&sb, "  t2:\n")
	sb.WriteString(c.t2.DebugDump())
	fmt.Fprintf(&sb, "  b1:\n")
	sb.WriteString(c.b1.DebugDump())
	fmt.Fprintf(&sb, "  b2:\n")
	sb.WriteString(c.b2.DebugDump())

	return sb.String()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
