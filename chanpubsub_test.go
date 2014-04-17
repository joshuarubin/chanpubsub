package chanpubsub

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubSub(t *testing.T) {
	Convey("ChanPubSub should be created properly", t, func() {
		var p *ChanPubSub
		So(p, ShouldEqual, nil)
		p = New()
		So(p, ShouldNotEqual, nil)
		So(p.control, ShouldNotEqual, nil)
		So(p.Size(), ShouldEqual, 0)
	})

	Convey("ChanPubSub should be able to be stopped", t, func() {
		p := New()
		So(p.control, ShouldNotEqual, nil)
		<-p.Stop()
		So(p.control, ShouldEqual, nil)
	})

	Convey("Subscribers should affect ChanPubSub size", t, func() {
		p := New()
		ch, done := p.Sub()
		<-done // wait for it...
		So(ch, ShouldNotEqual, nil)
		So(p.Size(), ShouldEqual, 1)
		<-p.UnSub(ch) // wait for it...
		So(p.Size(), ShouldEqual, 0)
	})

	Convey("Published messages should reach all subscribers", t, func() {
		const numSubs = 5
		p := New()
		done := make(chan bool)

		for i := 0; i < numSubs; i++ {
			recv, sync := p.Sub()
			<-sync // wait for it
			go func() {
				val := <-recv
				str, ok := val.(string)
				So(ok, ShouldEqual, true)
				So(str, ShouldEqual, "test")
				done <- true
			}()
		}

		p.Pub("test")

		for i := 0; i < numSubs; i++ {
			<-done
		}
	})
}

func Example() {
	p := New()                           // create and start the broker
	ch, done := p.Sub()                  // add a subscriber
	defer p.UnSub(ch)                    // ensure the subscriber is unsubscribed as necessary
	<-done                               // wait for the subscription to fully register
	p.Pub("test message")                // publish a message asynchronously
	fmt.Printf("received: \"%s\"", <-ch) // receive the message
	// Output: received: "test message"
}
