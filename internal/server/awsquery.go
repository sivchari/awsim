package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// QueryServiceHandler handles Query protocol requests for a specific service.
type QueryServiceHandler func(w http.ResponseWriter, r *http.Request)

// QueryProtocolDispatcher routes AWS Query protocol requests to the appropriate service
// based on the Action parameter.
type QueryProtocolDispatcher struct {
	// handlers maps service name to service handler.
	handlers map[string]QueryServiceHandler
}

// NewQueryProtocolDispatcher creates a new Query protocol dispatcher.
func NewQueryProtocolDispatcher() *QueryProtocolDispatcher {
	return &QueryProtocolDispatcher{
		handlers: make(map[string]QueryServiceHandler),
	}
}

// Register registers a service handler.
func (d *QueryProtocolDispatcher) Register(serviceName string, handler QueryServiceHandler) {
	d.handlers[serviceName] = handler
}

// ServeHTTP implements http.Handler and dispatches to the appropriate service.
func (d *QueryProtocolDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse form data.
	if err := r.ParseForm(); err != nil {
		writeQueryError(w, "InvalidParameterValue", "Failed to parse form data", http.StatusBadRequest)

		return
	}

	// Get Action parameter.
	action := r.FormValue("Action")
	if action == "" {
		writeQueryError(w, "MissingAction", "Action parameter is required", http.StatusBadRequest)

		return
	}

	// Convert form data to JSON and set X-Amz-Target header for the handler.
	jsonBody := formToJSON(r.Form)
	r.Body = io.NopCloser(bytes.NewReader(jsonBody))
	r.ContentLength = int64(len(jsonBody))
	r.Header.Set("Content-Type", "application/x-amz-json-1.0")

	// Dispatch to the appropriate handler based on the service action prefix.
	for serviceName, handler := range d.handlers {
		// Set the X-Amz-Target header for the handler.
		r.Header.Set("X-Amz-Target", serviceName+"."+action)
		handler(w, r)

		return
	}

	writeQueryError(w, "UnknownAction", "Unknown action: "+action, http.StatusBadRequest)
}

// formToJSON converts form values to JSON.
func formToJSON(form map[string][]string) []byte {
	result := make(map[string]any)

	for key, values := range form {
		if key == "Action" || key == "Version" {
			continue
		}

		// Convert key from Query format to JSON format.
		// e.g., "Attributes.entry.1.key" -> handled specially
		// Simple values: "Name" -> "Name"
		if len(values) == 1 {
			result[key] = values[0]
		} else if len(values) > 1 {
			result[key] = values
		}
	}

	// Handle nested attributes (like Attributes.entry.N.key/value).
	result = flattenAttributes(result)

	jsonBytes, _ := json.Marshal(result)

	return jsonBytes
}

// flattenAttributes converts nested Query protocol attributes to JSON format.
func flattenAttributes(data map[string]any) map[string]any {
	result := make(map[string]any)
	attrs := make(map[string]string)

	for key, value := range data {
		if !strings.HasPrefix(key, "Attributes.entry.") {
			result[key] = value

			continue
		}

		parseAttributeEntry(key, value, attrs)
	}

	buildAttributesMap(attrs, result)

	return result
}

// parseAttributeEntry parses an Attributes.entry.N.key/value pattern.
func parseAttributeEntry(key string, value any, attrs map[string]string) {
	parts := strings.Split(key, ".")
	if len(parts) != 4 {
		return
	}

	idx := parts[2]
	field := parts[3]

	strValue, ok := value.(string)
	if !ok {
		return
	}

	switch field {
	case "key":
		attrs[idx+"_key"] = strValue
	case "value":
		attrs[idx+"_value"] = strValue
	}
}

// buildAttributesMap builds the Attributes map from parsed key/value pairs.
func buildAttributesMap(attrs map[string]string, result map[string]any) {
	if len(attrs) == 0 {
		return
	}

	attrMap := make(map[string]string)

	for k, v := range attrs {
		if !strings.HasSuffix(k, "_key") {
			continue
		}

		idx := strings.TrimSuffix(k, "_key")

		if val, ok := attrs[idx+"_value"]; ok {
			attrMap[v] = val
		}
	}

	if len(attrMap) > 0 {
		result["Attributes"] = attrMap
	}
}

// writeQueryError writes an AWS Query protocol error response.
func writeQueryError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"__type":  code,
		"message": message,
	})
}
