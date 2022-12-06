/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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
