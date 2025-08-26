// Package errors defines gRPC server error constants.
package errors

import "errors"

// gRPC server error definitions.
var (
	// ErrGRPCServerInvalidTLSConfig indicates invalid TLS configuration.
	ErrGRPCServerInvalidTLSConfig = errors.New("invalid GRPC server TLS config")
)
