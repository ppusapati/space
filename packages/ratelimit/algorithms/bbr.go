package algorithms

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"p9e.in/samavaya/packages/ratelimit"
)

// BBRLimiter implements Google's BBR (Bottleneck Bandwidth and RTT) congestion control
type BBRLimiter struct {
	state              ratelimit.BBRState
	bandwidth          float64 // packets per second
	minRTT             time.Duration
	rtt                time.Duration
	cwnd               int64 // congestion window (in-flight requests)
	inflight           int64 // current in-flight count
	bdp                int64 // bandwidth delay product
	mu                 sync.RWMutex

	// Tracking
	rttSamples         []time.Duration
	bandwidthSamples   []float64
	probeRTTTimer      *time.Timer
	deliveredTime      int64 // nanoseconds
	deliveredPackets   int64
	cycleCount         int64
	roundCount         int64

	// Configuration
	startupGrowthTarget float64 // CWND growth target (default 1.25)
	pacing              float64 // Pacing rate multiplier
}

// NewBBRLimiter creates a new BBR rate limiter
func NewBBRLimiter() *BBRLimiter {
	return &BBRLimiter{
		state:               ratelimit.BBRStartup,
		bandwidth:           1000.0, // Start with 1000 req/s estimate
		minRTT:              math.MaxInt64 * time.Nanosecond,
		rtt:                 10 * time.Millisecond,
		cwnd:                100, // Initial congestion window
		inflight:            0,
		rttSamples:          make([]time.Duration, 0, 10),
		bandwidthSamples:    make([]float64, 0, 10),
		startupGrowthTarget: 1.25,
		pacing:              1.0,
	}
}

// Allow checks if a request is allowed (no pending requests)
func (bbr *BBRLimiter) Allow(ctx context.Context, key string) (bool, error) {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	// Check if we can admit a new request
	if bbr.inflight < bbr.cwnd {
		atomic.AddInt64(&bbr.inflight, 1)
		return true, nil
	}

	return false, nil
}

// AllowN allows N requests
func (bbr *BBRLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	if bbr.inflight+int64(n) <= bbr.cwnd {
		atomic.AddInt64(&bbr.inflight, int64(n))
		return true, nil
	}

	return false, nil
}

// Reserve reserves capacity for a future request
func (bbr *BBRLimiter) Reserve(ctx context.Context, key string) (*ratelimit.Reservation, error) {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	if bbr.inflight < bbr.cwnd {
		atomic.AddInt64(&bbr.inflight, 1)
		return &ratelimit.Reservation{
			ReadyAt: time.Now(),
			Delay:   0,
			OK:       true,
		}, nil
	}

	// Calculate delay based on pacing rate
	pacingRate := bbr.bandwidth / bbr.pacing
	delay := time.Duration(float64(time.Second) / pacingRate)

	return &ratelimit.Reservation{
		ReadyAt: time.Now().Add(delay),
		Delay:   delay,
		OK:       false,
	}, nil
}

// RecordRTT records a round-trip time measurement
func (bbr *BBRLimiter) RecordRTT(latency time.Duration) {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	// Update RTT samples
	bbr.rttSamples = append(bbr.rttSamples, latency)
	if len(bbr.rttSamples) > 10 {
		bbr.rttSamples = bbr.rttSamples[1:]
	}

	// Update min RTT
	if latency < bbr.minRTT {
		bbr.minRTT = latency
	}

	// Exponential weighted moving average
	if bbr.rtt == 0 {
		bbr.rtt = latency
	} else {
		bbr.rtt = time.Duration(float64(bbr.rtt)*0.8 + float64(latency)*0.2)
	}

	bbr.roundCount++
}

