package ebs

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// Block represents a snapshot block.
type Block struct {
	BlockIndex int32  `json:"BlockIndex"`
	BlockToken string `json:"BlockToken"`
}

// Snapshot represents an EBS snapshot.
type Snapshot struct {
	BlockSize        int32  `json:"BlockSize,omitempty"`
	Description      string `json:"Description,omitempty"`
	KmsKeyArn        string `json:"KmsKeyArn,omitempty"`
	OwnerID          string `json:"OwnerId,omitempty"`
	ParentSnapshotID string `json:"ParentSnapshotId,omitempty"`
	SnapshotID       string `json:"SnapshotId,omitempty"`
	StartTime        int64  `json:"StartTime,omitempty"`
	Status           string `json:"Status,omitempty"`
	Tags             []Tag  `json:"Tags,omitempty"`
	VolumeSize       int64  `json:"VolumeSize,omitempty"`
}

// StartSnapshotRequest represents a StartSnapshot request.
type StartSnapshotRequest struct {
	ClientToken      string `json:"ClientToken,omitempty"`
	Description      string `json:"Description,omitempty"`
	Encrypted        *bool  `json:"Encrypted,omitempty"`
	KmsKeyArn        string `json:"KmsKeyArn,omitempty"`
	ParentSnapshotID string `json:"ParentSnapshotId,omitempty"`
	Tags             []Tag  `json:"Tags,omitempty"`
	Timeout          int32  `json:"Timeout,omitempty"`
	VolumeSize       int64  `json:"VolumeSize"`
}

// CompleteSnapshotRequest represents a CompleteSnapshot request.
type CompleteSnapshotRequest struct {
	ChangedBlocksCount int32 `json:"ChangedBlocksCount"`
}

// ListSnapshotBlocksResponse represents a ListSnapshotBlocks response.
type ListSnapshotBlocksResponse struct {
	Blocks     []Block `json:"Blocks"`
	BlockSize  int32   `json:"BlockSize,omitempty"`
	ExpiryTime int64   `json:"ExpiryTime,omitempty"`
	NextToken  string  `json:"NextToken,omitempty"`
	VolumeSize int64   `json:"VolumeSize,omitempty"`
}

// PutSnapshotBlockResponse represents a PutSnapshotBlock response.
type PutSnapshotBlockResponse struct {
	Checksum          string `json:"Checksum"`
	ChecksumAlgorithm string `json:"ChecksumAlgorithm"`
}

// ChangedBlock represents a changed block between two snapshots.
type ChangedBlock struct {
	BlockIndex       int32  `json:"BlockIndex"`
	FirstBlockToken  string `json:"FirstBlockToken,omitempty"`
	SecondBlockToken string `json:"SecondBlockToken,omitempty"`
}

// ListChangedBlocksResponse represents a ListChangedBlocks response.
type ListChangedBlocksResponse struct {
	BlockSize     int32          `json:"BlockSize,omitempty"`
	ChangedBlocks []ChangedBlock `json:"ChangedBlocks"`
	ExpiryTime    int64          `json:"ExpiryTime,omitempty"`
	NextToken     string         `json:"NextToken,omitempty"`
	VolumeSize    int64          `json:"VolumeSize,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"Message"`
	Reason  string `json:"Reason,omitempty"`
}
