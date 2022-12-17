/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/mitchellh/hashstructure/v2"
	"hash/fnv"
	"io"
	"k8s.io/klog/v2"
)

func SimpleHash(obj interface{}) string {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)

	if err != nil {
		klog.Errorf("Not able convert Data to hash, error: %s", err.Error())
		return ""
	}

	return fmt.Sprintf("%x", hash)
}

func Hash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HashFNV(s string) string {
	hasher := fnv.New32a()
	// Hash.Write never returns an error
	_, _ = hasher.Write([]byte(s))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// GenerateRandom generates random string.
func GenerateRandom(n int) string {
	b := make([]byte, 8)
	_, _ = io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:n]
}
