package lambda

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const pathSegmentFunctions = "functions"

// CreateFunction handles the CreateFunction API.
func (s *Service) CreateFunction(w http.ResponseWriter, r *http.Request) {
	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFunctionError(w, ErrInvalidParameterValue, "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.FunctionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	if req.Role == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "Role is required", http.StatusBadRequest)

		return
	}

	fn, err := s.storage.CreateFunction(r.Context(), &req)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceConflict {
				status = http.StatusConflict
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := functionToCreateResponse(fn)
	writeJSONResponse(w, http.StatusCreated, resp)
}

// GetFunction handles the GetFunction API.
func (s *Service) GetFunction(w http.ResponseWriter, r *http.Request) {
	functionName := extractFunctionName(r.URL.Path)
	if functionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	fn, err := s.storage.GetFunction(r.Context(), functionName)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceNotFound {
				status = http.StatusNotFound
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := &GetFunctionResponse{
		Configuration: functionToConfiguration(fn),
		Code: &FunctionCodeLocation{
			RepositoryType: "S3",
			Location:       s.baseURL + "/lambda-code/" + functionName,
		},
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// DeleteFunction handles the DeleteFunction API.
func (s *Service) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	functionName := extractFunctionName(r.URL.Path)
	if functionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	err := s.storage.DeleteFunction(r.Context(), functionName)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceNotFound {
				status = http.StatusNotFound
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListFunctions handles the ListFunctions API.
func (s *Service) ListFunctions(w http.ResponseWriter, r *http.Request) {
	marker := r.URL.Query().Get("Marker")
	maxItemsStr := r.URL.Query().Get("MaxItems")

	maxItems := 50

	if maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil {
			maxItems = parsed
		}
	}

	functions, nextMarker, err := s.storage.ListFunctions(r.Context(), marker, maxItems)
	if err != nil {
		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	configs := make([]*FunctionConfiguration, 0, len(functions))
	for _, fn := range functions {
		configs = append(configs, functionToConfiguration(fn))
	}

	resp := &ListFunctionsResponse{
		Functions:  configs,
		NextMarker: nextMarker,
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// UpdateFunctionCode handles the UpdateFunctionCode API.
func (s *Service) UpdateFunctionCode(w http.ResponseWriter, r *http.Request) {
	functionName := extractFunctionNameFromCodePath(r.URL.Path)
	if functionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	var req UpdateFunctionCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFunctionError(w, ErrInvalidParameterValue, "Invalid request body", http.StatusBadRequest)

		return
	}

	fn, err := s.storage.UpdateFunctionCode(r.Context(), functionName, &req)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceNotFound {
				status = http.StatusNotFound
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := functionToCreateResponse(fn)
	writeJSONResponse(w, http.StatusOK, resp)
}

// UpdateFunctionConfiguration handles the UpdateFunctionConfiguration API.
func (s *Service) UpdateFunctionConfiguration(w http.ResponseWriter, r *http.Request) {
	functionName := extractFunctionNameFromConfigPath(r.URL.Path)
	if functionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	var req UpdateFunctionConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFunctionError(w, ErrInvalidParameterValue, "Invalid request body", http.StatusBadRequest)

		return
	}

	fn, err := s.storage.UpdateFunctionConfiguration(r.Context(), functionName, &req)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceNotFound {
				status = http.StatusNotFound
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp := functionToCreateResponse(fn)
	writeJSONResponse(w, http.StatusOK, resp)
}

// Invoke handles the Invoke API.
func (s *Service) Invoke(w http.ResponseWriter, r *http.Request) {
	functionName := extractFunctionNameFromInvokePath(r.URL.Path)
	if functionName == "" {
		writeFunctionError(w, ErrInvalidParameterValue, "FunctionName is required", http.StatusBadRequest)

		return
	}

	// Verify function exists.
	if err := s.verifyFunctionExists(r, functionName, w); err != nil {
		return
	}

	// Read the payload.
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeFunctionError(w, ErrInvalidParameterValue, "Failed to read request body", http.StatusBadRequest)

		return
	}

	writeInvokeResponse(w, r, payload)
}

// verifyFunctionExists checks if a function exists and writes an error response if not.
func (s *Service) verifyFunctionExists(r *http.Request, functionName string, w http.ResponseWriter) error {
	_, err := s.storage.GetFunction(r.Context(), functionName)
	if err != nil {
		var lambdaErr *FunctionError
		if errors.As(err, &lambdaErr) {
			status := http.StatusBadRequest
			if lambdaErr.Type == ErrResourceNotFound {
				status = http.StatusNotFound
			}

			writeFunctionError(w, lambdaErr.Type, lambdaErr.Message, status)

			return err
		}

		writeFunctionError(w, ErrServiceException, "Internal server error", http.StatusInternalServerError)

		return err
	}

	return nil
}

// writeInvokeResponse writes the invoke response based on invocation type.
func writeInvokeResponse(w http.ResponseWriter, r *http.Request, payload []byte) {
	invocationType := r.Header.Get("X-Amz-Invocation-Type")
	if invocationType == "" {
		invocationType = "RequestResponse"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Amz-Executed-Version", "$LATEST")
	w.Header().Set("X-Amz-Request-Id", uuid.New().String())

	switch invocationType {
	case "Event":
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("{}"))
	case "DryRun":
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusOK)
		if len(payload) == 0 {
			_, _ = w.Write([]byte("null"))
		} else {
			_, _ = w.Write(payload)
		}
	}
}

// extractFunctionName extracts function name from path like /2015-03-31/functions/{name}.
func extractFunctionName(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 && parts[1] == pathSegmentFunctions {
		return parts[2]
	}

	return ""
}

// extractFunctionNameFromCodePath extracts function name from path like /2015-03-31/functions/{name}/code.
func extractFunctionNameFromCodePath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 4 && parts[1] == pathSegmentFunctions && parts[3] == "code" {
		return parts[2]
	}

	return ""
}

// extractFunctionNameFromConfigPath extracts function name from path like /2015-03-31/functions/{name}/configuration.
func extractFunctionNameFromConfigPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 4 && parts[1] == pathSegmentFunctions && parts[3] == "configuration" {
		return parts[2]
	}

	return ""
}

// extractFunctionNameFromInvokePath extracts function name from path like /2015-03-31/functions/{name}/invocations.
func extractFunctionNameFromInvokePath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 4 && parts[1] == pathSegmentFunctions && parts[3] == "invocations" {
		return parts[2]
	}

	return ""
}

// functionToCreateResponse converts a Function to CreateFunctionResponse.
func functionToCreateResponse(fn *Function) *CreateFunctionResponse {
	return &CreateFunctionResponse{
		FunctionName:    fn.FunctionName,
		FunctionArn:     fn.FunctionArn,
		Runtime:         fn.Runtime,
		Role:            fn.Role,
		Handler:         fn.Handler,
		CodeSize:        fn.CodeSize,
		Description:     fn.Description,
		Timeout:         fn.Timeout,
		MemorySize:      fn.MemorySize,
		LastModified:    fn.LastModified.Format("2006-01-02T15:04:05.000+0000"),
		CodeSha256:      fn.CodeSha256,
		Version:         fn.Version,
		State:           fn.State,
		StateReason:     fn.StateReason,
		StateReasonCode: fn.StateReasonCode,
		PackageType:     fn.PackageType,
		Architectures:   fn.Architectures,
		Environment:     fn.Environment,
	}
}

// functionToConfiguration converts a Function to FunctionConfiguration.
func functionToConfiguration(fn *Function) *FunctionConfiguration {
	return &FunctionConfiguration{
		FunctionName:    fn.FunctionName,
		FunctionArn:     fn.FunctionArn,
		Runtime:         fn.Runtime,
		Role:            fn.Role,
		Handler:         fn.Handler,
		CodeSize:        fn.CodeSize,
		Description:     fn.Description,
		Timeout:         fn.Timeout,
		MemorySize:      fn.MemorySize,
		LastModified:    fn.LastModified.Format("2006-01-02T15:04:05.000+0000"),
		CodeSha256:      fn.CodeSha256,
		Version:         fn.Version,
		State:           fn.State,
		StateReason:     fn.StateReason,
		StateReasonCode: fn.StateReasonCode,
		PackageType:     fn.PackageType,
		Architectures:   fn.Architectures,
		Environment:     fn.Environment,
	}
}

// writeJSONResponse writes a JSON response.
func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Amzn-Requestid", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeFunctionError writes a Lambda error response.
func writeFunctionError(w http.ResponseWriter, errType, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Amzn-Requestid", uuid.New().String())
	w.Header().Set("X-Amzn-Errortype", errType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"Type":    errType,
		"Message": message,
	})
}
