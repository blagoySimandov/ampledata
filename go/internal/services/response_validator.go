package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/blagoySimandov/ampledata/go/internal/models"
)

func ValidateAndCoerceTypes(
	extractedData map[string]interface{},
	columnsMetadata []*models.ColumnMetadata,
	confidence map[string]*models.FieldConfidenceInfo,
) map[string]interface{} {
	validated := make(map[string]interface{})

	for _, col := range columnsMetadata {
		value, exists := extractedData[col.Name]
		if !exists {
			continue // Field not extracted, skip
		}

		switch col.Type {
		case models.ColumnTypeString:
			validated[col.Name] = coerceToString(value, col.Name, confidence)

		case models.ColumnTypeNumber:
			validated[col.Name] = coerceToNumber(value, col.Name, confidence)

		case models.ColumnTypeBoolean:
			validated[col.Name] = coerceToBoolean(value, col.Name, confidence)

		case models.ColumnTypeDate:
			validated[col.Name] = coerceToDate(value, col.Name, confidence)
		}
	}

	return validated
}

func coerceToString(value interface{}, fieldName string, confidence map[string]*models.FieldConfidenceInfo) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		// Array → string: join elements
		var parts []string
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		result := strings.Join(parts, ", ")

		// Add note
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += fmt.Sprintf(" (Note: Multiple values found and joined: %d items)", len(v))
		}

		return result
	default:
		return fmt.Sprintf("%v", value) // Fallback: convert to string
	}
}

func coerceToNumber(value interface{}, fieldName string, confidence map[string]*models.FieldConfidenceInfo) interface{} {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Try to parse string to number
		var num float64
		_, err := fmt.Sscanf(v, "%f", &num)
		if err == nil {
			return num
		}
		// Try to extract number from string (e.g., "$100.50" → 100.50)
		cleaned := extractNumberFromString(v)
		if cleaned != "" {
			_, err := fmt.Sscanf(cleaned, "%f", &num)
			if err == nil {
				// Add note for extracted numbers
				if conf, exists := confidence[fieldName]; exists {
					conf.Reason += " (Note: Number extracted from string)"
				}
				return num
			}
		}
		// Failed to parse
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Could not coerce string '%s' to number)", v)
		}
		return nil
	case []interface{}:
		// Array → take first numeric element or average
		var numbers []float64
		for _, item := range v {
			if num, ok := item.(float64); ok {
				numbers = append(numbers, num)
			} else if num, ok := item.(int); ok {
				numbers = append(numbers, float64(num))
			}
		}
		if len(numbers) > 0 {
			avg := numbers[0]
			if len(numbers) > 1 {
				sum := 0.0
				for _, n := range numbers {
					sum += n
				}
				avg = sum / float64(len(numbers))
				// Add note for averaged values
				if conf, exists := confidence[fieldName]; exists {
					conf.Reason += fmt.Sprintf(" (Note: Averaged %d numeric values)", len(numbers))
				}
			}
			return avg
		}
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += " (Error: No numeric values found in array)"
		}
		return nil
	case bool:
		// bool → number: true=1, false=0
		if v {
			return 1.0
		}
		return 0.0
	default:
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Cannot coerce type %T to number)", value)
		}
		return nil
	}
}

func extractNumberFromString(s string) string {
	var result strings.Builder
	hasDecimal := false

	for i, char := range s {
		if char >= '0' && char <= '9' {
			result.WriteRune(char)
		} else if char == '.' || char == ',' {
			if !hasDecimal && i > 0 && i < len(s)-1 { // Not at start or end
				hasDecimal = true
				result.WriteRune('.')
			}
		} else if char == '-' && result.Len() == 0 { // Minus sign at start
			result.WriteRune(char)
		}
	}
	return result.String()
}

func coerceToBoolean(value interface{}, fieldName string, confidence map[string]*models.FieldConfidenceInfo) interface{} {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		// Normalize and check common boolean representations
		normalized := strings.ToLower(strings.TrimSpace(v))

		// True values
		if normalized == "true" || normalized == "yes" || normalized == "1" ||
			normalized == "on" || normalized == "y" || normalized == "t" {
			return true
		}

		// False values
		if normalized == "false" || normalized == "no" || normalized == "0" ||
			normalized == "off" || normalized == "n" || normalized == "f" {
			return false
		}

		// Failed to parse
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Could not coerce string '%s' to boolean)", v)
		}
		return nil
	case float64:
		// Numeric → boolean: 0=false, non-zero=true
		result := v != 0
		// Add note for numeric coercion
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += " (Note: Coerced from numeric value)"
		}
		return result
	case int:
		result := v != 0
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += " (Note: Coerced from numeric value)"
		}
		return result
	case int64:
		result := v != 0
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += " (Note: Coerced from numeric value)"
		}
		return result
	case []interface{}:
		// Array → boolean: empty=false, non-empty=true (or first element if single item)
		if len(v) == 0 {
			return false
		}
		if len(v) == 1 {
			// Recursively coerce first element
			return coerceToBoolean(v[0], fieldName, confidence)
		}
		// Multiple elements → true
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += fmt.Sprintf(" (Note: Coerced from array with %d elements)", len(v))
		}
		return true
	default:
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Cannot coerce type %T to boolean)", value)
		}
		return nil
	}
}

func coerceToDate(value interface{}, fieldName string, confidence map[string]*models.FieldConfidenceInfo) interface{} {
	switch v := value.(type) {
	case string:
		// Try multiple date formats
		dateFormats := []string{
			"2006-01-02",                // ISO 8601
			"01/02/2006",                // US format
			"02/01/2006",                // EU format
			"2006-01-02T15:04:05Z07:00", // RFC 3339
			"2006-01-02 15:04:05",       // DateTime
			"January 2, 2006",           // Long format
			"Jan 2, 2006",               // Short format
		}

		for _, format := range dateFormats {
			if parsed, err := time.Parse(format, strings.TrimSpace(v)); err == nil {
				return parsed.Format("2006-01-02") // Return as ISO 8601
			}
		}

		// Failed to parse
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Could not parse date string '%s')", v)
		}
		return nil
	case float64:
		// Unix timestamp
		parsed := time.Unix(int64(v), 0)
		// Add note for timestamp conversion
		if conf, exists := confidence[fieldName]; exists {
			conf.Reason += " (Note: Coerced from Unix timestamp)"
		}
		return parsed.Format("2006-01-02")
	case []interface{}:
		// Array → take first date element
		for _, item := range v {
			result := coerceToDate(item, fieldName, confidence)
			if result != nil {
				if conf, exists := confidence[fieldName]; exists {
					conf.Reason += " (Note: Date extracted from array)"
				}
				return result
			}
		}
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += " (Error: No valid date found in array)"
		}
		return nil
	default:
		if conf, exists := confidence[fieldName]; exists {
			conf.Score = 0.0
			conf.Reason += fmt.Sprintf(" (Error: Cannot coerce type %T to date)", value)
		}
		return nil
	}
}
