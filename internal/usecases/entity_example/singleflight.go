package entity_example

import "golang.org/x/sync/singleflight"

// FlightGroup deduplicates concurrent requests for the same entity.
// Prevents cache stampede (thundering herd) when many goroutines
// query the same entity during a cache miss.
type FlightGroup struct {
	byID singleflight.Group
}

// NewFlightGroup creates a new FlightGroup.
func NewFlightGroup() *FlightGroup {
	return &FlightGroup{}
}
