package amplify

// App represents an Amplify app.
type App struct {
	AppArn                string            `json:"appArn"`
	AppID                 string            `json:"appId"`
	CreateTime            float64           `json:"createTime"`
	DefaultDomain         string            `json:"defaultDomain"`
	Description           string            `json:"description"`
	EnableBasicAuth       bool              `json:"enableBasicAuth"`
	EnableBranchAutoBuild bool              `json:"enableBranchAutoBuild"`
	EnvironmentVariables  map[string]string `json:"environmentVariables"`
	Name                  string            `json:"name"`
	Platform              string            `json:"platform"`
	Repository            string            `json:"repository"`
	UpdateTime            float64           `json:"updateTime"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// Branch represents an Amplify branch.
type Branch struct {
	ActiveJobID              string            `json:"activeJobId"`
	BranchArn                string            `json:"branchArn"`
	BranchName               string            `json:"branchName"`
	CreateTime               float64           `json:"createTime"`
	CustomDomains            []string          `json:"customDomains"`
	Description              string            `json:"description"`
	DisplayName              string            `json:"displayName"`
	EnableAutoBuild          bool              `json:"enableAutoBuild"`
	EnableNotification       bool              `json:"enableNotification"`
	EnablePullRequestPreview bool              `json:"enablePullRequestPreview"`
	EnvironmentVariables     map[string]string `json:"environmentVariables"`
	Framework                string            `json:"framework"`
	Stage                    string            `json:"stage"`
	TTL                      string            `json:"ttl"`
	TotalNumberOfJobs        string            `json:"totalNumberOfJobs"`
	UpdateTime               float64           `json:"updateTime"`
	Tags                     map[string]string `json:"tags,omitempty"`
}

// CreateAppInput represents a CreateApp request.
type CreateAppInput struct {
	Name                  string            `json:"name"`
	Description           string            `json:"description,omitempty"`
	Repository            string            `json:"repository,omitempty"`
	Platform              string            `json:"platform,omitempty"`
	EnableBasicAuth       *bool             `json:"enableBasicAuth,omitempty"`
	EnableBranchAutoBuild *bool             `json:"enableBranchAutoBuild,omitempty"`
	EnvironmentVariables  map[string]string `json:"environmentVariables,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// UpdateAppInput represents an UpdateApp request.
type UpdateAppInput struct {
	Name                  string            `json:"name,omitempty"`
	Description           string            `json:"description,omitempty"`
	Platform              string            `json:"platform,omitempty"`
	EnableBasicAuth       *bool             `json:"enableBasicAuth,omitempty"`
	EnableBranchAutoBuild *bool             `json:"enableBranchAutoBuild,omitempty"`
	EnvironmentVariables  map[string]string `json:"environmentVariables,omitempty"`
}

// CreateBranchInput represents a CreateBranch request.
type CreateBranchInput struct {
	BranchName           string            `json:"branchName"`
	Description          string            `json:"description,omitempty"`
	Framework            string            `json:"framework,omitempty"`
	Stage                string            `json:"stage,omitempty"`
	EnableAutoBuild      *bool             `json:"enableAutoBuild,omitempty"`
	EnableNotification   *bool             `json:"enableNotification,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
}

// AppResponse represents a response containing an app.
type AppResponse struct {
	App *App `json:"app"`
}

// AppsResponse represents a response containing a list of apps.
type AppsResponse struct {
	Apps []App `json:"apps"`
}

// BranchResponse represents a response containing a branch.
type BranchResponse struct {
	Branch *Branch `json:"branch"`
}

// BranchesResponse represents a response containing a list of branches.
type BranchesResponse struct {
	Branches []Branch `json:"branches"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message"`
}
