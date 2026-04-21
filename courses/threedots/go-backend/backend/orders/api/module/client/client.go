// Package client defines internal module contracts for the orders module.
// These types are used for synchronous communication within the service between different modules
// and should not be exposed as public API. They provide a stable internal interface
// that decouples service implementations from each other.
package client

type Orders interface {
	PingOrders(PingOrdersRequest) error
}

type PingOrdersRequest struct{}
