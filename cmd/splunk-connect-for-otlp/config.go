// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension"
)

const (
	DefaultGrpcPort      = 4317
	DefaultHTTPPort      = 4318
	DefaultListenAddress = "0.0.0.0"
)

type XMLInput struct {
	ServerHost    string    `xml:"server_host"`
	ServerURI     string    `xml:"server_uri"`
	SessionKey    string    `xml:"session_key"`
	Configuration XMLConfig `xml:"configuration"`
}

type XMLConfig struct {
	Stanza XMLStanza `xml:"stanza"`
}

type XMLStanza struct {
	Name   string     `xml:"name,attr"`
	App    string     `xml:"app,attr"`
	Params []XMLParam `xml:"param"`
}

type XMLParam struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

type InputConfig struct {
	AuthConfig    *splunkauthextension.Config
	ListenAddress string
	Source        string
	Sourcetype    string
	ServerCert    string
	ServerKey     string
	ServerHost    string
	ServerURI     string
	SessionKey    string
	Realm         string
	AccessToken   string
	GrpcPort      int
	HTTPPort      int
	EnableSSL     bool
}

func (x XMLInput) Extract() (InputConfig, error) {
	authConfig := splunkauthextension.NewDefaultConfig()
	grpcPort := DefaultGrpcPort
	httpPort := DefaultHTTPPort
	listeningAddress := DefaultListenAddress
	source := ""
	sourcetype := ""
	enableSSL := true
	serverCert := ""
	serverKey := ""
	realm := ""
	accessToken := ""

	serverHost := x.ServerHost
	serverURI := x.ServerURI
	sessionKey := x.SessionKey

	for _, p := range x.Configuration.Stanza.Params {
		switch p.Name {
		case "grpc_port":
			grpcPort, _ = strconv.Atoi(p.Value)
		case "http_port":
			httpPort, _ = strconv.Atoi(p.Value)
		case "listen_address":
			listeningAddress = p.Value
		case "source":
			source = p.Value
		case "sourcetype":
			sourcetype = p.Value
		case "enableSSL":
			enableSSL = convertBool(p.Value)
		case "serverCert":
			serverCert = p.Value
		case "serverKey":
			serverKey = p.Value
		case "realm":
			realm = p.Value
		case "access_token":
			accessToken = p.Value
		case "refresh_interval", "refresh_timeout", "max_staleness":
			duration, err := time.ParseDuration(strings.TrimSpace(p.Value))
			if err != nil {
				return InputConfig{}, fmt.Errorf("invalid %s: %w", p.Name, err)
			}
			switch p.Name {
			case "refresh_interval":
				authConfig.RefreshInterval = duration
			case "refresh_timeout":
				authConfig.RefreshTimeout = duration
			case "max_staleness":
				authConfig.MaxStaleness = duration
			}
		}
	}

	return InputConfig{
		AuthConfig:    authConfig,
		GrpcPort:      grpcPort,
		HTTPPort:      httpPort,
		ListenAddress: listeningAddress,
		Source:        source,
		Sourcetype:    sourcetype,
		ServerHost:    serverHost,
		ServerURI:     serverURI,
		SessionKey:    sessionKey,
		EnableSSL:     enableSSL,
		ServerCert:    serverCert,
		ServerKey:     serverKey,
		Realm:         realm,
		AccessToken:   accessToken,
	}, nil
}

func (c InputConfig) Validate() error {
	if c.EnableSSL && (c.ServerCert == "" || c.ServerKey == "") {
		return errors.New("enableSSL requires both serverCert and serverKey to be set")
	}
	if c.AuthConfig == nil {
		return errors.New("authentication configuration is required")
	}
	return c.AuthConfig.Validate()
}

func convertBool(value string) bool {
	switch strings.ToLower(value) {
	case "true":
		return true
	case "1":
		return true
	default:
		return false
	}
}

func ReadFromStdin() (XMLInput, error) {
	scanner := bufio.NewScanner(os.Stdin)
	b := strings.Builder{}
	for scanner.Scan() {
		b.WriteString(scanner.Text())
	}

	var config XMLInput
	err := xml.Unmarshal([]byte(b.String()), &config)
	return config, err
}
