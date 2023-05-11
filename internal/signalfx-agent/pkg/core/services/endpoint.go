// Package services has service endpoint types.  An endpoint is a single port
// on a single instance of a service/application.  The two most important
// attributes of an endpoint are the host and port.  Host can be either an IP
// address or a DNS name.  Endpoints are created by observers.
//
// Most of the core logic for endpoints is on the EndpointCore type, which all
// endpoints must embed.
//
// There is the notion of a "self-configured" endpoint, which means that
// it specifies what monitor type to use to monitor it as well as the
// configuration for that monitor.
package services

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// ID uniquely identifies a service instance
type ID string

// Endpoint is the generic interface that all types of service instances should
// implement.  All consumers of services should use this interface only.
type Endpoint interface {
	config.CustomConfigurable

	// Core returns the EndpointCore that all endpoints are required to have
	Core() *EndpointCore

	// Dimensions that are specific to this endpoint (e.g. container name)
	Dimensions() map[string]string
	// AddDimension adds a single dimension to the endpoint
	AddDimension(string, string)
	// RemoveDimension removes a single dimension from the endpoint
	RemoveDimension(string)
}

// HasDerivedFields is an interface with a single method that can be called to
// get fields that are derived from a service.  This is useful for things like
// aliased fields or computed fields.
type HasDerivedFields interface {
	DerivedFields() map[string]interface{}
}

// EndpointAsMap converts an endpoint to a map that contains all of the
// information about the endpoint.  This makes it easy to use endpoints in
// evaluating rules as well as in collectd templates.
func EndpointAsMap(endpoint Endpoint) map[string]interface{} {
	asMap, err := utils.ConvertToMapViaYAML(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"endpoint": spew.Sdump(endpoint),
		}).Error("Could not convert endpoint to map")
		return nil
	}

	if asMap == nil {
		return nil
	}

	if df, ok := endpoint.(HasDerivedFields); ok {
		return utils.MergeInterfaceMaps(asMap, df.DerivedFields())
	}
	return asMap
}
