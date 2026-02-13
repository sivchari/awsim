package apigateway

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CreateRestApi handles the CreateRestApi API.
func (s *Service) CreateRestApi(w http.ResponseWriter, r *http.Request) {
	var req CreateRestApiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	if req.Name == "" {
		writeError(w, "BadRequestException", "Name is required", http.StatusBadRequest)

		return
	}

	api, err := s.storage.CreateRestApi(r.Context(), &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toRestApiResponse(api)
	writeResponse(w, resp, http.StatusCreated)
}

// GetRestApi handles the GetRestApi API.
func (s *Service) GetRestApi(w http.ResponseWriter, r *http.Request) {
	restApiID := extractPathParam(r.URL.Path, "/apigateway/restapis/")

	api, err := s.storage.GetRestApi(r.Context(), restApiID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toRestApiResponse(api)
	writeResponse(w, resp, http.StatusOK)
}

// GetRestApis handles the GetRestApis API.
func (s *Service) GetRestApis(w http.ResponseWriter, r *http.Request) {
	apis, nextPosition, err := s.storage.GetRestApis(r.Context(), 25, "")
	if err != nil {
		handleError(w, err)

		return
	}

	items := make([]CreateRestApiResponse, len(apis))

	for i, api := range apis {
		items[i] = *toRestApiResponse(api)
	}

	resp := &GetRestApisResponse{
		Items:    items,
		Position: nextPosition,
	}

	writeResponse(w, resp, http.StatusOK)
}

// DeleteRestApi handles the DeleteRestApi API.
func (s *Service) DeleteRestApi(w http.ResponseWriter, r *http.Request) {
	restApiID := extractPathParam(r.URL.Path, "/apigateway/restapis/")

	if err := s.storage.DeleteRestApi(r.Context(), restApiID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// CreateResource handles the CreateResource API.
func (s *Service) CreateResource(w http.ResponseWriter, r *http.Request) {
	restApiID, parentID := extractResourceParams(r.URL.Path)

	var req CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	resource, err := s.storage.CreateResource(r.Context(), restApiID, parentID, req.PathPart)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toResourceResponse(resource)
	writeResponse(w, resp, http.StatusCreated)
}

// GetResource handles the GetResource API.
func (s *Service) GetResource(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID := extractRestApiAndResourceID(r.URL.Path)

	resource, err := s.storage.GetResource(r.Context(), restApiID, resourceID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toResourceResponse(resource)
	writeResponse(w, resp, http.StatusOK)
}

// GetResources handles the GetResources API.
func (s *Service) GetResources(w http.ResponseWriter, r *http.Request) {
	// Extract restApiId from path: /apigateway/restapis/{restApiId}/resources
	path := strings.TrimPrefix(r.URL.Path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")
	restApiID := parts[0]

	resources, nextPosition, err := s.storage.GetResources(r.Context(), restApiID, 25, "")
	if err != nil {
		handleError(w, err)

		return
	}

	items := make([]ResourceResponse, len(resources))

	for i, res := range resources {
		items[i] = *toResourceResponse(res)
	}

	resp := &GetResourcesResponse{
		Items:    items,
		Position: nextPosition,
	}

	writeResponse(w, resp, http.StatusOK)
}

// DeleteResource handles the DeleteResource API.
func (s *Service) DeleteResource(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID := extractRestApiAndResourceID(r.URL.Path)

	if err := s.storage.DeleteResource(r.Context(), restApiID, resourceID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// PutMethod handles the PutMethod API.
func (s *Service) PutMethod(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID, httpMethod := extractMethodParams(r.URL.Path)

	var req PutMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	method, err := s.storage.PutMethod(r.Context(), restApiID, resourceID, httpMethod, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toMethodOutput(method)
	writeResponse(w, resp, http.StatusCreated)
}

// GetMethod handles the GetMethod API.
func (s *Service) GetMethod(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID, httpMethod := extractMethodParams(r.URL.Path)

	method, err := s.storage.GetMethod(r.Context(), restApiID, resourceID, httpMethod)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toMethodOutput(method)
	writeResponse(w, resp, http.StatusOK)
}

// PutIntegration handles the PutIntegration API.
func (s *Service) PutIntegration(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID, httpMethod := extractIntegrationParams(r.URL.Path)

	var req PutIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	integration, err := s.storage.PutIntegration(r.Context(), restApiID, resourceID, httpMethod, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toIntegrationOutput(integration)
	writeResponse(w, resp, http.StatusCreated)
}

// GetIntegration handles the GetIntegration API.
func (s *Service) GetIntegration(w http.ResponseWriter, r *http.Request) {
	restApiID, resourceID, httpMethod := extractIntegrationParams(r.URL.Path)

	integration, err := s.storage.GetIntegration(r.Context(), restApiID, resourceID, httpMethod)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toIntegrationOutput(integration)
	writeResponse(w, resp, http.StatusOK)
}

// CreateDeployment handles the CreateDeployment API.
func (s *Service) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	restApiID := extractDeploymentRestApiID(r.URL.Path)

	var req CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	deployment, err := s.storage.CreateDeployment(r.Context(), restApiID, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toDeploymentResponse(deployment)
	writeResponse(w, resp, http.StatusCreated)
}

// GetDeployment handles the GetDeployment API.
func (s *Service) GetDeployment(w http.ResponseWriter, r *http.Request) {
	restApiID, deploymentID := extractRestApiAndDeploymentID(r.URL.Path)

	deployment, err := s.storage.GetDeployment(r.Context(), restApiID, deploymentID)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toDeploymentResponse(deployment)
	writeResponse(w, resp, http.StatusOK)
}

// GetDeployments handles the GetDeployments API.
func (s *Service) GetDeployments(w http.ResponseWriter, r *http.Request) {
	restApiID := extractDeploymentRestApiID(r.URL.Path)

	deployments, nextPosition, err := s.storage.GetDeployments(r.Context(), restApiID, 25, "")
	if err != nil {
		handleError(w, err)

		return
	}

	items := make([]DeploymentResponse, len(deployments))

	for i, d := range deployments {
		items[i] = *toDeploymentResponse(d)
	}

	resp := &GetDeploymentsResponse{
		Items:    items,
		Position: nextPosition,
	}

	writeResponse(w, resp, http.StatusOK)
}

// DeleteDeployment handles the DeleteDeployment API.
func (s *Service) DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	restApiID, deploymentID := extractRestApiAndDeploymentID(r.URL.Path)

	if err := s.storage.DeleteDeployment(r.Context(), restApiID, deploymentID); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// CreateStage handles the CreateStage API.
func (s *Service) CreateStage(w http.ResponseWriter, r *http.Request) {
	restApiID := extractStageRestApiID(r.URL.Path)

	var req CreateStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BadRequestException", "Invalid request body", http.StatusBadRequest)

		return
	}

	stage, err := s.storage.CreateStage(r.Context(), restApiID, &req)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toStageResponse(stage)
	writeResponse(w, resp, http.StatusCreated)
}

// GetStage handles the GetStage API.
func (s *Service) GetStage(w http.ResponseWriter, r *http.Request) {
	restApiID, stageName := extractRestApiAndStageName(r.URL.Path)

	stage, err := s.storage.GetStage(r.Context(), restApiID, stageName)
	if err != nil {
		handleError(w, err)

		return
	}

	resp := toStageResponse(stage)
	writeResponse(w, resp, http.StatusOK)
}

// GetStages handles the GetStages API.
func (s *Service) GetStages(w http.ResponseWriter, r *http.Request) {
	restApiID := extractStageRestApiID(r.URL.Path)

	stages, err := s.storage.GetStages(r.Context(), restApiID)
	if err != nil {
		handleError(w, err)

		return
	}

	items := make([]StageResponse, len(stages))

	for i, stage := range stages {
		items[i] = *toStageResponse(stage)
	}

	resp := &GetStagesResponse{
		Items: items,
	}

	writeResponse(w, resp, http.StatusOK)
}

// DeleteStage handles the DeleteStage API.
func (s *Service) DeleteStage(w http.ResponseWriter, r *http.Request) {
	restApiID, stageName := extractRestApiAndStageName(r.URL.Path)

	if err := s.storage.DeleteStage(r.Context(), restApiID, stageName); err != nil {
		handleError(w, err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// toRestApiResponse converts a RestApi to CreateRestApiResponse.
func toRestApiResponse(api *RestApi) *CreateRestApiResponse {
	return &CreateRestApiResponse{
		ID:                     api.ID,
		Name:                   api.Name,
		Description:            api.Description,
		CreatedDate:            float64(api.CreatedDate.Unix()),
		Version:                api.Version,
		ApiKeySource:           api.ApiKeySource,
		EndpointConfiguration:  api.EndpointConfiguration,
		DisableExecuteApiEndpt: api.DisableExecuteApiEndpt,
		Tags:                   api.Tags,
		RootResourceId:         api.RootResourceID,
	}
}

// toResourceResponse converts a Resource to ResourceResponse.
func toResourceResponse(res *Resource) *ResourceResponse {
	methods := make(map[string]MethodOutput)

	for k, m := range res.ResourceMethods {
		methods[k] = *toMethodOutput(&m)
	}

	return &ResourceResponse{
		ID:              res.ID,
		ParentID:        res.ParentID,
		PathPart:        res.PathPart,
		Path:            res.Path,
		ResourceMethods: methods,
	}
}

// toMethodOutput converts a Method to MethodOutput.
func toMethodOutput(m *Method) *MethodOutput {
	output := &MethodOutput{
		HTTPMethod:        m.HTTPMethod,
		AuthorizationType: m.AuthorizationType,
		ApiKeyRequired:    m.ApiKeyRequired,
		OperationName:     m.OperationName,
	}

	if m.MethodIntegration != nil {
		output.MethodIntegration = toIntegrationOutput(m.MethodIntegration)
	}

	return output
}

// toIntegrationOutput converts an Integration to IntegrationOutput.
func toIntegrationOutput(i *Integration) *IntegrationOutput {
	return &IntegrationOutput{
		Type:                i.Type,
		HTTPMethod:          i.HTTPMethod,
		URI:                 i.URI,
		ConnectionType:      i.ConnectionType,
		ConnectionID:        i.ConnectionID,
		PassthroughBehavior: i.PassthroughBehavior,
		ContentHandling:     i.ContentHandling,
		TimeoutInMillis:     i.TimeoutInMillis,
		CacheNamespace:      i.CacheNamespace,
		CacheKeyParameters:  i.CacheKeyParameters,
		RequestParameters:   i.RequestParameters,
		RequestTemplates:    i.RequestTemplates,
	}
}

// toDeploymentResponse converts a Deployment to DeploymentResponse.
func toDeploymentResponse(d *Deployment) *DeploymentResponse {
	return &DeploymentResponse{
		ID:          d.ID,
		Description: d.Description,
		CreatedDate: float64(d.CreatedDate.Unix()),
	}
}

// toStageResponse converts a Stage to StageResponse.
func toStageResponse(s *Stage) *StageResponse {
	return &StageResponse{
		StageName:           s.StageName,
		DeploymentID:        s.DeploymentID,
		Description:         s.Description,
		CacheClusterEnabled: s.CacheClusterEnabled,
		CacheClusterSize:    s.CacheClusterSize,
		CreatedDate:         float64(s.CreatedDate.Unix()),
		LastUpdatedDate:     float64(s.LastUpdatedDate.Unix()),
		Tags:                s.Tags,
	}
}

// extractPathParam extracts the path parameter after the given prefix.
func extractPathParam(path, prefix string) string {
	return strings.TrimPrefix(path, prefix)
}

// extractResourceParams extracts restApiId and parentId from the path.
func extractResourceParams(path string) (restApiID, parentID string) {
	// Path: /apigateway/restapis/{restApiId}/resources/{parentId}
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 3 {
		return parts[0], parts[2]
	}

	return "", ""
}

// extractRestApiAndResourceID extracts restApiId and resourceId from the path.
func extractRestApiAndResourceID(path string) (restApiID, resourceID string) {
	// Path: /apigateway/restapis/{restApiId}/resources/{resourceId}
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 3 {
		return parts[0], parts[2]
	}

	return "", ""
}

// extractMethodParams extracts restApiId, resourceId, and httpMethod from the path.
func extractMethodParams(path string) (restApiID, resourceID, httpMethod string) {
	// Path: /apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 5 {
		return parts[0], parts[2], parts[4]
	}

	return "", "", ""
}

// extractIntegrationParams extracts restApiId, resourceId, and httpMethod from the path.
func extractIntegrationParams(path string) (restApiID, resourceID, httpMethod string) {
	// Path: /apigateway/restapis/{restApiId}/resources/{resourceId}/methods/{httpMethod}/integration
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 6 {
		return parts[0], parts[2], parts[4]
	}

	return "", "", ""
}

// extractDeploymentRestApiID extracts restApiId from the deployments path.
func extractDeploymentRestApiID(path string) string {
	// Path: /apigateway/restapis/{restApiId}/deployments
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 1 {
		return parts[0]
	}

	return ""
}

// extractRestApiAndDeploymentID extracts restApiId and deploymentId from the path.
func extractRestApiAndDeploymentID(path string) (restApiID, deploymentID string) {
	// Path: /apigateway/restapis/{restApiId}/deployments/{deploymentId}
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 3 {
		return parts[0], parts[2]
	}

	return "", ""
}

// extractStageRestApiID extracts restApiId from the stages path.
func extractStageRestApiID(path string) string {
	// Path: /apigateway/restapis/{restApiId}/stages
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 1 {
		return parts[0]
	}

	return ""
}

// extractRestApiAndStageName extracts restApiId and stageName from the path.
func extractRestApiAndStageName(path string) (restApiID, stageName string) {
	// Path: /apigateway/restapis/{restApiId}/stages/{stageName}
	path = strings.TrimPrefix(path, "/apigateway/restapis/")
	parts := strings.Split(path, "/")

	if len(parts) >= 3 {
		return parts[0], parts[2]
	}

	return "", ""
}

// writeResponse writes a JSON response.
func writeResponse(w http.ResponseWriter, resp any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Type: code, Message: message})
}

// handleError handles service errors.
func handleError(w http.ResponseWriter, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		status := getErrorStatus(svcErr.Code)
		writeError(w, svcErr.Code, svcErr.Message, status)

		return
	}

	writeError(w, "InternalServiceError", err.Error(), http.StatusInternalServerError)
}

// getErrorStatus returns the HTTP status code for a given error code.
func getErrorStatus(code string) int {
	switch code {
	case "NotFoundException":
		return http.StatusNotFound
	case "BadRequestException":
		return http.StatusBadRequest
	case "ConflictException":
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}
