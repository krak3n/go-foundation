package foundation

import "sync"

type EventHookFunc func()

type EventHook interface {
	Done(fns ...EventHookFunc)
	Stop(fns ...EventHookFunc)
}

type eventHook uint8

const (
	doneEvent eventHook = iota + 1
	stopEvent
)

type eventHooks struct {
	mtx   sync.RWMutex
	hooks map[eventHook][]EventHookFunc
}

func newEventHooks() *eventHooks {
	return &eventHooks{
		hooks: make(map[eventHook][]EventHookFunc),
	}
}

func (e *eventHooks) Done(fns ...EventHookFunc) {
	e.add(doneEvent, fns...)
}

func (e *eventHooks) Stop(fns ...EventHookFunc) {
	e.add(stopEvent, fns...)
}

func (e *eventHooks) add(event eventHook, fns ...EventHookFunc) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	e.hooks[event] = append(e.hooks[event], fns...)
}

func (e *eventHooks) get(event eventHook) []EventHookFunc {
	e.mtx.RLock()
	defer e.mtx.RUnlock()

	return e.hooks[event]
}
