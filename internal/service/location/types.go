package location

import "time"

// MapConfiguration represents the configuration for a map resource.
type MapConfiguration struct {
	Style string
}

// MapResource represents an Amazon Location Service map resource.
type MapResource struct {
	Name          string
	ARN           string
	Description   string
	Configuration MapConfiguration
	PricingPlan   string
	Tags          map[string]string
	CreateTime    time.Time
	UpdateTime    time.Time
}

// DataSourceConfiguration represents the data source configuration for a place index.
type DataSourceConfiguration struct {
	IntendedUse string
}

// PlaceIndex represents an Amazon Location Service place index resource.
type PlaceIndex struct {
	IndexName               string
	ARN                     string
	Description             string
	DataSource              string
	DataSourceConfiguration DataSourceConfiguration
	PricingPlan             string
	Tags                    map[string]string
	CreateTime              time.Time
	UpdateTime              time.Time
}

// RouteCalculator represents an Amazon Location Service route calculator resource.
type RouteCalculator struct {
	CalculatorName string
	ARN            string
	Description    string
	DataSource     string
	PricingPlan    string
	Tags           map[string]string
	CreateTime     time.Time
	UpdateTime     time.Time
}

// GeofenceCollection represents an Amazon Location Service geofence collection resource.
type GeofenceCollection struct {
	CollectionName string
	ARN            string
	Description    string
	PricingPlan    string
	Tags           map[string]string
	CreateTime     time.Time
	UpdateTime     time.Time
}

// Tracker represents an Amazon Location Service tracker resource.
type Tracker struct {
	TrackerName       string
	ARN               string
	Description       string
	PricingPlan       string
	PositionFiltering string
	Tags              map[string]string
	CreateTime        time.Time
	UpdateTime        time.Time
}

// Error represents a Location service error.
type Error struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

// --- Maps ---

// MapConfigurationInput represents the configuration input for creating a map.
type MapConfigurationInput struct {
	Style string `json:"Style"`
}

// CreateMapRequest represents the CreateMap API request.
type CreateMapRequest struct {
	MapName       string                `json:"MapName"`
	Configuration MapConfigurationInput `json:"Configuration"`
	Description   string                `json:"Description,omitempty"`
	PricingPlan   string                `json:"PricingPlan,omitempty"`
	Tags          map[string]string     `json:"Tags,omitempty"`
}

// CreateMapResponse represents the CreateMap API response.
type CreateMapResponse struct {
	MapName    string    `json:"MapName"`
	MapArn     string    `json:"MapArn"`
	CreateTime time.Time `json:"CreateTime"`
}

// MapConfigurationOutput represents the configuration output for a map.
type MapConfigurationOutput struct {
	Style string `json:"Style"`
}

// DescribeMapResponse represents the DescribeMap API response.
type DescribeMapResponse struct {
	MapName       string                 `json:"MapName"`
	MapArn        string                 `json:"MapArn"`
	Configuration MapConfigurationOutput `json:"Configuration"`
	Description   string                 `json:"Description,omitempty"`
	PricingPlan   string                 `json:"PricingPlan,omitempty"`
	Tags          map[string]string      `json:"Tags,omitempty"`
	CreateTime    time.Time              `json:"CreateTime"`
	UpdateTime    time.Time              `json:"UpdateTime"`
}

// UpdateMapRequest represents the UpdateMap API request.
type UpdateMapRequest struct {
	Description string `json:"Description,omitempty"`
	PricingPlan string `json:"PricingPlan,omitempty"`
}

// UpdateMapResponse represents the UpdateMap API response.
type UpdateMapResponse struct {
	MapName    string    `json:"MapName"`
	MapArn     string    `json:"MapArn"`
	UpdateTime time.Time `json:"UpdateTime"`
}

