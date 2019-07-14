package tcp

import "time"

type Limiter struct {
	Threshold int
	Connects int
}

func NewLimiter(MaxConnects int) Limiter {
	return Limiter{
		Threshold: MaxConnects,
	}
}

func (limiter *Limiter) Fuse()  {
	if limiter.Connects <= limiter.Threshold {
		limiter.Connects++
	} else {
		time.Sleep(time.Second)
	}
}