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

package transport

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	a2a "github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/internal/jsonrpc2"
)

// We track the incoming request ID inside the handler context using
// idContextValue, so that notifications and server->client calls that occur in
// the course of handling incoming requests are correlated with the incoming
// request that caused them, and can be dispatched as server-sent events to the
// correct HTTP request.
//
// Currently, this is implemented in [ServerSession.handle]. This is not ideal,
// because it means that a user of the A2A package couldn't implement the
// streamable transport, as they'd lack this privileged access.
//
// If we ever wanted to expose this mechanism, we have a few options:
//  1. Make ServerSession an interface, and provide an implementation of
//     ServerSession to handlers that closes over the incoming request ID.
//  2. Expose a 'HandlerTransport' interface that allows transports to provide
//     a handler middleware, so that we don't hard-code this behavior in
//     ServerSession.handle.
//  3. Add a `func ForRequest(context.Context) jsonrpc.ID` accessor that lets
//     any transport access the incoming request ID.
//
// For now, by giving only the StreamableServerTransport access to the request
// ID, we avoid having to make this API decision.
type idContextKey struct{}

// MethodHandler handles A2A messages.
// For methods, exactly one of the return values must be nil.
// For notifications, both must be nil.
type MethodHandler[S Session] func(ctx context.Context, session S, method a2a.Method, params a2a.Params) (result a2a.Result, err error)

// methodHandler is a MethodHandler[Session] for some session.
// We need to give up type safety here, or we will end up with a type cycle somewhere
// else. For example, if Session.methodHandler returned a MethodHandler[Session],
// the compiler would complain.
type methodHandler any // MethodHandler[*ClientSession] | MethodHandler[*ServerSession]

// Session is either a [ClientSession] or a [ServerSession].
type Session interface {
	*ClientSession | *ServerSession

	// ID returns the session ID, or the empty string if there is none.
	ID() string

	sendingMethodInfo() map[a2a.Method]methodInfo
	receivingMethodInfo() map[a2a.Method]methodInfo
	sendingMethodHandler() methodHandler
	receivingMethodHandler() methodHandler
	Conn() *jsonrpc2.Connection
	Close() error
	Wait() error
}

// Middleware is a function from [MethodHandler] to [MethodHandler].
type Middleware[S Session] func(MethodHandler[S]) MethodHandler[S]

// AddMiddleware wraps the handler in the middleware functions.
func AddMiddleware[S Session](handler *MethodHandler[S], middleware []Middleware[S]) {
	for _, m := range slices.Backward(middleware) {
		*handler = m(*handler)
	}
}

// DefaultSendingMethodHandler is the initial [MethodHandler] for servers and clients, before being wrapped by middleware.
func DefaultSendingMethodHandler[S Session](ctx context.Context, session S, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	info, ok := session.sendingMethodInfo()[method]
	if !ok {
		// This can be called from user code, with an arbitrary value for method.
		return nil, fmt.Errorf("DefaultSendingMethodHandler: %w", jsonrpc2.ErrNotHandled)
	}
	// Create the result to unmarshal into.
	// The concrete type of the result is the return type of the receiving function.
	res := info.newResult()
	if err := call(ctx, session.Conn(), method, params, res); err != nil {
		return nil, fmt.Errorf("call %q: %w", method, err)
	}
	return res, nil
}

// DefaultReceivingMethodHandler is the initial [MethodHandler] for servers and clients, before being wrapped by middleware.
func DefaultReceivingMethodHandler[S Session](ctx context.Context, session S, method a2a.Method, params a2a.Params) (a2a.Result, error) {
	info, ok := session.receivingMethodInfo()[method]
	if !ok {
		// This can be called from user code, with an arbitrary value for method.
		return nil, fmt.Errorf("DefaultReceivingMethodHandler: %w", jsonrpc2.ErrNotHandled)
	}
	return info.handleMethod.(MethodHandler[S])(ctx, session, method, params)
}

// call executes and awaits a jsonrpc2 call on the given connection,
// translating errors into the a2a domain.
func call(ctx context.Context, conn *jsonrpc2.Connection, method a2a.Method, params a2a.Params, result a2a.Result) error {
	// TODO(zchee): the "%w"s in this function effectively make [a2a.Error] part of the API.
	// Consider alternatives.
	call := conn.Call(ctx, string(method), params)
	err := call.Await(ctx, result)

	switch {
	case errors.Is(err, a2a.ErrClientClosing), errors.Is(err, a2a.ErrServerClosing):
		return fmt.Errorf("calling %q: %w", method, ErrConnectionClosed)
	case err != nil:
		return fmt.Errorf("calling %q: %w", method, err)
	}

	return nil
}