// ListMapsRequest represents the ListMaps API request.
type ListMapsRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListMapsEntry represents a single entry in the ListMaps response.
type ListMapsEntry struct {
	MapName     string    `json:"MapName"`
	Description string    `json:"Description,omitempty"`
	DataSource  string    `json:"DataSource,omitempty"`
	CreateTime  time.Time `json:"CreateTime"`
	UpdateTime  time.Time `json:"UpdateTime"`
}

// ListMapsResponse represents the ListMaps API response.
type ListMapsResponse struct {
	Entries   []ListMapsEntry `json:"Entries"`
	NextToken string          `json:"NextToken,omitempty"`
}

// --- Place Indexes ---

// DataSourceConfigurationInput represents the data source configuration input.
type DataSourceConfigurationInput struct {
	IntendedUse string `json:"IntendedUse,omitempty"`
}

// CreatePlaceIndexRequest represents the CreatePlaceIndex API request.
type CreatePlaceIndexRequest struct {
	IndexName               string                        `json:"IndexName"`
	DataSource              string                        `json:"DataSource"`
	DataSourceConfiguration *DataSourceConfigurationInput `json:"DataSourceConfiguration,omitempty"`
	Description             string                        `json:"Description,omitempty"`
	PricingPlan             string                        `json:"PricingPlan,omitempty"`
	Tags                    map[string]string             `json:"Tags,omitempty"`
}

// CreatePlaceIndexResponse represents the CreatePlaceIndex API response.
type CreatePlaceIndexResponse struct {
	IndexName  string    `json:"IndexName"`
	IndexArn   string    `json:"IndexArn"`
	CreateTime time.Time `json:"CreateTime"`
}

// DataSourceConfigurationOutput represents the data source configuration output.
type DataSourceConfigurationOutput struct {
	IntendedUse string `json:"IntendedUse,omitempty"`
}

// DescribePlaceIndexResponse represents the DescribePlaceIndex API response.
type DescribePlaceIndexResponse struct {
	IndexName               string                        `json:"IndexName"`
	IndexArn                string                        `json:"IndexArn"`
	DataSource              string                        `json:"DataSource"`
	DataSourceConfiguration DataSourceConfigurationOutput `json:"DataSourceConfiguration"`
	Description             string                        `json:"Description,omitempty"`
	PricingPlan             string                        `json:"PricingPlan,omitempty"`
	Tags                    map[string]string             `json:"Tags,omitempty"`
	CreateTime              time.Time                     `json:"CreateTime"`
	UpdateTime              time.Time                     `json:"UpdateTime"`
}

// UpdatePlaceIndexRequest represents the UpdatePlaceIndex API request.
type UpdatePlaceIndexRequest struct {
	DataSourceConfiguration *DataSourceConfigurationInput `json:"DataSourceConfiguration,omitempty"`
	Description             string                        `json:"Description,omitempty"`
	PricingPlan             string                        `json:"PricingPlan,omitempty"`
}

// UpdatePlaceIndexResponse represents the UpdatePlaceIndex API response.
type UpdatePlaceIndexResponse struct {
	IndexName  string    `json:"IndexName"`
	IndexArn   string    `json:"IndexArn"`
	UpdateTime time.Time `json:"UpdateTime"`
}

// ListPlaceIndexesRequest represents the ListPlaceIndexes API request.
type ListPlaceIndexesRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListPlaceIndexesEntry represents a single entry in the ListPlaceIndexes response.
type ListPlaceIndexesEntry struct {
	IndexName   string    `json:"IndexName"`
	Description string    `json:"Description,omitempty"`
	DataSource  string    `json:"DataSource"`
	CreateTime  time.Time `json:"CreateTime"`
	UpdateTime  time.Time `json:"UpdateTime"`
}

// ListPlaceIndexesResponse represents the ListPlaceIndexes API response.
type ListPlaceIndexesResponse struct {
	Entries   []ListPlaceIndexesEntry `json:"Entries"`
	NextToken string                  `json:"NextToken,omitempty"`
}

// --- Route Calculators ---

