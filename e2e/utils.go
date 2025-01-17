package e2e

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kmsg"
	"go.uber.org/zap"
)

// createHistogramBuckets creates the buckets for the histogram based on the number of desired buckets (10) and the
// upper bucket size.
func createHistogramBuckets(maxLatency time.Duration) []float64 {
	// Since this is an exponential bucket we need to take Log base2 or binary as the upper bound
	// Divide by 10 for the argument because the base is counted as 20ms and we want to normalize it as base 2 instead of 20
	// +2 because it starts at 5ms or 0.005 sec, to account 5ms and 10ms before it goes to the base which in this case is 0.02 sec or 20ms
	// and another +1 to account for decimal points on int parsing
	latencyCount := math.Logb(float64(maxLatency.Milliseconds() / 10))
	count := int(latencyCount) + 3
	bucket := prometheus.ExponentialBuckets(0.005, 2, count)

	return bucket
}

func containsStr(ar []string, x string) (bool, int) {
	for i, item := range ar {
		if item == x {
			return true, i
		}
	}
	return false, -1
}

// logs all errors, returns number of errors
func (s *Service) logCommitErrors(r *kmsg.OffsetCommitResponse, err error) int {
	if err != nil {
		s.logger.Error("offset commit failed", zap.Error(err))
		return 1
	}

	errCount := 0
	for _, t := range r.Topics {
		for _, p := range t.Partitions {
			err := kerr.ErrorForCode(p.ErrorCode)
			if err != nil {
				s.logger.Error("error committing partition offset",
					zap.String("topic", t.Topic),
					zap.Int32("partition_id", p.Partition),
					zap.Error(err),
				)
				errCount++
			}
		}
	}

	return errCount
}

func safeUnwrap(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
