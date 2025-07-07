package tick

import (
	"sync"

	"go.krak3n.io/foundation"
)

type eventHooks struct {
	f        foundation.F
	doneOnce sync.Once
	stopOnce sync.Once
}

func newEventHooks(f foundation.F) *eventHooks {
	return &eventHooks{
		f: f,
	}
}

func (e *eventHooks) Done(fns ...foundation.EventHookFunc) {
	e.stopOnce.Do(func() {
		e.f.On().Done(fns...)
	})
}

func (e *eventHooks) Stop(fns ...foundation.EventHookFunc) {
	e.stopOnce.Do(func() {
		e.f.On().Stop(fns...)
	})
}