// CreateRouteCalculatorRequest represents the CreateRouteCalculator API request.
type CreateRouteCalculatorRequest struct {
	CalculatorName string            `json:"CalculatorName"`
	DataSource     string            `json:"DataSource"`
	Description    string            `json:"Description,omitempty"`
	PricingPlan    string            `json:"PricingPlan,omitempty"`
	Tags           map[string]string `json:"Tags,omitempty"`
}

// CreateRouteCalculatorResponse represents the CreateRouteCalculator API response.
type CreateRouteCalculatorResponse struct {
	CalculatorName string    `json:"CalculatorName"`
	CalculatorArn  string    `json:"CalculatorArn"`
	CreateTime     time.Time `json:"CreateTime"`
}

// DescribeRouteCalculatorResponse represents the DescribeRouteCalculator API response.
type DescribeRouteCalculatorResponse struct {
	CalculatorName string            `json:"CalculatorName"`
	CalculatorArn  string            `json:"CalculatorArn"`
	DataSource     string            `json:"DataSource"`
	Description    string            `json:"Description,omitempty"`
	PricingPlan    string            `json:"PricingPlan,omitempty"`
	Tags           map[string]string `json:"Tags,omitempty"`
	CreateTime     time.Time         `json:"CreateTime"`
	UpdateTime     time.Time         `json:"UpdateTime"`
}

// UpdateRouteCalculatorRequest represents the UpdateRouteCalculator API request.
type UpdateRouteCalculatorRequest struct {
	Description string `json:"Description,omitempty"`
	PricingPlan string `json:"PricingPlan,omitempty"`
}

// UpdateRouteCalculatorResponse represents the UpdateRouteCalculator API response.
type UpdateRouteCalculatorResponse struct {
	CalculatorName string    `json:"CalculatorName"`
	CalculatorArn  string    `json:"CalculatorArn"`
	UpdateTime     time.Time `json:"UpdateTime"`
}

// ListRouteCalculatorsRequest represents the ListRouteCalculators API request.
type ListRouteCalculatorsRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListRouteCalculatorsEntry represents a single entry in the ListRouteCalculators response.
type ListRouteCalculatorsEntry struct {
	CalculatorName string    `json:"CalculatorName"`
	Description    string    `json:"Description,omitempty"`
	DataSource     string    `json:"DataSource"`
	CreateTime     time.Time `json:"CreateTime"`
	UpdateTime     time.Time `json:"UpdateTime"`
}

// ListRouteCalculatorsResponse represents the ListRouteCalculators API response.
type ListRouteCalculatorsResponse struct {
	Entries   []ListRouteCalculatorsEntry `json:"Entries"`
	NextToken string                      `json:"NextToken,omitempty"`
}

// --- Geofence Collections ---

// CreateGeofenceCollectionRequest represents the CreateGeofenceCollection API request.
type CreateGeofenceCollectionRequest struct {
	CollectionName string            `json:"CollectionName"`
	Description    string            `json:"Description,omitempty"`
	PricingPlan    string            `json:"PricingPlan,omitempty"`
	Tags           map[string]string `json:"Tags,omitempty"`
}

// CreateGeofenceCollectionResponse represents the CreateGeofenceCollection API response.
type CreateGeofenceCollectionResponse struct {
	CollectionName string    `json:"CollectionName"`
	CollectionArn  string    `json:"CollectionArn"`
	CreateTime     time.Time `json:"CreateTime"`
}

// DescribeGeofenceCollectionResponse represents the DescribeGeofenceCollection API response.
type DescribeGeofenceCollectionResponse struct {
	CollectionName string            `json:"CollectionName"`
	CollectionArn  string            `json:"CollectionArn"`
	Description    string            `json:"Description,omitempty"`
	PricingPlan    string            `json:"PricingPlan,omitempty"`
	Tags           map[string]string `json:"Tags,omitempty"`
	CreateTime     time.Time         `json:"CreateTime"`
	UpdateTime     time.Time         `json:"UpdateTime"`
}

