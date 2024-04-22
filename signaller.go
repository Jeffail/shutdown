package shutdown

import (
	"context"
	"sync"
)

// Signaller is a mechanism owned by components that support graceful
// shut down and is used as a way to signal from outside that any goroutines
// owned by the component should begin to close.
//
// Shutting down can happen in two tiers of urgency, the first is a "soft stop",
// meaning if you're in the middle of something it's okay to do that first
// before terminating, but please do not commit to new work.
//
// The second tier is a "hard stop", where you need to clean up resources and
// terminate as soon as possible, regardless of any tasks that you are currently
// attempting to finish.
//
// Finally, there is also a signal of having stopped, which is made by the
// component and can be used from outside to determine whether the component
// has finished terminating.
type Signaller struct {
	softStopChan chan struct{}
	softStopOnce sync.Once

	hardStopChan chan struct{}
	hardStopOnce sync.Once

	hasStoppedChan chan struct{}
	hasStoppedOnce sync.Once
}

// NewSignaller creates a new signaller.
func NewSignaller() *Signaller {
	return &Signaller{
		softStopChan:   make(chan struct{}),
		hardStopChan:   make(chan struct{}),
		hasStoppedChan: make(chan struct{}),
	}
}

// TriggerSoftStop signals to the owner of this Signaller that it should
// terminate at its own leisure, meaning it's okay to complete any tasks that
// are in progress but no new work should be started.
func (s *Signaller) TriggerSoftStop() {
	s.softStopOnce.Do(func() {
		close(s.softStopChan)
	})
}

// TriggerHardStop signals to the owner of this Signaller that it should
// terminate right now regardless of any in progress tasks.
func (s *Signaller) TriggerHardStop() {
	s.TriggerSoftStop()
	s.hardStopOnce.Do(func() {
		close(s.hardStopChan)
	})
}

// TriggerHasStopped is a signal made by the component that it and all of its
// owned resources have terminated.
func (s *Signaller) TriggerHasStopped() {
	s.hasStoppedOnce.Do(func() {
		close(s.hasStoppedChan)
	})
}

//------------------------------------------------------------------------------

// IsSoftStopSignalled returns true if the signaller has received the signal to
// soft stop.
func (s *Signaller) IsSoftStopSignalled() bool {
	select {
	case <-s.SoftStopChan():
		return true
	default:
	}
	return false
}

// SoftStopChan returns a channel that will be closed when the signal to soft or
// hard stop has been made.
func (s *Signaller) SoftStopChan() <-chan struct{} {
	return s.softStopChan
}

// SoftStopCtx returns a context.Context that will be terminated when either the
// provided context is cancelled or the signal to soft or hard stop has been
// made.
func (s *Signaller) SoftStopCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-s.softStopChan:
		}
		cancel()
	}()
	return ctx, cancel
}

// IsHardStopSignalled returns true if the signaller has received the signal to
// hard stop.
func (s *Signaller) IsHardStopSignalled() bool {
	select {
	case <-s.HardStopChan():
		return true
	default:
	}
	return false
}

// HardStopChan returns a channel that will be closed when the signal to hard
// stop has been made.
func (s *Signaller) HardStopChan() <-chan struct{} {
	return s.hardStopChan
}

// HardStopCtx returns a context.Context that will be terminated when either the
// provided context is cancelled or the signal to hard stop has been made.
func (s *Signaller) HardStopCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-s.hardStopChan:
		}
		cancel()
	}()
	return ctx, cancel
}

// IsHasStoppedSignalled returns true if the signaller has received the signal
// that the component has stopped.
func (s *Signaller) IsHasStoppedSignalled() bool {
	select {
	case <-s.HasStoppedChan():
		return true
	default:
	}
	return false
}

// HasStoppedChan returns a channel that will be closed when the signal that the
// component has stopped has been made.
func (s *Signaller) HasStoppedChan() <-chan struct{} {
	return s.hasStoppedChan
}

// HasStoppedCtx returns a context.Context that will be cancelled when either
// the provided context is cancelled or the signal that the component has
// stopped has been made.
func (s *Signaller) HasStoppedCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-s.hasStoppedChan:
		}
		cancel()
	}()
	return ctx, cancel
}
