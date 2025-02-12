package clock

import "time"

type Clock interface {
	Now() time.Time
}

func WithTime(t time.Time) func(*ClockImpl) {
	return func(ci *ClockImpl) {
		ci.now = &t
	}
}

func New(opts ...func(*ClockImpl)) Clock {
	clock := &ClockImpl{
		now: nil,
	}

	for _, opt := range opts {
		opt(clock)
	}

	return clock
}

type ClockImpl struct {
	now *time.Time
}

func (c *ClockImpl) Now() time.Time {
	if c.now != nil {
		return *c.now
	}
	return time.Now()
}
