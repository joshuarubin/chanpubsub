// Package chanpubsub provides a simple to use PubSub implementation using channels
package chanpubsub

type operation int

const (
	_ = iota
	sub
	unsub
	pub
	size
	stop
)

type cmdT struct {
	op   operation
	ch   chan interface{}
	rch  <-chan interface{}
	msg  interface{}
	done chan<- struct{}
}

// ChanPubSub is an object encapsulating publish-subscribe behavior
type ChanPubSub struct {
	control chan cmdT
	subs    []chan interface{}
}

// New creates a new ChanPubSub object and calls `Start()` on it.
func New() *ChanPubSub {
	return (&ChanPubSub{}).Start()
}

// Start starts the control goroutine that listens for published messages and
// distributes them to each of the subscribers.
func (ps *ChanPubSub) Start() *ChanPubSub {
	if ps.control != nil {
		return ps
	}

	ps.control = make(chan cmdT, 128)

	go func() {
		var stopCmd *cmdT

	loop:
		for {
			select {
			case cmd := <-ps.control:
				switch cmd.op {
				case sub:
					ps.subs = append(ps.subs, cmd.ch)
					cmd.done <- struct{}{}
				case unsub:
					for i, test := range ps.subs {
						if test == cmd.rch {
							ps.subs = append(ps.subs[:i], ps.subs[i+1:]...)
							break
						}
					}
					cmd.done <- struct{}{}
				case pub:
					for _, ch := range ps.subs {
						ch <- cmd.msg
					}
					cmd.done <- struct{}{}
				case size:
					cmd.ch <- len(ps.subs)
				case stop:
					stopCmd = &cmd
					break loop
				}
			}
		}

		ps.control = nil
		stopCmd.done <- struct{}{}
	}()

	return ps
}

// Sub returns a chan that is subscribed to all messages.
// Also, returns a chan to notify when the command has completed.
func (ps *ChanPubSub) Sub() (<-chan interface{}, <-chan struct{}) {
	ch := make(chan interface{}, 128)
	done := make(chan struct{}, 1)

	ps.control <- cmdT{
		op:   sub,
		ch:   ch,
		done: done,
	}

	return ch, done
}

// UnSub unsubscribes a chan from all messages.
// Often used with `defer`
// Also, returns a chan to notify when the command has completed.
func (ps *ChanPubSub) UnSub(ch <-chan interface{}) <-chan struct{} {
	done := make(chan struct{}, 1)

	ps.control <- cmdT{
		op:   unsub,
		rch:  ch,
		done: done,
	}

	return done
}

// Pub publishes a message to all subscribers.
// Returns a chan to notify when the command has completed.
func (ps *ChanPubSub) Pub(msg interface{}) <-chan struct{} {
	done := make(chan struct{}, 1)

	ps.control <- cmdT{
		op:   pub,
		msg:  msg,
		done: done,
	}

	return done
}

// Size returns the number of subscribers. This command is synchronous.
func (ps *ChanPubSub) Size() int {
	resp := make(chan interface{})

	ps.control <- cmdT{
		op: size,
		ch: resp,
	}

	return (<-resp).(int)
}

// Stop stops the control goroutine that listens for published messages and
// distributes them to each of the subscribers.
func (ps *ChanPubSub) Stop() <-chan struct{} {
	done := make(chan struct{}, 1)

	ps.control <- cmdT{
		op:   stop,
		done: done,
	}

	return done
}
