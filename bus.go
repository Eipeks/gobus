package gobus

import (
    "reflect"
    "log"
    "runtime"
)

// Factory method for EventBus objects.
// Creates a new EventBus with a dispatcher un-buffered channel.
func NewEventBus() (*EventBus) {
    bus := &EventBus{
        subscription: make(Subscription),
        dispatcher:   make(chan interface{}),
    }

    go bus.pollerBus()
    return bus
}

// Factory method for EventBus objects.
// Creates a new EventBus with a dispatcher buffered channel.
func NewEventBusBuffered(chanSize int) (*EventBus) {
    bus := &EventBus{
        subscription: make(Subscription),
        dispatcher:   make(chan interface{}, chanSize),
    }

    go bus.pollerBus()
    return bus
}

// Closes the EventBus, waiting for all the goroutines to complete and signals
// the poller goroutine to quit.
// Should be deferred after the factory call:
//
//      func main() {
//          bus := gobus.NewEventBus()
//          defer bus.Destruct()
//          ...
//      }
//
func (bus *EventBus) Destruct() {
    bus.waitGroup.Wait()
    close(bus.dispatcher)
}

// Subscribe a listener to certain events.
// The listener must be an unary function with no return arguments (a.k.a. procedure).
// Uses variadic arguments and chaining methods pattern for great expressiveness.
func (bus *EventBus) Subscribe(listeners ...interface{}) (*EventBus) {
    for _, listener := range listeners {
        bus.subscription.AddListener(listener)
    }
    return bus
}

// Unsubscribe a listener from the event bus.
// Uses variadic arguments and chaining methods pattern for great expressiveness.
func (bus *EventBus) UnSubscribe(listeners ...interface{}) (*EventBus) {
    for _, listener := range listeners {
        bus.subscription.RemoveListener(listener)
    }
    return bus
}

// Publish an event to EventBus.
// The event bus notifies the poller goroutine, which will retrieve the correct subscribed
// listeners and calls them with a copy of the event published.
func (bus *EventBus) Publish(event interface{}) (*EventBus) {
    bus.waitGroup.Add(1)    // Waiting for alerting
    bus.dispatcher <- event // Publishing event into the dispatcher channel
    return bus
}

// Retrieves all the listener subscribed to the event type
// and calls them asynchronously (decorated listeners for waitgroup signal)
func (bus *EventBus) alertListeners(event interface{}) {
    listeners, err := bus.subscription.GetListeners(reflect.TypeOf(event))
    if err == nil {
        for _, listener := range listeners.Values() {
            bus.waitGroup.Add(1)    // Waiting for listener callback
            go bus.executingWithWaiting(listener, event)    // Decorator :-)
        }
    } else {
        log.Print(err)
    }

    // Alerting finished
    bus.waitGroup.Done()
}

// Decorator for listener execution on the event.
// Calls the listener and signals completion on the EventBus waitgroup.
func (bus *EventBus) executingWithWaiting(listener interface{}, event interface{}) {
    funct, evt := reflect.ValueOf(listener), reflect.ValueOf(event)
    funct.Call([]reflect.Value{evt})

    bus.waitGroup.Done()
}

// Bus poller loop, executed asynchronously on bus creation.
// Listens for new event incoming and dispatches them to the listeners-alerting goroutine.
func (bus *EventBus) pollerBus() {
    for {
        select {
        case v, ok := <-bus.dispatcher:
            // No more values added and channel closed, returns
            if !ok {
                return
            }
            // New event received, alerting listeners asynchronously
            go bus.alertListeners(v)

        default:
            // Yields the processor and let other goroutines to execute
            runtime.Gosched()
        }
    }
}
