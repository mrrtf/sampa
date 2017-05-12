package sampa

import (
	"fmt"
	"strconv"
)

// Cluster describes a Sampa cluster, i.e. a set of
// ADC samples.
// Note that we use ints whereas each value really
// is 10 bits (or 20 bits for samples in sum mode)
type Cluster struct {
	ts      int   // timestamp
	samples []int // samples
}

func (c *Cluster) String() string {
	v := fmt.Sprintf("(%d) [%d]", c.ts, len(c.samples))
	for _, s := range c.samples {
		v += strconv.FormatInt((int64)(s), 10)
		v += " "
	}
	return v
}

func (c *Cluster) MeanSigma() (float64, float64) {
	var m, s float64 = 0, 0
	for _, x := range c.samples {
		m += float64(x)
	}
	m /= float64(len(c.samples))
	for _, x := range c.samples {
		d := float64(x) - m
		s += d * d
	}
	s /= float64(len(c.samples) - 1)
	return m, s
}
