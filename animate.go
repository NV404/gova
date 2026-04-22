package gova

import (
	"sync"
	"sync/atomic"
	"time"
)

type EaseCurve int

const (
	EaseLinear EaseCurve = iota
	EaseInOut
	EaseIn
	EaseOut
)

type Animation struct {
	Duration    time.Duration
	Curve       EaseCurve
	AutoReverse bool
	RepeatCount int // -1 = forever
}

const AnimateForever = -1

func Animate(duration time.Duration) Animation {
	return Animation{Duration: duration, Curve: EaseInOut}
}

func (a Animation) WithCurve(c EaseCurve) Animation {
	a.Curve = c
	return a
}

func (a Animation) Reversed() Animation {
	a.AutoReverse = true
	return a
}

func (a Animation) Repeat(n int) Animation {
	a.RepeatCount = n
	return a
}

// UseAnimation returns a runner that can start/stop an animation.
// The tick callback receives progress from 0.0 to 1.0.
func UseAnimation(s *Scope, anim Animation) *AnimationHandle {
	key := callerKey(2)
	return getOrCreateRef(s, "anim:"+key, &AnimationHandle{
		anim: anim,
	}).Get()
}

type AnimationHandle struct {
	anim    Animation
	running atomic.Bool
	stopFn  func()
}

// Start begins the animation with the given tick callback.
// The tick function receives values from 0.0 to 1.0.
func (h *AnimationHandle) Start(tick func(float32)) {
	if h.running.Load() && h.stopFn != nil {
		h.stopFn()
	}
	h.running.Store(true)

	duration := h.anim.Duration
	curve := h.anim.Curve
	autoReverse := h.anim.AutoReverse
	repeatCount := h.anim.RepeatCount

	stopped := make(chan struct{})
	var once sync.Once
	h.stopFn = func() {
		once.Do(func() { close(stopped) })
		h.running.Store(false)
	}

	go func() {
		iterations := repeatCount + 1
		if repeatCount < 0 {
			iterations = -1 // forever
		}

		for i := 0; iterations < 0 || i < iterations; i++ {
			start := time.Now()
			for {
				select {
				case <-stopped:
					return
				default:
				}

				elapsed := time.Since(start)
				if elapsed >= duration {
					tick(applyCurve(1.0, curve))
					break
				}

				progress := float32(elapsed) / float32(duration)
				tick(applyCurve(progress, curve))
				time.Sleep(16 * time.Millisecond) // ~60fps
			}

			if autoReverse {
				start = time.Now()
				for {
					select {
					case <-stopped:
						return
					default:
					}

					elapsed := time.Since(start)
					if elapsed >= duration {
						tick(applyCurve(0.0, curve))
						break
					}

					progress := 1.0 - float32(elapsed)/float32(duration)
					tick(applyCurve(progress, curve))
					time.Sleep(16 * time.Millisecond)
				}
			}
		}
		h.running.Store(false)
	}()
}

func (h *AnimationHandle) Stop() {
	if h.running.Load() && h.stopFn != nil {
		h.stopFn()
	}
}

func (h *AnimationHandle) Running() bool {
	return h.running.Load()
}

func applyCurve(t float32, curve EaseCurve) float32 {
	switch curve {
	case EaseIn:
		return t * t
	case EaseOut:
		return t * (2 - t)
	case EaseInOut:
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	default:
		return t
	}
}
