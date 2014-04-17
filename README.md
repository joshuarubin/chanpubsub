# chanpubsub

Provides a simple to use PubSub implementation using channels

## Install

`go get -v -u -t github.com/joshuarubin/chanpubsub`

## package chanpubsub
`import "github.com/joshuarubin/chanpubsub"`


### Example:
```go
	p := New()                           // create and start the broker
	ch, done := p.Sub()                  // add a subscriber
	defer p.UnSub(ch)                    // ensure the subscriber is unsubscribed as necessary
	<-done                               // wait for the subscription to fully register
	p.Pub("test message")                // publish a message asynchronously
	fmt.Printf("received: \"%s\"", <-ch) // receive the message
	// Output: received: "test message"
```

### TYPES

```go
type ChanPubSub struct {
    // contains filtered or unexported fields
}
```

The object.

---

```go
func New() *ChanPubSub
```

Create a new ChanPubSub object and call `Start()` on it.

---

```go
func (this *ChanPubSub) Pub(msg interface{}) <-chan bool
```

Publish a message to all subscribers. Returns a chan to notify when the command has completed.

---

```go
func (this *ChanPubSub) Size() int
```

Returns the number of subscribers. This command is synchronous.

---

```go
func (this *ChanPubSub) Start() *ChanPubSub
```

Starts the control goroutine that listens for published messages and distributes them to each of the subscribers.

---

```go
func (this *ChanPubSub) Stop() <-chan bool
```

Stops the control goroutine that listens for published messages and distributes them to each of the subscribers.

---

```go
func (this *ChanPubSub) Sub() (<-chan interface{}, <-chan bool)
```

Subscribe to all messages. Returns the chan to receive messages. Also, returns a chan to notify when the command has completed.

---

```go
func (this *ChanPubSub) UnSub(ch <-chan interface{}) <-chan bool

```

Unsubscribe a chan. Often used with `defer` Also, returns a chan to notify when the command has completed.
