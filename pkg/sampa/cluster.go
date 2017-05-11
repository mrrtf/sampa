package sampa

// Cluster describes a Sampa cluster, i.e. a set of
// ADC samples.
// Note that we use ints whereas each value really
// is 10 bits (or 20 bits for samples in sum mode)
type Cluster struct {
	hadd    int   // hardwar address
	chadd   int   // channel address
	n       int   // number of samples
	ts      int   // timestamp
	samples []int // samples
}
