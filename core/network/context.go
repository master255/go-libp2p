package network

import (
	"context"
	"time"
)

// DialPeerTimeout is the default timeout for a single call to `DialPeer`. When
// there are multiple concurrent calls to `DialPeer`, this timeout will apply to
// each independently.
var DialPeerTimeout = 60 * time.Second

type noDialCtxKey struct{}
type dialPeerTimeoutCtxKey struct{}
type forceDirectDialCtxKey struct{}
type allowLimitedConnCtxKey struct{}
type simConnectCtxKey struct{ isClient bool }

var noDial = noDialCtxKey{}
var forceDirectDial = forceDirectDialCtxKey{}
var allowLimitedConn = allowLimitedConnCtxKey{}
var simConnectIsServer = simConnectCtxKey{}
var simConnectIsClient = simConnectCtxKey{isClient: true}

// EXPERIMENTAL
// WithForceDirectDial constructs a new context with an option that instructs the network
// to attempt to force a direct connection to a peer via a dial even if a proxied connection to it already exists.
func WithForceDirectDial(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, forceDirectDial, reason)
}

// EXPERIMENTAL
// GetForceDirectDial returns true if the force direct dial option is set in the context.
func GetForceDirectDial(ctx context.Context) (forceDirect bool, reason string) {
	v := ctx.Value(forceDirectDial)
	if v != nil {
		return true, v.(string)
	}

	return false, ""
}

// WithSimultaneousConnect constructs a new context with an option that instructs the transport
// to apply hole punching logic where applicable.
// EXPERIMENTAL
func WithSimultaneousConnect(ctx context.Context, isClient bool, reason string) context.Context {
	if isClient {
		return context.WithValue(ctx, simConnectIsClient, reason)
	}
	return context.WithValue(ctx, simConnectIsServer, reason)
}

// GetSimultaneousConnect returns true if the simultaneous connect option is set in the context.
// EXPERIMENTAL
func GetSimultaneousConnect(ctx context.Context) (simconnect bool, isClient bool, reason string) {
	if v := ctx.Value(simConnectIsClient); v != nil {
		return true, true, v.(string)
	}
	if v := ctx.Value(simConnectIsServer); v != nil {
		return true, false, v.(string)
	}
	return false, false, ""
}

// WithNoDial constructs a new context with an option that instructs the network
// to not attempt a new dial when opening a stream.
func WithNoDial(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, noDial, reason)
}

// GetNoDial returns true if the no dial option is set in the context.
func GetNoDial(ctx context.Context) (nodial bool, reason string) {
	v := ctx.Value(noDial)
	if v != nil {
		return true, v.(string)
	}

	return false, ""
}

// GetDialPeerTimeout returns the current DialPeer timeout (or the default).
func GetDialPeerTimeout(ctx context.Context) time.Duration {
	if to, ok := ctx.Value(dialPeerTimeoutCtxKey{}).(time.Duration); ok {
		return to
	}
	return DialPeerTimeout
}

// WithDialPeerTimeout returns a new context with the DialPeer timeout applied.
//
// This timeout overrides the default DialPeerTimeout and applies per-dial
// independently.
func WithDialPeerTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, dialPeerTimeoutCtxKey{}, timeout)
}

// WithAllowLimitedConn constructs a new context with an option that instructs
// the network that it is acceptable to use a limited connection when opening a
// new stream.
func WithAllowLimitedConn(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, allowLimitedConn, reason)
}

// WithUseTransient constructs a new context with an option that instructs the network
// that it is acceptable to use a transient connection when opening a new stream.
//
// Deprecated: Use WithAllowLimitedConn instead.
func WithUseTransient(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, allowLimitedConn, reason)
}

// GetAllowLimitedConn returns true if the allow limited conn option is set in the context.
func GetAllowLimitedConn(ctx context.Context) (usetransient bool, reason string) {
	v := ctx.Value(allowLimitedConn)
	if v != nil {
		return true, v.(string)
	}
	return false, ""
}

// GetUseTransient returns true if the use transient option is set in the context.
//
// Deprecated: Use GetAllowLimitedConn instead.
func GetUseTransient(ctx context.Context) (usetransient bool, reason string) {
	v := ctx.Value(allowLimitedConn)
	if v != nil {
		return true, v.(string)
	}
	return false, ""
}
