package geoindex

import (
	"time"
)

// A set interface.
type set interface {
	Add(id string, value interface{})
	Get(id string) (value interface{}, ok bool)
	Remove(id string)
	Values() []interface{}
	Size() int
	Clone() set
}

// A set that contains values.
type basicSet map[string]interface{}

func newSet() set {
	return basicSet(make(map[string]interface{}))
}

// Clone creates a copy of the set where the values in clone set point to the same underlying reference as the original set
func (set basicSet) Clone() set {
	clone := basicSet(make(map[string]interface{}))
	for k, v := range set {
		clone[k] = v
	}

	return clone
}

func (set basicSet) Add(id string, value interface{}) {
	set[id] = value
}

func (set basicSet) Remove(id string) {
	delete(set, id)
}

func (set basicSet) Values() []interface{} {
	result := make([]interface{}, 0, len(set))

	for _, point := range set {
		result = append(result, point)
	}

	return result
}

func (set basicSet) Get(id string) (value interface{}, ok bool) {
	value, ok = set[id]
	return
}

func (set basicSet) Size() int {
	return len(set)
}

// An expiring set that removes the points after X minutes.
type expiringSet struct {
	values         set
	insertionOrder *queue
	expiration     Minutes
	onExpire       func(id string, value interface{})
	lastInserted   map[string]time.Time
}

type timestampedValue struct {
	id        string
	value     interface{}
	timestamp time.Time
}

// Clone panics - We currently do not allow cloning of an expiry set
func (set *expiringSet) Clone() set {
	panic("Cannot clone an expiry set")
}

func newExpiringSet(expiration Minutes) *expiringSet {
	return &expiringSet{newSet(), newQueue(1), expiration, nil, make(map[string]time.Time)}
}

func (set *expiringSet) hasExpired(time time.Time) bool {
	currentTime := getNow()
	return int(currentTime.Sub(time).Minutes()) > int(set.expiration)
}

func (set *expiringSet) expire() {
	for !set.insertionOrder.IsEmpty() {
		lastInserted := set.insertionOrder.Peek().(*timestampedValue)

		if set.hasExpired(lastInserted.timestamp) {
			set.insertionOrder.Pop()

			if set.hasExpired(set.lastInserted[lastInserted.id]) {
				set.values.Remove(lastInserted.id)

				if set.onExpire != nil {
					set.onExpire(lastInserted.id, lastInserted.value)
				}
			}
		} else {
			break
		}
	}
}

func (set *expiringSet) Add(id string, value interface{}) {
	set.expire()
	set.values.Add(id, value)
	insertionTime := getNow()
	set.lastInserted[id] = insertionTime
	set.insertionOrder.Push(&timestampedValue{id, value, insertionTime})
}

func (set *expiringSet) Remove(id string) {
	set.expire()
	set.values.Remove(id)
	delete(set.lastInserted, id)
}

func (set *expiringSet) Get(id string) (value interface{}, ok bool) {
	set.expire()
	value, ok = set.values.Get(id)
	return
}

func (set *expiringSet) Size() int {
	set.expire()
	return set.values.Size()
}

func (set *expiringSet) Values() []interface{} {
	set.expire()
	return set.values.Values()
}

func (set *expiringSet) OnExpire(onExpire func(id string, value interface{})) {
	set.onExpire = onExpire
}
