package connectivity

import "time"

// Option configures a recorder instance.
type Option interface {
	configure(r *Recorder)
}

type withProbeInterval time.Duration

// WithProbeInterval configures the duration between consecutive connectivity
// probes.
func WithProbeInterval(d time.Duration) Option {
	return withProbeInterval(d)
}

func (i withProbeInterval) configure(r *Recorder) {
	r.probeInterval = time.Duration(i)
}

// WithProbeTimeout configures the timeout for individual connectivity probes.
func WithProbeTimeout(d time.Duration) Option {
	return withProbeTimeout(d)
}

func (i withProbeTimeout) configure(r *Recorder) {
	r.probeTimeout = time.Duration(i)
}

type withProbe Probe

// WithProbe configuress the used connectivity probe.
func WithProbe(probe Probe) Option {
	return withProbe(probe)
}

type withProbeTimeout time.Duration

func (f withProbe) configure(r *Recorder) {
	r.probe = Probe(f)
}

type withMaxSpans int

// WithMaxSpans configures the maximum number of spans to retain. If the number
// of recorded spans exceeds n, the oldest span is deleted.
func WithMaxSpans(n int) Option {
	return withMaxSpans(n)
}

func (n withMaxSpans) configure(r *Recorder) {
	r.maxSpans = int(n)
}
