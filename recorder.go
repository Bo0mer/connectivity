// Package connectivity provides primitives for recording one's internet
// connectivity.
package connectivity

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Probe represents a connectivity probe.
type Probe func(context.Context) error

// Kind represents a span kind.
type Kind int

// Kinds of connectivity spans.
const (
	Offline Kind = iota
	Online
)

// Span describes one's connectivity for a time period.
type Span struct {
	Kind      Kind // Offline or Online span.
	StartTime time.Time
	EndTime   time.Time
}

// Duration returns the duration of the span.
func (s *Span) Duration() time.Duration { return s.EndTime.Sub(s.StartTime) }

// Recorder records one's connectivity.
type Recorder struct {
	mu    sync.Mutex // guards
	spans []Span

	probe Probe

	probeInterval time.Duration
	probeTimeout  time.Duration

	done chan struct{}

	maxSpans int
}

// NewRecorder returns new connectivity recorder.
func NewRecorder(opts ...Option) *Recorder {
	c := &Recorder{
		probe: func(ctx context.Context) error {
			req, err := http.NewRequest("HEAD", "http://connectivitycheck.gstatic.com/generate_204", nil)
			if err != nil {
				return err
			}
			req = req.WithContext(ctx)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("connectivity: unexpected response code: %d", resp.StatusCode)
			}
			return nil
		},
		probeInterval: 2 * time.Second,
		probeTimeout:  500 * time.Millisecond,
		done:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt.configure(c)
	}

	go c.recordLoop()
	return c
}

// Spans returns all recorded connectivity spans.
func (c *Recorder) Spans() []Span {
	c.mu.Lock()
	defer c.mu.Unlock()

	spans := make([]Span, len(c.spans))
	copy(spans, c.spans)
	return spans
}

func (c *Recorder) recordLoop() {
	tick := time.NewTicker(c.probeInterval)
	defer tick.Stop()
	for {
		select {
		case <-c.done:
			close(c.done)
			return
		case <-tick.C:
			c.record()
		}
	}
}

func (c *Recorder) record() {
	kind := Online
	ctx, cancel := context.WithTimeout(context.Background(), c.probeTimeout)
	defer cancel()
	if err := c.probe(ctx); err != nil {
		kind = Offline
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.spans) == 0 || c.spans[len(c.spans)-1].Kind != kind {
		now := time.Now()
		if c.maxSpans > 0 && len(c.spans) == c.maxSpans {
			c.spans = c.spans[1:]
		}
		c.spans = append(c.spans, Span{
			StartTime: now,
			EndTime:   now,
			Kind:      kind,
		})
		return
	}
	c.spans[len(c.spans)-1].EndTime = time.Now()
}

// Stop stops recording.
func (c *Recorder) Stop() {
	c.done <- struct{}{}
	<-c.done
}