func handleSend[R a2a.Result, S Session](ctx context.Context, session S, method a2a.Method, params a2a.Params) (R, error) {
	mh := session.sendingMethodHandler().(MethodHandler[S])
	// mh might be user code, so ensure that it returns the right values for the jsonrpc2 protocol.
	res, err := mh(ctx, session, method, params)
	if err != nil {
		var z R
		return z, fmt.Errorf("handleSend: send %q: %w", method, err)
	}

	switch res := res.(type) {
	case *messageOrTaskResult:
		if res.result == nil {
			var z R
			return z, fmt.Errorf("handleSend: send %q: nil result in messageOrTaskResult", method)
		}
		if result, ok := res.result.(R); ok {
			return result, nil
		}
		var z R
		return z, fmt.Errorf("handleSend: send %q: result type %T cannot be converted to %T",
			method, res.result, z)
	case *sendStreamingMessageResponse:
		if res.result == nil {
			var z R
			return z, fmt.Errorf("handleSend: send %q: nil result in sendStreamingMessageResponse", method)
		}
		if result, ok := res.result.(R); ok {
			return result, nil
		}
		var z R
		return z, fmt.Errorf("handleSend: send %q: result type %T cannot be converted to %T",
			method, res.result, z)
	}

	if result, ok := res.(R); ok {
		return result, nil
	}
	var z R
	return z, fmt.Errorf("handleSend: send %q: result type %T cannot be converted to %T", method, res, z)
}

func handleReceive[S Session](ctx context.Context, session S, req *jsonrpc2.Request) (a2a.Result, error) {
	info, ok := session.receivingMethodInfo()[a2a.Method(req.Method)]
	if !ok {
		return nil, jsonrpc2.ErrNotHandled
	}
	params, err := info.unmarshalParams(req.Params)
	if err != nil {
		return nil, fmt.Errorf("handleReceive: unmarshal params %q: %w", req.Method, err)
	}

	mh := session.receivingMethodHandler().(MethodHandler[S])
	// mh might be user code, so ensure that it returns the right values for the jsonrpc2 protocol.
	res, err := mh(ctx, session, a2a.Method(req.Method), params)
	if err != nil {
		return nil, fmt.Errorf("handleReceive: receive %q: %w", req.Method, err)
	}
	return res, nil
}

// orZero is the helper methods to avoid typed nil.
func orZero[T any, P *U, U any](p P) T {
	if p == nil {
		var zero T
		return zero
	}
	return any(p).(T)
}

// methodInfo is information about sending and receiving a method.
type methodInfo struct {
	// Unmarshal params from the wire into a Params struct.
	// Used on the receive side.
	unmarshalParams func(jsontext.Value) (a2a.Params, error)
	// Run the code when a call to the method is received.
	// Used on the receive side.
	handleMethod methodHandler
	// Create a pointer to a Result struct.
	// Used on the send side.
	newResult func() a2a.Result
}

// The following definitions support converting from typed to untyped method handlers.
// Type parameter meanings:
// - S: sessions
// - P: params
// - R: results

// typedMethodHandler is like a [MethodHandler], but with type information.
type typedMethodHandler[S Session, P a2a.Params, R a2a.Result] func(context.Context, S, P) (R, error)

type paramsPtr[T any] interface {
	*T
	a2a.Params
}

var (
	messageOrTaskType                = reflect.TypeFor[a2a.MessageOrTask]()
	sendStreamingMessageResponseType = reflect.TypeFor[a2a.SendStreamingMessageResponse]()
)

// newMethodInfo creates a [methodInfo] from a [typedMethodHandler].
func newMethodInfo[T any, S Session, P paramsPtr[T], R a2a.Result](h typedMethodHandler[S, P, R]) methodInfo {
	return methodInfo{
		unmarshalParams: func(v jsontext.Value) (a2a.Params, error) {
			var p P
			if v != nil {
				if err := json.Unmarshal(v, &p); err != nil {
					return nil, fmt.Errorf("unmarshaling %q into a %T: %w", v, p, err)
				}
			}
			return orZero[a2a.Params](p), nil
		},
		handleMethod: MethodHandler[S](func(ctx context.Context, session S, _ a2a.Method, params a2a.Params) (a2a.Result, error) {
			if params == nil {
				return h(ctx, session, nil)
			}
			return h(ctx, session, params.(P))
		}),
		// newResult is used on the send side, to construct the value to unmarshal the result into.
		// Special handling for interface types like MessageOrTask that cannot be instantiated directly.
		newResult: func() a2a.Result {
			t := reflect.TypeFor[R]()
			switch t {
			case messageOrTaskType:
				return &messageOrTaskResult{}
			case sendStreamingMessageResponseType:
				return &sendStreamingMessageResponse{}
			}
			// For concrete types, use reflection as before
			return reflect.New(t.Elem()).Interface().(R)
		},
	}
}

// sessionMethod is glue for creating a [typedMethodHandler] from a method on [ClientSession] or [ServerSession].
func sessionMethod[S Session, P a2a.Params, R a2a.Result](f func(S, context.Context, P) (R, error)) typedMethodHandler[S, P, R] {
	return func(ctx context.Context, session S, p P) (R, error) {
		return f(session, ctx, p)
	}
}

// serverMethod is glue for creating a typedMethodHandler from a method on Server.
func serverMethod[P a2a.Params, R a2a.Result](f func(Server, context.Context, *ServerSession, P) (R, error)) typedMethodHandler[*ServerSession, P, R] {
	return func(ctx context.Context, ss *ServerSession, p P) (R, error) {
		return f(ss.Server, ctx, ss, p)
	}
}

// clientMethod is glue for creating a typedMethodHandler from a method on Server.
func clientMethod[P a2a.Params, R a2a.Result](f func(Client, context.Context, *ClientSession, P) (R, error)) typedMethodHandler[*ClientSession, P, R] {
	return func(ctx context.Context, cs *ClientSession, p P) (R, error) {
		return f(cs.Client, ctx, cs, p)
	}
}
