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
	done chan<- bool
}

// The object.
type ChanPubSub struct {
	control chan cmdT
	subs    []chan interface{}
}

// Create a new ChanPubSub object and call `Start()` on it.
func New() *ChanPubSub {
	return (&ChanPubSub{}).Start()
}

// Starts the control goroutine that listens for published messages and
// distributes them to each of the subscribers.
func (this *ChanPubSub) Start() *ChanPubSub {
	if this.control != nil {
		return this
	}

	this.control = make(chan cmdT, 128)

	go func() {
		var stopCmd *cmdT

	loop:
		for {
			select {
			case cmd := <-this.control:
				switch cmd.op {
				case sub:
					this.subs = append(this.subs, cmd.ch)
					cmd.done <- true
				case unsub:
					for i, test := range this.subs {
						if test == cmd.rch {
							this.subs = append(this.subs[:i], this.subs[i+1:]...)
							break
						}
					}
					cmd.done <- true
				case pub:
					for _, ch := range this.subs {
						ch <- cmd.msg
					}
					cmd.done <- true
				case size:
					cmd.ch <- len(this.subs)
				case stop:
					stopCmd = &cmd
					break loop
				}
			}
		}

		this.control = nil
		stopCmd.done <- true
	}()

	return this
}

// Subscribe to all messages.
// Returns the chan to receive messages.
// Also, returns a chan to notify when the command has completed.
func (this *ChanPubSub) Sub() (<-chan interface{}, <-chan bool) {
	ch := make(chan interface{}, 128)
	done := make(chan bool, 1)

	this.control <- cmdT{
		op:   sub,
		ch:   ch,
		done: done,
	}

	return ch, done
}

// Unsubscribe a chan.
// Often used with `defer`
// Also, returns a chan to notify when the command has completed.
func (this *ChanPubSub) UnSub(ch <-chan interface{}) <-chan bool {
	done := make(chan bool, 1)

	this.control <- cmdT{
		op:   unsub,
		rch:  ch,
		done: done,
	}

	return done
}

// Publish a message to all subscribers.
// Returns a chan to notify when the command has completed.
func (this *ChanPubSub) Pub(msg interface{}) <-chan bool {
	done := make(chan bool, 1)

	this.control <- cmdT{
		op:   pub,
		msg:  msg,
		done: done,
	}

	return done
}

// Returns the number of subscribers. This command is synchronous.
func (this *ChanPubSub) Size() int {
	resp := make(chan interface{})

	this.control <- cmdT{
		op: size,
		ch: resp,
	}

	return (<-resp).(int)
}

// Stops the control goroutine that listens for published messages and
// distributes them to each of the subscribers.
func (this *ChanPubSub) Stop() <-chan bool {
	done := make(chan bool, 1)

	this.control <- cmdT{
		op:   stop,
		done: done,
	}

	return done
}
