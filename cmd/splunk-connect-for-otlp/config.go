// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultGrpcPort      = 4317
	DefaultHTTPPort      = 4318
	DefaultListenAddress = "0.0.0.0"
)

type XMLInput struct {
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
	ListenAddress string
	Source        string
	Sourcetype    string
	ServerCert    string
	ServerKey     string
	ServerURI     string
	SessionKey    string
	GrpcPort      int
	HTTPPort      int
	EnableSSL     bool
}

func (x XMLInput) Extract() InputConfig {
	grpcPort := DefaultGrpcPort
	httpPort := DefaultHTTPPort
	listeningAddress := DefaultListenAddress
	source := ""
	sourcetype := ""
	enableSSL := true
	serverCert := ""
	serverKey := ""

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
		}
	}

	return InputConfig{
		GrpcPort:      grpcPort,
		HTTPPort:      httpPort,
		ListenAddress: listeningAddress,
		Source:        source,
		Sourcetype:    sourcetype,
		ServerURI:     serverURI,
		SessionKey:    sessionKey,
		EnableSSL:     enableSSL,
		ServerCert:    serverCert,
		ServerKey:     serverKey,
	}
}

func (c InputConfig) Validate() error {
	if c.EnableSSL && (c.ServerCert == "" || c.ServerKey == "") {
		return errors.New("enableSSL requires both serverCert and serverKey to be set")
	}
	return nil
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
