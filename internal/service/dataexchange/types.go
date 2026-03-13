package dataexchange

import "time"

// DataSet represents an AWS Data Exchange data set.
type DataSet struct {
	Arn         string            `json:"Arn"`
	AssetType   string            `json:"AssetType"`
	CreatedAt   time.Time         `json:"CreatedAt"`
	Description string            `json:"Description"`
	ID          string            `json:"Id"`
	Name        string            `json:"Name"`
	Origin      string            `json:"Origin"`
	UpdatedAt   time.Time         `json:"UpdatedAt"`
	SourceID    string            `json:"SourceId,omitempty"`
	Tags        map[string]string `json:"Tags,omitempty"`
}

// Revision represents an AWS Data Exchange revision.
type Revision struct {
	Arn               string    `json:"Arn"`
	Comment           string    `json:"Comment,omitempty"`
	CreatedAt         time.Time `json:"CreatedAt"`
	DataSetID         string    `json:"DataSetId"`
	Finalized         bool      `json:"Finalized"`
	ID                string    `json:"Id"`
	RevocationComment string    `json:"RevocationComment,omitempty"`
	Revoked           bool      `json:"Revoked"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
	SourceID          string    `json:"SourceId,omitempty"`
}

// Job represents an AWS Data Exchange job.
type Job struct {
	Arn       string    `json:"Arn"`
	CreatedAt time.Time `json:"CreatedAt"`
	ID        string    `json:"Id"`
	State     string    `json:"State"`
	Type      string    `json:"Type"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

// CreateDataSetInput represents a CreateDataSet request body.
type CreateDataSetInput struct {
	AssetType   string            `json:"AssetType"`
	Description string            `json:"Description"`
	Name        string            `json:"Name"`
	Tags        map[string]string `json:"Tags,omitempty"`
}

// UpdateDataSetInput represents an UpdateDataSet request body.
type UpdateDataSetInput struct {
	Description string `json:"Description,omitempty"`
	Name        string `json:"Name,omitempty"`
}

// CreateRevisionInput represents a CreateRevision request body.
type CreateRevisionInput struct {
	Comment string            `json:"Comment,omitempty"`
	Tags    map[string]string `json:"Tags,omitempty"`
}

// UpdateRevisionInput represents an UpdateRevision request body.
type UpdateRevisionInput struct {
	Comment   string `json:"Comment,omitempty"`
	Finalized *bool  `json:"Finalized,omitempty"`
}

// CreateJobInput represents a CreateJob request body.
type CreateJobInput struct {
	Type string `json:"Type"`
}

// DataSetsResponse represents a ListDataSets response.
type DataSetsResponse struct {
	DataSets  []DataSet `json:"DataSets"`
	NextToken string    `json:"NextToken,omitempty"`
}

// RevisionsResponse represents a ListDataSetRevisions response.
type RevisionsResponse struct {
	Revisions []Revision `json:"Revisions"`
	NextToken string     `json:"NextToken,omitempty"`
}

// JobsResponse represents a ListJobs response.
type JobsResponse struct {
	Jobs      []Job  `json:"Jobs"`
	NextToken string `json:"NextToken,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"Message"`
	Type    string `json:"Type"`
}
