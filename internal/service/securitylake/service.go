package securitylake

import "github.com/sivchari/awsim/internal/service"

// Compile-time check to ensure Service implements service.Service.
var _ service.Service = (*Service)(nil)

func init() {
	service.Register(New(NewMemoryStorage()))
}

// Service implements the Security Lake service.
type Service struct {
	storage Storage
}

// New creates a new Security Lake service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "securitylake"
}

// RegisterRoutes registers the Security Lake routes.
func (s *Service) RegisterRoutes(r service.Router) {
	// Data Lake operations
	r.Handle("POST", "/v1/datalake", s.CreateDataLake)
	r.Handle("POST", "/v1/datalake/delete", s.DeleteDataLake)
	r.Handle("GET", "/v1/datalakes", s.ListDataLakes)
	r.Handle("PUT", "/v1/datalake", s.UpdateDataLake)

	// Subscriber operations
	r.Handle("POST", "/v1/subscribers", s.CreateSubscriber)
	r.Handle("GET", "/v1/subscribers/{subscriberId}", s.GetSubscriber)
	r.Handle("DELETE", "/v1/subscribers/{subscriberId}", s.DeleteSubscriber)
	r.Handle("PUT", "/v1/subscribers/{subscriberId}", s.UpdateSubscriber)
	r.Handle("GET", "/v1/subscribers", s.ListSubscribers)

	// Log Source operations
	r.Handle("POST", "/v1/datalake/logsources/aws", s.CreateAwsLogSource)
	r.Handle("POST", "/v1/datalake/logsources/aws/delete", s.DeleteAwsLogSource)
	r.Handle("POST", "/v1/datalake/logsources/list", s.ListLogSources)

	// Tag operations
	r.Handle("POST", "/v1/tags/{resourceArn}", s.TagResource)
	r.Handle("DELETE", "/v1/tags/{resourceArn}", s.UntagResource)
	r.Handle("GET", "/v1/tags/{resourceArn}", s.ListTagsForResource)
}

// Prefix returns the URL prefix for Security Lake.
func (s *Service) Prefix() string {
	return "/securitylake"
}
