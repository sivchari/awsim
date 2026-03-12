package glacier

// Vault represents a Glacier vault.
type Vault struct {
	CreationDate      string `json:"CreationDate,omitempty"`
	LastInventoryDate string `json:"LastInventoryDate,omitempty"`
	NumberOfArchives  int64  `json:"NumberOfArchives"`
	SizeInBytes       int64  `json:"SizeInBytes"`
	VaultARN          string `json:"VaultARN,omitempty"`
	VaultName         string `json:"VaultName,omitempty"`
}

// ListVaultsResponse represents a ListVaults response.
type ListVaultsResponse struct {
	Marker    string  `json:"Marker,omitempty"`
	VaultList []Vault `json:"VaultList"`
}

// Archive represents a Glacier archive.
type Archive struct {
	ArchiveID          string `json:"ArchiveId,omitempty"`
	ArchiveDescription string `json:"ArchiveDescription,omitempty"`
	CreationDate       string `json:"CreationDate,omitempty"`
	SHA256TreeHash     string `json:"SHA256TreeHash,omitempty"`
	Size               int64  `json:"Size"`
	VaultARN           string `json:"VaultARN,omitempty"`
}

// UploadArchiveResponse represents an UploadArchive response.
type UploadArchiveResponse struct {
	ArchiveID      string `json:"archiveId,omitempty"`
	Checksum       string `json:"checksum,omitempty"`
	Location       string `json:"location,omitempty"`
	SHA256TreeHash string `json:"x-amz-sha256-tree-hash,omitempty"`
}

// ErrorResponse represents a Glacier error response.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
