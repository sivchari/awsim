package eks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// CreateCluster handles the CreateCluster operation.
func (s *Service) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var req CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name is required")
		return
	}

	if req.RoleArn == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Role ARN is required")
		return
	}

	cluster, err := s.storage.CreateCluster(r.Context(), &req)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			status := http.StatusBadRequest
			if eksErr.Code == "ResourceInUseException" {
				status = http.StatusConflict
			}
			writeError(w, status, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &CreateClusterResponse{Cluster: cluster})
}

// DeleteCluster handles the DeleteCluster operation.
func (s *Service) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	name := extractClusterName(r.URL.Path)
	if name == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name is required")
		return
	}

	cluster, err := s.storage.DeleteCluster(r.Context(), name)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			status := http.StatusNotFound
			if eksErr.Code == "ResourceInUseException" {
				status = http.StatusConflict
			}
			writeError(w, status, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DeleteClusterResponse{Cluster: cluster})
}

// DescribeCluster handles the DescribeCluster operation.
func (s *Service) DescribeCluster(w http.ResponseWriter, r *http.Request) {
	name := extractClusterName(r.URL.Path)
	if name == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name is required")
		return
	}

	cluster, err := s.storage.DescribeCluster(r.Context(), name)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			writeError(w, http.StatusNotFound, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DescribeClusterResponse{Cluster: cluster})
}

// ListClusters handles the ListClusters operation.
func (s *Service) ListClusters(w http.ResponseWriter, r *http.Request) {
	maxResults := 100
	if v := r.URL.Query().Get("maxResults"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxResults = n
		}
	}
	nextToken := r.URL.Query().Get("nextToken")

	clusters, next, err := s.storage.ListClusters(r.Context(), maxResults, nextToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &ListClustersResponse{
		Clusters:  clusters,
		NextToken: next,
	})
}

// CreateNodegroup handles the CreateNodegroup operation.
func (s *Service) CreateNodegroup(w http.ResponseWriter, r *http.Request) {
	clusterName := extractClusterName(r.URL.Path)
	if clusterName == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name is required")
		return
	}

	var req CreateNodegroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Invalid request body")
		return
	}

	req.ClusterName = clusterName

	if req.NodegroupName == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Nodegroup name is required")
		return
	}

	if req.NodeRole == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Node role is required")
		return
	}

	if len(req.Subnets) == 0 {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Subnets are required")
		return
	}

	nodegroup, err := s.storage.CreateNodegroup(r.Context(), &req)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			status := http.StatusBadRequest
			switch eksErr.Code {
			case "ResourceNotFoundException":
				status = http.StatusNotFound
			case "ResourceInUseException":
				status = http.StatusConflict
			}
			writeError(w, status, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &CreateNodegroupResponse{Nodegroup: nodegroup})
}

// DeleteNodegroup handles the DeleteNodegroup operation.
func (s *Service) DeleteNodegroup(w http.ResponseWriter, r *http.Request) {
	clusterName, nodegroupName := extractClusterAndNodegroupName(r.URL.Path)
	if clusterName == "" || nodegroupName == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name and nodegroup name are required")
		return
	}

	nodegroup, err := s.storage.DeleteNodegroup(r.Context(), clusterName, nodegroupName)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			writeError(w, http.StatusNotFound, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DeleteNodegroupResponse{Nodegroup: nodegroup})
}

// DescribeNodegroup handles the DescribeNodegroup operation.
func (s *Service) DescribeNodegroup(w http.ResponseWriter, r *http.Request) {
	clusterName, nodegroupName := extractClusterAndNodegroupName(r.URL.Path)
	if clusterName == "" || nodegroupName == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name and nodegroup name are required")
		return
	}

	nodegroup, err := s.storage.DescribeNodegroup(r.Context(), clusterName, nodegroupName)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			writeError(w, http.StatusNotFound, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DescribeNodegroupResponse{Nodegroup: nodegroup})
}

// ListNodegroups handles the ListNodegroups operation.
func (s *Service) ListNodegroups(w http.ResponseWriter, r *http.Request) {
	clusterName := extractClusterName(r.URL.Path)
	if clusterName == "" {
		writeError(w, http.StatusBadRequest, "InvalidParameterException", "Cluster name is required")
		return
	}

	maxResults := 100
	if v := r.URL.Query().Get("maxResults"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxResults = n
		}
	}
	nextToken := r.URL.Query().Get("nextToken")

	nodegroups, next, err := s.storage.ListNodegroups(r.Context(), clusterName, maxResults, nextToken)
	if err != nil {
		if eksErr, ok := err.(*Error); ok {
			writeError(w, http.StatusNotFound, eksErr.Code, eksErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &ListNodegroupsResponse{
		Nodegroups: nodegroups,
		NextToken:  next,
	})
}

// extractClusterName extracts the cluster name from the URL path.
// Expected paths: /clusters/{name} or /clusters/{name}/node-groups...
func extractClusterName(path string) string {
	// Remove leading slash and split
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// Expected: clusters/{name} or clusters/{name}/node-groups...
	if len(parts) >= 2 && parts[0] == "clusters" {
		return parts[1]
	}

	return ""
}

// extractClusterAndNodegroupName extracts both cluster and nodegroup names from the URL path.
// Expected path: /clusters/{clusterName}/node-groups/{nodegroupName}
func extractClusterAndNodegroupName(path string) (string, string) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// Expected: clusters/{clusterName}/node-groups/{nodegroupName}
	if len(parts) >= 4 && parts[0] == "clusters" && parts[2] == "node-groups" {
		return parts[1], parts[3]
	}

	return "", ""
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := struct {
		Message string `json:"message"`
		Code    string `json:"__type"`
	}{
		Message: message,
		Code:    code,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
