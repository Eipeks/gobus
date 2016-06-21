package gobus

import "sync"

type (
    // Map that holds all listener references, indexed through input argument name.
    // Uses an IListenerSet interface as return type, for optimization purposes.
    // Example of subscriptions map:
    //     map
    //     |--> string (built-in)
    //     |    |--> printString1(), printString2()
    //     |
    //     |--> Struct1 (user-defined)
    //          |--> printStruct1(), doSomethingStruct1()
    //
    Subscription map[string]IListenerSet

    // EventBus
    EventBus struct {
        dispatcher   chan interface{}
        quit         chan bool
        subscription Subscription
        waitGroup    sync.WaitGroup
    }

    // ListenerSet is a struct that uses an interface{} slice
    // to implement a listeners set (IListenerSet interface).
    ListenerSet struct {
        listeners []interface{}
    }

    // Interface for listener set.
    IListenerSet interface {
        Add(listener interface{})    IListenerSet
        Remove(listener interface{}) IListenerSet
        Values() []interface{}
        Empty()  bool
    }
)
