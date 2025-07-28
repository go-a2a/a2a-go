// Copyright 2025 The Go A2A Authors.
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
//
// SPDX-License-Identifier: Apache-2.0

// Package pool provides generic type pooling, and provides [*bytes.Buffer] and [*strings.Builder] pooling objects.
package pool

import (
	"bytes"
	"strings"
	"sync"

	"github.com/go-json-experiment/json/jsontext"
)

// Pool is a generics wrapper around [syncx.Pool] to provide strongly-typed object pooling.
type Pool[T any] struct {
	p sync.Pool
}

type Reseter interface {
	Reset()
}

// New returns a new [Pool] for T, and will use fn to construct new T's when the pool is empty.
func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get gets a T from the pool, or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put returns x into the pool.
func (p *Pool[T]) Put(x T) {
	if xx, ok := any(x).(Reseter); ok {
		xx.Reset()
	}
	p.p.Put(x)
}

// PoolVar wraps [sync.Pool].
type PoolVar[T any] struct {
	New     func() T
	p       sync.Pool
	newOnce sync.Once
}

func (p *PoolVar[T]) init() {
	p.newOnce.Do(func() {
		p.p.New = func() any {
			return p.New()
		}
	})
}

// Get wraps [sync.Pool.Get].
func (p *PoolVar[T]) Get() T {
	p.init()
	return p.p.Get().(T)
}

// Put wraps [sync.Pool.Put].
func (p *PoolVar[T]) Put(x T) {
	p.init()
	p.p.Put(x)
}

// Bytes provides the [*bytes.Bytes] pooling objects.
var Bytes = New(func() *bytes.Buffer {
	return &bytes.Buffer{}
})

// String provides the [*strings.Builder] pooling objects.
var String = New(func() *strings.Builder {
	return &strings.Builder{}
})

// Value provides the [*jsontext.Value] pooling objects.
var Value = New(func() *jsontext.Value {
	v := make(jsontext.Value, 1024)
	return &v
})

// Value provides the slice of [jsontext.Value] pooling objects.
var Values = New(func() []jsontext.Value {
	return []jsontext.Value{}
})
