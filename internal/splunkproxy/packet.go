// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code is copied from original work under this license:
// MIT License
//
//Copyright (c) 2019 Junyu Wang
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package splunkproxy

import (
	"fmt"
	"io"
	"strconv"
)

const (
	// OPCODE_REQUEST_INIT is a bit mask that's used to verify if request is the first request
	OPCODE_REQUEST_INIT = 0x01
	// OPCODE_REQUEST_BLOCK is a bit mask that's used to verify if request contains input block
	OPCODE_REQUEST_BLOCK = 0x02
	// OPCODE_REQUEST_END is a bit mask that's used to verify if request is the end
	OPCODE_REQUEST_END = 0x04
	// OPCODE_REQUEST_ALLOW_STREAM is a bit mask that's used to verify if request indicates streaming handler is allowed
	OPCODE_REQUEST_ALLOW_STREAM = 0x08
)

// requestPacket object representing a received packet
type requestPacket struct {
	commandArg string
	block      string
	command    []string
	opcode     rune
}

// if this packet represents the beginning of the request
func (p *requestPacket) isFirst() bool {
	return (p.opcode & OPCODE_REQUEST_INIT) != 0
}

// if this packet represents the end of the request
func (p *requestPacket) isLast() bool {
	return (p.opcode & OPCODE_REQUEST_END) != 0
}

// if this packet contains an input block for the request
func (p *requestPacket) hasBlock() bool {
	return (p.opcode & OPCODE_REQUEST_BLOCK) != 0
}

// if this packet allows stream ???
// TODO: figure out how this is used.
func (p *requestPacket) allowStream() bool {
	return (p.opcode & OPCODE_REQUEST_ALLOW_STREAM) != 0
}

// readPacket creates a packet based on input and communication protocol set by splunkd.
func readPacket(reader io.Reader) (*requestPacket, error) {
	packet := &requestPacket{}
	if err := packet.readOpcode(reader); err != nil {
		return nil, err
	}
	if packet.isFirst() {
		// if packet is the beginning of the request, read command and command args
		if err := packet.readCommandAndArgs(reader); err != nil {
			return nil, err
		}
	}
	if packet.hasBlock() {
		// if packet contains input, read block
		if err := packet.readInputBlock(reader); err != nil {
			return nil, err
		}
	}
	// fmt.Println("Received request packet:", packet)
	return packet, nil
}

// Read opcode from an IO reader and set its value to this packet
// and as a side effect it moves the pointer on the reader to the
// byte after the opcode byte
func (p *requestPacket) readOpcode(reader io.Reader) error {
	for {
		// opcode is the first NON-NEW-LINE byte of the input reader's content
		opbyte := make([]byte, 1, 1)
		_, err := reader.Read(opbyte)
		// if unknown error returend or EOF reached (io.EOF will be returned)
		if err != nil {
			return err
		}
		opbyteStr := string(opbyte)
		if opbyteStr != "\n" { // ignores newlines before opcode
			// NOTE 1: a rune represents an unicode code point, a rune could be equivalent to multiple bytes
			// depending on if the converted is ASCII or unicode, but one byte is at most one rune. (https://yourbasic.org/golang/rune/)
			// NOTE 2: in golang, a string is by default unicode text encoded in UTF-8.
			p.opcode = []rune(opbyteStr)[0]
			break
		}
	}
	return nil
}

// read command and args from input and set its value to this packet.
// As a side-effect the pointer on the input reader will be moved to the byte
// after command and command args.
func (p *requestPacket) readCommandAndArgs(reader io.Reader) error {
	// read number of commands to read from -- protocol
	// <num_of_commands>\n
	numOfCommandPieces, err := readNumber(reader)
	if err != nil {
		return err
	}
	// read commands -- command protocol
	// <command_1_len>\n<command_1>\n<command_2_len>\n<command_2>\n....<command_n_len>\n<command_n>\n
	p.command = make([]string, numOfCommandPieces, numOfCommandPieces)
	for i := 0; i < numOfCommandPieces; i++ {
		command, err := readString(reader)
		if err != nil {
			return err
		}
		p.command[i] = command
	}
	// read command arg -- command arg protocol
	// <command_arg_len>\n<command_arg>\n
	commandArg, err := readString(reader)
	if err != nil {
		return err
	}
	p.commandArg = commandArg
	return nil
}

// read input block from input reader and set its value to this packet.
func (p *requestPacket) readInputBlock(reader io.Reader) error {
	block, err := readString(reader)
	if err != nil {
		return err
	}
	p.block = block
	return nil
}

// readString reads the first line to get the byte length of the info (on the second line)
// then read the actual info from second line based on length from the first line.
// <info_byte_len>\n<info>\n
func readString(reader io.Reader) (string, error) {
	numBytes, err := readNumber(reader)
	if err != nil {
		return "", err
	}
	content := make([]byte, numBytes, numBytes)
	_, err = reader.Read(content)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// readNumber reads the input until it finds a number ending with a newline character or reached EOF
func readNumber(reader io.Reader) (int, error) {
	for {
		content, err := readToEOL(reader)
		if err != nil {
			return 0, err
		}
		if content == "" { // ignore empty lines before a count
			continue
		}
		count, err := strconv.ParseInt(content, 10, 64)
		if err != nil {
			return 0, err
		}
		if count < 0 {
			return -1, fmt.Errorf("expected non-negative integer, got \"%d\"", count)
		}
		return int(count), nil
	}
}

// readToEOL reads the input line by line and moves the pointer to the start of the next line of the input
func readToEOL(reader io.Reader) (string, error) {
	content := make([]byte, 0)
	for {
		buffer := make([]byte, 1, 1)
		_, err := reader.Read(buffer)
		if err != nil {
			return "", err
		}
		if string(buffer) == "\n" {
			break
		}
		content = append(content, buffer[0])
	}
	return string(content), nil
}