// UpdateGeofenceCollectionRequest represents the UpdateGeofenceCollection API request.
type UpdateGeofenceCollectionRequest struct {
	Description string `json:"Description,omitempty"`
	PricingPlan string `json:"PricingPlan,omitempty"`
}

// UpdateGeofenceCollectionResponse represents the UpdateGeofenceCollection API response.
type UpdateGeofenceCollectionResponse struct {
	CollectionName string    `json:"CollectionName"`
	CollectionArn  string    `json:"CollectionArn"`
	UpdateTime     time.Time `json:"UpdateTime"`
}

// ListGeofenceCollectionsRequest represents the ListGeofenceCollections API request.
type ListGeofenceCollectionsRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListGeofenceCollectionsEntry represents a single entry in the ListGeofenceCollections response.
type ListGeofenceCollectionsEntry struct {
	CollectionName string    `json:"CollectionName"`
	Description    string    `json:"Description,omitempty"`
	CreateTime     time.Time `json:"CreateTime"`
	UpdateTime     time.Time `json:"UpdateTime"`
}

// ListGeofenceCollectionsResponse represents the ListGeofenceCollections API response.
type ListGeofenceCollectionsResponse struct {
	Entries   []ListGeofenceCollectionsEntry `json:"Entries"`
	NextToken string                         `json:"NextToken,omitempty"`
}

// --- Trackers ---

// CreateTrackerRequest represents the CreateTracker API request.
type CreateTrackerRequest struct {
	TrackerName       string            `json:"TrackerName"`
	Description       string            `json:"Description,omitempty"`
	PricingPlan       string            `json:"PricingPlan,omitempty"`
	PositionFiltering string            `json:"PositionFiltering,omitempty"`
	Tags              map[string]string `json:"Tags,omitempty"`
}

// CreateTrackerResponse represents the CreateTracker API response.
type CreateTrackerResponse struct {
	TrackerName string    `json:"TrackerName"`
	TrackerArn  string    `json:"TrackerArn"`
	CreateTime  time.Time `json:"CreateTime"`
}

// DescribeTrackerResponse represents the DescribeTracker API response.
type DescribeTrackerResponse struct {
	TrackerName       string            `json:"TrackerName"`
	TrackerArn        string            `json:"TrackerArn"`
	Description       string            `json:"Description,omitempty"`
	PricingPlan       string            `json:"PricingPlan,omitempty"`
	PositionFiltering string            `json:"PositionFiltering,omitempty"`
	Tags              map[string]string `json:"Tags,omitempty"`
	CreateTime        time.Time         `json:"CreateTime"`
	UpdateTime        time.Time         `json:"UpdateTime"`
}

// UpdateTrackerRequest represents the UpdateTracker API request.
type UpdateTrackerRequest struct {
	Description       string `json:"Description,omitempty"`
	PricingPlan       string `json:"PricingPlan,omitempty"`
	PositionFiltering string `json:"PositionFiltering,omitempty"`
}

// UpdateTrackerResponse represents the UpdateTracker API response.
type UpdateTrackerResponse struct {
	TrackerName string    `json:"TrackerName"`
	TrackerArn  string    `json:"TrackerArn"`
	UpdateTime  time.Time `json:"UpdateTime"`
}

// ListTrackersRequest represents the ListTrackers API request.
type ListTrackersRequest struct {
	MaxResults *int32 `json:"MaxResults,omitempty"`
	NextToken  string `json:"NextToken,omitempty"`
}

// ListTrackersEntry represents a single entry in the ListTrackers response.
type ListTrackersEntry struct {
	TrackerName string    `json:"TrackerName"`
	Description string    `json:"Description,omitempty"`
	CreateTime  time.Time `json:"CreateTime"`
	UpdateTime  time.Time `json:"UpdateTime"`
}

// ListTrackersResponse represents the ListTrackers API response.
type ListTrackersResponse struct {
	Entries   []ListTrackersEntry `json:"Entries"`
	NextToken string              `json:"NextToken,omitempty"`
}

// --- Error Response ---

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Type    string `json:"__type"`
	Message string `json:"message"`
}