// RecordDelivery records packet delivery for bandwidth estimation
func (bbr *BBRLimiter) RecordDelivery(packets int, latency time.Duration) {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	// Estimate delivery rate (packets per second)
	if latency > 0 {
		rate := float64(packets) / latency.Seconds()
		bbr.bandwidthSamples = append(bbr.bandwidthSamples, rate)
		if len(bbr.bandwidthSamples) > 10 {
			bbr.bandwidthSamples = bbr.bandwidthSamples[1:]
		}

		// Keep max of samples
		if rate > bbr.bandwidth {
			bbr.bandwidth = rate
		}
	}

	bbr.deliveredPackets += int64(packets)
	atomic.AddInt64(&bbr.inflight, -int64(packets))

	// Update state machine
	bbr.updateState()
}

// GetStats returns current rate limit stats
func (bbr *BBRLimiter) GetStats(ctx context.Context, key string) (*ratelimit.Stats, error) {
	bbr.mu.RLock()
	defer bbr.mu.RUnlock()

	return &ratelimit.Stats{
		Key:          key,
		AllowedCount: bbr.cwnd,
		CurrentLimit: int64(bbr.bandwidth),
		Metrics: map[string]interface{}{
			"state":           bbr.state,
			"bandwidth":       bbr.bandwidth,
			"cwnd":            bbr.cwnd,
			"inflight":        bbr.inflight,
			"rtt_ms":          bbr.rtt.Milliseconds(),
			"min_rtt_ms":      bbr.minRTT.Milliseconds(),
			"bdp":             bbr.bdp,
			"pacing_rate":     bbr.pacing,
		},
	}, nil
}

// Reset resets the BBR limiter state
func (bbr *BBRLimiter) Reset(ctx context.Context, key string) error {
	bbr.mu.Lock()
	defer bbr.mu.Unlock()

	bbr.state = ratelimit.BBRStartup
	bbr.bandwidth = 1000.0
	bbr.minRTT = math.MaxInt64 * time.Nanosecond
	bbr.rtt = 10 * time.Millisecond
	bbr.cwnd = 100
	bbr.inflight = 0
	bbr.rttSamples = make([]time.Duration, 0, 10)
	bbr.bandwidthSamples = make([]float64, 0, 10)

	return nil
}

// Helper methods

// updateState implements BBR state machine
func (bbr *BBRLimiter) updateState() {
	switch bbr.state {
	case ratelimit.BBRStartup:
		// In startup: increase CWND aggressively
		// Stay in startup while CWND growth > 1.25
		if bbr.cwnd > int64(float64(bbr.bdp)*bbr.startupGrowthTarget) {
			bbr.state = ratelimit.BBRDrain
		} else {
			bbr.cwnd = int64(float64(bbr.cwnd) * 2.0) // Double CWND
		}

	case ratelimit.BBRDrain:
		// In drain: reduce CWND to BDP
		if bbr.cwnd <= bbr.bdp {
			bbr.state = ratelimit.BBRProbeBW
			bbr.cycleCount = 0
		} else {
			bbr.cwnd = bbr.cwnd * 2 / 3 // Reduce by 33%
		}

	case ratelimit.BBRProbeBW:
		// In probe bandwidth: cycle through bandwidth probes
		// Every 8 RTTs, try to increase bandwidth
		probeInterval := int64(8)
		if bbr.roundCount%probeInterval == 0 {
			bbr.state = ratelimit.BBRProbeRTT
		}

	case ratelimit.BBRProbeRTT:
		// In probe RTT: measure minimum RTT, use minimal CWND
		minCWND := int64(4)
		if bbr.cwnd > minCWND {
			bbr.cwnd = minCWND
		}

		// After 200ms of probing, return to ProbeBW
		if bbr.roundCount%10 == 0 {
			bbr.state = ratelimit.BBRProbeBW
		}
	}

	// Update BDP = bandwidth * RTT
	rttSeconds := bbr.rtt.Seconds()
	bbr.bdp = int64(bbr.bandwidth * rttSeconds)

	// Ensure minimum CWND
	if bbr.cwnd < 4 {
		bbr.cwnd = 4
	}
}
