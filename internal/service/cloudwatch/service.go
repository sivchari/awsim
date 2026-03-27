package cloudwatch

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sivchari/kumo/internal/server"
	"github.com/sivchari/kumo/internal/service"
)

// Compile-time check that Service implements io.Closer.
var _ io.Closer = (*Service)(nil)

func init() {
	var opts []Option
	if dir := os.Getenv("KUMO_DATA_DIR"); dir != "" {
		opts = append(opts, WithDataDir(dir))
	}

	storage := NewMemoryStorage("", opts...)
	service.Register(New(storage))
}

// Service implements the CloudWatch service.
type Service struct {
	storage Storage
}

// New creates a new CloudWatch service.
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return "monitoring"
}

// RegisterRoutes registers routes with the router.
// CloudWatch uses CBOR protocol, so routes are registered via DispatchCBORAction.
func (s *Service) RegisterRoutes(_ service.Router) {
	// CloudWatch uses RPC v2 CBOR protocol, routing is handled by DispatchCBORAction.
}

// ServiceName returns the Smithy service name for RPC v2 CBOR protocol.
func (s *Service) ServiceName() string {
	return "GraniteServiceVersion20100801"
}

// CBORProtocol is a marker method that indicates CloudWatch uses RPC v2 CBOR protocol.
func (s *Service) CBORProtocol() {}

// DispatchCBORAction handles RPC v2 CBOR protocol requests.
func (s *Service) DispatchCBORAction(w http.ResponseWriter, r *http.Request, operation string) {
	switch operation {
	case "PutMetricData":
		s.PutMetricDataCBOR(w, r)
	case "GetMetricData":
		s.GetMetricDataCBOR(w, r)
	case "GetMetricStatistics":
		s.GetMetricStatisticsCBOR(w, r)
	case "ListMetrics":
		s.ListMetricsCBOR(w, r)
	case "PutMetricAlarm":
		s.PutMetricAlarmCBOR(w, r)
	case "DeleteAlarms":
		s.DeleteAlarmsCBOR(w, r)
	case "DescribeAlarms":
		s.DescribeAlarmsCBOR(w, r)
	default:
		server.WriteCBORError(w, "InvalidAction", "The action "+operation+" is not valid", http.StatusBadRequest)
	}
}

// Close saves the storage state if persistence is enabled.
func (s *Service) Close() error {
	if c, ok := s.storage.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close storage: %w", err)
		}
	}

	return nil
}
