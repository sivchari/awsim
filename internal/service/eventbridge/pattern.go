package eventbridge

import (
	"encoding/json"
	"strings"
)

// matchEventPattern checks if an event matches an EventPattern.
// EventPattern is a JSON object where each key maps to an array of acceptable values.
// An event matches if ALL keys in the pattern match.
// For each key, the event's value must match at least one value in the pattern array.
func matchEventPattern(patternJSON string, event PutEventsRequestEntry) bool {
	if patternJSON == "" {
		return true
	}

	var pattern map[string]json.RawMessage
	if err := json.Unmarshal([]byte(patternJSON), &pattern); err != nil {
		return false
	}

	for key, rawValues := range pattern {
		switch key {
		case "source":
			if !matchStringField(rawValues, event.Source) {
				return false
			}
		case "detail-type":
			if !matchStringField(rawValues, event.DetailType) {
				return false
			}
		case "detail":
			if !matchDetailField(rawValues, event.Detail) {
				return false
			}
		}
	}

	return true
}

// matchStringField checks if a field value matches any of the pattern values.
// Pattern values can be a JSON array of strings: ["value1", "value2"].
func matchStringField(rawValues json.RawMessage, fieldValue string) bool {
	var values []string
	if err := json.Unmarshal(rawValues, &values); err != nil {
		return false
	}

	for _, v := range values {
		if v == fieldValue {
			return true
		}
	}

	return false
}

// matchDetailField checks if an event detail matches the detail pattern.
// The detail pattern is a nested JSON object with arrays of acceptable values.
func matchDetailField(rawPattern json.RawMessage, detailJSON string) bool {
	if detailJSON == "" {
		return false
	}

	var pattern map[string]json.RawMessage
	if err := json.Unmarshal(rawPattern, &pattern); err != nil {
		return false
	}

	var detail map[string]json.RawMessage
	if err := json.Unmarshal([]byte(detailJSON), &detail); err != nil {
		return false
	}

	for key, patternValue := range pattern {
		detailValue, exists := detail[key]
		if !exists {
			return false
		}

		if !matchDetailValue(patternValue, detailValue) {
			return false
		}
	}

	return true
}

// matchDetailValue checks if a detail value matches a pattern value.
// Handles both arrays of primitive values and nested objects.
func matchDetailValue(patternValue, detailValue json.RawMessage) bool {
	// Try as array of strings first.
	var strValues []string
	if err := json.Unmarshal(patternValue, &strValues); err == nil {
		var actual string
		if err := json.Unmarshal(detailValue, &actual); err == nil {
			for _, v := range strValues {
				if v == actual {
					return true
				}
			}

			return false
		}
	}

	// Try as array of numbers.
	var numValues []float64
	if err := json.Unmarshal(patternValue, &numValues); err == nil {
		var actual float64
		if err := json.Unmarshal(detailValue, &actual); err == nil {
			for _, v := range numValues {
				if v == actual {
					return true
				}
			}

			return false
		}
	}

	// Try as nested object pattern.
	patternStr := strings.TrimSpace(string(patternValue))
	if strings.HasPrefix(patternStr, "{") {
		return matchDetailField(patternValue, string(detailValue))
	}

	return false
}
