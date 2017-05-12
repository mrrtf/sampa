package sampa

import "fmt"

// Packet describes a Sampa packet, i.e. a set of
// clusters (sets of ADC samples).
// Note that we use ints whereas each value really
// is 10 bits (or 20 bits for samples in sum mode)
type Packet struct {
	sdh      SampaDataHeader
	clusters []Cluster // clusters
}

func (p *Packet) AddCluster(timestamp int, samples []int) {
	p.clusters = append(p.clusters, Cluster{ts: timestamp, samples: samples})
}

func (p *Packet) String() string {
	v := fmt.Sprintf("[%d,%d] ", p.sdh.Hadd(), p.sdh.CHadd())

	for _, c := range p.clusters {
		m, s := c.MeanSigma()
		v += fmt.Sprintf("%f +- %f ", m, s)
		v += " | "
	}
	return v
}
