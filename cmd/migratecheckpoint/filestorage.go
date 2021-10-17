// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"time"

	"go.etcd.io/bbolt"
)

// This code is from opentelemetry-collector-contrib/extension/storage/filestorage
type opType int

const (
	Get opType = iota
	Set
	Delete
)

type operation struct {
	// Key specifies key which is going to be get/set/deleted
	Key string
	// Value specifies value that is going to be set or holds result of get operation
	Value []byte
	// Type describes the operation type
	Type opType
}

type Operation *operation

func SetOperation(key string, value []byte) Operation {
	return &operation{
		Key:   key,
		Value: value,
		Type:  Set,
	}
}

var defaultBucket = []byte(`default`)

type fileStorageClient struct {
	db *bbolt.DB
}

func newClient(filePath string, timeout time.Duration) (*fileStorageClient, error) {
	options := &bbolt.Options{
		Timeout: timeout,
		NoSync:  true,
	}
	db, err := bbolt.Open(filePath, 0600, options)
	if err != nil {
		return nil, err
	}

	initBucket := func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	}
	if err := db.Update(initBucket); err != nil {
		return nil, err
	}

	return &fileStorageClient{db}, nil
}

// Set will store data. The data can be retrieved using the same key
func (c *fileStorageClient) Set(key string, value []byte) error {
	return c.Batch(SetOperation(key, value))
}

// Batch executes the specified operations in order. Get operation results are updated in place
func (c *fileStorageClient) Batch(ops ...Operation) error {
	batch := func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("storage not initialized")
		}

		var err error
		for _, op := range ops {
			switch op.Type {
			case Get:
				op.Value = bucket.Get([]byte(op.Key))
			case Set:
				err = bucket.Put([]byte(op.Key), op.Value)
			case Delete:
				err = bucket.Delete([]byte(op.Key))
			default:
				return errors.New("wrong operation type")
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	return c.db.Update(batch)
}

// Close will close the database
func (c *fileStorageClient) Close() error {
	return c.db.Close()
}
