package connectivity_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bo0mer/connectivity"
	"github.com/Bo0mer/connectivitycheck"
)

func TestRecorderSimple(t *testing.T) {
	probe := func(_ context.Context) error {
		return nil
	}

	r := connectivity.NewRecorder(
		connectivity.WithProbe(probe),
		connectivity.WithProbeInterval(time.Millisecond),
	)
	defer r.Stop()

	time.Sleep(time.Millisecond * 20)
	spans := r.Spans()
	if len(spans) != 1 {
		t.Errorf("want 1 span, got: %d", len(spans))
	}
}

func TestRecorderAlternatingProbes(t *testing.T) {
	var nProbes int
	var err error
	probe := func(_ context.Context) error {
		nProbes++
		if err != nil {
			err = nil
		} else {
			err = errors.New("i/o timeout")
		}
		return err
	}

	r := connectivity.NewRecorder(
		connectivity.WithProbe(probe),
		connectivity.WithProbeInterval(time.Millisecond),
	)

	time.Sleep(time.Millisecond * 20)
	r.Stop()

	spans := r.Spans()
	if len(spans) != nProbes {
		t.Errorf("want %d span, got: %d", nProbes, len(spans))
	}
	if spans[0].Kind != connectivity.Offline {
		t.Errorf("wrong initial span Kind -- want %d, got: %d", connectivitycheck.Offline, spans[0].Kind)
	}
	for i := 0; i < len(spans)-1; i += 2 {
		if spans[i].Kind == spans[i+1].Kind {
			t.Errorf("spans %d and %d have same kind, they shouldn't", i, i+1)
		}
	}
}
