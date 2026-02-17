package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
)

// CBORServiceHandler handles RPC v2 CBOR protocol requests for a specific service.
type CBORServiceHandler func(w http.ResponseWriter, r *http.Request, operation string)

// CBORProtocolDispatcher routes Smithy RPC v2 CBOR protocol requests to the appropriate service
// based on the URL path: /service/{serviceName}/operation/{operationName}.
type CBORProtocolDispatcher struct {
	// handlers maps service name to service handler
	// e.g., "GraniteServiceVersion20100801" -> CloudWatch handler
	handlers map[string]CBORServiceHandler
}

// NewCBORProtocolDispatcher creates a new CBOR protocol dispatcher.
func NewCBORProtocolDispatcher() *CBORProtocolDispatcher {
	return &CBORProtocolDispatcher{
		handlers: make(map[string]CBORServiceHandler),
	}
}

// Register registers a service handler for the given service name.
// The service name matches the one in the URL path:
// /service/{serviceName}/operation/{operationName}.
func (d *CBORProtocolDispatcher) Register(serviceName string, handler CBORServiceHandler) {
	d.handlers[serviceName] = handler
}

// ServeHTTP implements http.Handler and dispatches to the appropriate service.
// It handles requests to /service/{serviceName}/operation/{operationName}.
func (d *CBORProtocolDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /service/{serviceName}/operation/{operationName}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/service/") {
		WriteCBORError(w, "InvalidPath", "Path must start with /service/", http.StatusBadRequest)

		return
	}

	// Remove "/service/" prefix
	remaining := strings.TrimPrefix(path, "/service/")
	parts := strings.Split(remaining, "/operation/")

	if len(parts) != 2 {
		WriteCBORError(w, "InvalidPath", "Path must be /service/{serviceName}/operation/{operationName}", http.StatusBadRequest)

		return
	}

	serviceName := parts[0]
	operationName := parts[1]

	if serviceName == "" || operationName == "" {
		WriteCBORError(w, "InvalidPath", "Service name and operation name are required", http.StatusBadRequest)

		return
	}

	handler, ok := d.handlers[serviceName]
	if !ok {
		WriteCBORError(w, "UnknownService", "Unknown service: "+serviceName, http.StatusBadRequest)

		return
	}

	handler(w, r, operationName)
}

// DecodeCBORRequest decodes a CBOR request body into the given value.
func DecodeCBORRequest(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	if err := cbor.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal CBOR: %w", err)
	}

	return nil
}

// WriteCBORResponse writes a CBOR response with the smithy-protocol header.
func WriteCBORResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/cbor")
	w.Header().Set("smithy-protocol", "rpc-v2-cbor")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(http.StatusOK)

	encoded, err := cbor.Marshal(v)
	if err != nil {
		// Fall back to empty response on encode error
		return
	}

	_, _ = w.Write(encoded)
}

// WriteCBORError writes an error response in CBOR format.
func WriteCBORError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/cbor")
	w.Header().Set("smithy-protocol", "rpc-v2-cbor")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)

	errorResponse := map[string]string{
		"__type":  code,
		"message": message,
	}

	encoded, err := cbor.Marshal(errorResponse)
	if err != nil {
		// If CBOR encoding fails, write nothing
		return
	}

	_, _ = w.Write(encoded)
}
