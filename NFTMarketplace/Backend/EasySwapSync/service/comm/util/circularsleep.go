package util

import "time"

type CircularSleepTime struct {
	value int
	max   int
}

func NewCircularSleepTime(max int) *CircularSleepTime {
	return &CircularSleepTime{1, max}
}

func (c *CircularSleepTime) Inc() {
	c.value = 1 + ((c.value % c.max) % c.max)
}

func (c *CircularSleepTime) Get() int {
	return c.value
}
func (c *CircularSleepTime) Reset() {
	c.value = 1
}

func (c *CircularSleepTime) Sleep() {
	// 获取持续时间
	duration := time.Duration(c.Get()) * time.Second

	// 使用select语句等待duration时间
	select {
	case <-time.After(duration):
		// 当duration时间过去后，调用Inc方法
		c.Inc()
	}
}
