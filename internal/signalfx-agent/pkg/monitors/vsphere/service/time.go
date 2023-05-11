package service

import "time"

type TimeSvc struct {
	gateway IGateway
}

func NewTimeSvc(gateway IGateway) *TimeSvc {
	return &TimeSvc{gateway: gateway}
}

func (svc *TimeSvc) RetrieveCurrentTime() (*time.Time, error) {
	return svc.gateway.retrieveCurrentTime()
}
