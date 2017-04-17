package geoindex

import (
	"fmt"
	"time"
)

type Minutes int

type counter interface {
	Add(point Point)
	Remove(point Point)
	Point() *CountPoint
}

type timestampedCounter struct {
	counter   accumulatingCounter
	timestamp time.Time
}

// Expiring counter.
type expiringCounter struct {
	counters   *queue
	minutes    Minutes
	count      accumulatingCounter
	newCounter func(point Point) accumulatingCounter
}

func newExpiringCounter(expiration Minutes) *expiringCounter {
	return &expiringCounter{
		newQueue(int(expiration) + 1),
		expiration,
		&singleValueAccumulatingCounter{0.0, 0.0, 0},
		newSingleValueAccumulatingCounter,
	}
}

func newExpiringMultiCounter(expiration Minutes) *expiringCounter {
	return &expiringCounter{
		newQueue(int(expiration) + 1),
		expiration,
		&multiValueAccumulatingCounter{
			&singleValueAccumulatingCounter{0.0, 0.0, 0},
			make(map[string]int),
		},
		newMultiValueCounter,
	}
}

func newExpiringAverageCounter(expiration Minutes) *expiringCounter {
	return &expiringCounter{
		newQueue(int(expiration) + 1),
		expiration,
		&averageAccumulatingCounter{
			&singleValueAccumulatingCounter{0.0, 0.0, 0},
			0.0,
		},
		newAverageAccumulatingCounter,
	}
}

func (c *expiringCounter) expire() {
	for !c.counters.IsEmpty() {
		counter := c.counters.Peek().(*timestampedCounter)
		counterAgeInMinutes := int(getNow().Sub(counter.timestamp).Minutes())

		if counterAgeInMinutes > int(c.minutes) {
			c.counters.Pop()
			c.count.Minus(counter.counter)
		} else {
			break
		}
	}
}

func (c *expiringCounter) Add(point Point) {
	c.expire()
	c.count.Plus(c.newCounter(point))

	lastCounter := c.counters.PeekBack()

	if lastCounter != nil && lastCounter.(*timestampedCounter).timestamp.Minute() == getNow().Minute() {
		lastCounter.(*timestampedCounter).counter.Add(point)
	} else {
		counter := &timestampedCounter{c.newCounter(point), getNow()}
		c.counters.Push(counter)
	}
}

func (c *expiringCounter) Remove(point Point) {
	panic("Unsupported operation. Too complicated.")
}

func (c *expiringCounter) Point() *CountPoint {
	c.expire()
	return c.count.Point()
}

func (c *expiringCounter) Count() accumulatingCounter {
	c.expire()
	return c.count
}

func (c *expiringCounter) String() string {
	return fmt.Sprintf("counters=%s minutes=%d", c.counters, c.minutes)
}

// Accumulating counter.
type accumulatingCounter interface {
	Add(point Point)
	Remove(point Point)
	Point() *CountPoint
	Plus(c accumulatingCounter)
	Minus(c accumulatingCounter)
}

// Single value counter.
func newSingleValueAccumulatingCounter(point Point) accumulatingCounter {
	return &singleValueAccumulatingCounter{point.Lat(), point.Lon(), 1}
}

type singleValueAccumulatingCounter struct {
	latSum float64
	lonSum float64
	count  int
}

func (c *singleValueAccumulatingCounter) Add(point Point) {
	c.latSum += point.Lat()
	c.lonSum += point.Lon()
	c.count++
}

func (c *singleValueAccumulatingCounter) Remove(point Point) {
	c.latSum -= point.Lat()
	c.lonSum -= point.Lon()
	c.count--
}

func (c *singleValueAccumulatingCounter) Point() *CountPoint {
	if c.count > 0 {
		return &CountPoint{&GeoPoint{"", c.latSum / float64(c.count), c.lonSum / float64(c.count)}, c.count}
	}

	return nil
}

func (c1 *singleValueAccumulatingCounter) Plus(value accumulatingCounter) {
	c2 := value.(*singleValueAccumulatingCounter)
	c1.latSum += c2.latSum
	c1.lonSum += c2.lonSum
	c1.count += c2.count
}

func (c1 *singleValueAccumulatingCounter) Minus(value accumulatingCounter) {
	c2 := value.(*singleValueAccumulatingCounter)
	c1.latSum -= c2.latSum
	c1.lonSum -= c2.lonSum
	c1.count -= c2.count
}

func (c *singleValueAccumulatingCounter) String() string {
	return fmt.Sprintf("%f %f %d", c.latSum, c.lonSum, c.count)
}

// Multi value counter.
func newMultiValueCounter(point Point) accumulatingCounter {
	values := make(map[string]int)
	values[point.Id()] = 1
	return &multiValueAccumulatingCounter{
		newSingleValueAccumulatingCounter(point).(*singleValueAccumulatingCounter),
		values,
	}
}

type multiValueAccumulatingCounter struct {
	point  *singleValueAccumulatingCounter
	values map[string]int
}

func (counter *multiValueAccumulatingCounter) Add(point Point) {
	counter.point.Add(point)
	counter.values[point.Id()] += 1
}

func (counter *multiValueAccumulatingCounter) Remove(point Point) {
	counter.point.Remove(point)
	counter.values[point.Id()] -= 1
}

func (counter *multiValueAccumulatingCounter) Point() *CountPoint {
	center := counter.point.Point()

	if center == nil {
		return nil
	}

	return &CountPoint{&GeoPoint{"", center.Lat(), center.Lon()}, counter.values}
}

func (counter *multiValueAccumulatingCounter) Plus(value accumulatingCounter) {
	c := value.(*multiValueAccumulatingCounter)
	counter.point.Plus(c.point)

	for key, value := range c.values {
		counter.values[key] += value
	}
}

func (counter *multiValueAccumulatingCounter) Minus(value accumulatingCounter) {
	c := value.(*multiValueAccumulatingCounter)
	counter.point.Minus(c.point)

	for key, value := range c.values {
		counter.values[key] -= value
	}
}

// Average accumulating counter. Expect adding and removing CountPoints.
func newAverageAccumulatingCounter(point Point) accumulatingCounter {
	return &averageAccumulatingCounter{
		newSingleValueAccumulatingCounter(point).(*singleValueAccumulatingCounter),
		point.(*CountPoint).Count.(float64),
	}
}

type averageAccumulatingCounter struct {
	point *singleValueAccumulatingCounter
	sum   float64
}

func (counter *averageAccumulatingCounter) Add(point Point) {
	counter.point.Add(point)
	counter.sum += point.(*CountPoint).Count.(float64)
}

func (counter *averageAccumulatingCounter) Remove(point Point) {
	counter.point.Remove(point)
	counter.sum -= point.(*CountPoint).Count.(float64)
}

func (counter *averageAccumulatingCounter) Point() *CountPoint {
	center := counter.point.Point()

	if center == nil {
		return nil
	}

	return &CountPoint{&GeoPoint{"", center.Lat(), center.Lon()}, counter.sum / float64(center.Count.(int))}
}

func (counter *averageAccumulatingCounter) Plus(value accumulatingCounter) {
	c := value.(*averageAccumulatingCounter)
	counter.point.Plus(c.point)
	counter.sum += c.sum
}

func (counter *averageAccumulatingCounter) Minus(value accumulatingCounter) {
	c := value.(*averageAccumulatingCounter)
	counter.point.Minus(c.point)
	counter.sum -= c.sum
}
