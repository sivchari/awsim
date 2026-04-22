package eventbridge

import (
	"testing"
)

func TestMatchEventPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		event   PutEventsRequestEntry
		want    bool
	}{
		{
			name:    "empty pattern matches everything",
			pattern: "",
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    true,
		},
		{
			name:    "source match",
			pattern: `{"source": ["my.app"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    true,
		},
		{
			name:    "source mismatch",
			pattern: `{"source": ["other.app"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    false,
		},
		{
			name:    "source multiple values",
			pattern: `{"source": ["app.a", "app.b", "my.app"]}`,
			event:   PutEventsRequestEntry{Source: "my.app"},
			want:    true,
		},
		{
			name:    "detail-type match",
			pattern: `{"detail-type": ["OrderCreated"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    true,
		},
		{
			name:    "detail-type mismatch",
			pattern: `{"detail-type": ["OrderDeleted"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    false,
		},
		{
			name:    "source AND detail-type both match",
			pattern: `{"source": ["my.app"], "detail-type": ["OrderCreated"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    true,
		},
		{
			name:    "source matches but detail-type does not",
			pattern: `{"source": ["my.app"], "detail-type": ["OrderDeleted"]}`,
			event:   PutEventsRequestEntry{Source: "my.app", DetailType: "OrderCreated"},
			want:    false,
		},
		{
			name:    "detail field match",
			pattern: `{"source": ["my.app"], "detail": {"status": ["completed"]}}`,
			event:   PutEventsRequestEntry{Source: "my.app", Detail: `{"status": "completed", "amount": 100}`},
			want:    true,
		},
		{
			name:    "detail field mismatch",
			pattern: `{"detail": {"status": ["pending"]}}`,
			event:   PutEventsRequestEntry{Source: "my.app", Detail: `{"status": "completed"}`},
			want:    false,
		},
		{
			name:    "detail nested object match",
			pattern: `{"detail": {"order": {"type": ["premium"]}}}`,
			event:   PutEventsRequestEntry{Detail: `{"order": {"type": "premium", "id": "123"}}`},
			want:    true,
		},
		{
			name:    "detail number match",
			pattern: `{"detail": {"count": [1, 2, 3]}}`,
			event:   PutEventsRequestEntry{Detail: `{"count": 2}`},
			want:    true,
		},
		{
			name:    "invalid pattern JSON",
			pattern: `{invalid}`,
			event:   PutEventsRequestEntry{Source: "my.app"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := matchEventPattern(tt.pattern, tt.event)
			if got != tt.want {
				t.Errorf("matchEventPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}
