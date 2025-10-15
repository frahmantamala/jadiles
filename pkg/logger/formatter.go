package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type ValueObfuscator func(interface{}) interface{}

type KeyMatcher interface {
	Match(key string) (ValueObfuscator, bool)
}

type exactKeyMatcher struct {
	matchers map[string]ValueObfuscator
}

func (m *exactKeyMatcher) Match(key string) (ValueObfuscator, bool) {
	obfuscator, exists := m.matchers[key]
	return obfuscator, exists
}

func KeyMatcherExact(matchers map[string]ValueObfuscator) KeyMatcher {
	return &exactKeyMatcher{
		matchers: matchers,
	}
}

func ValueSHA256Hashed() ValueObfuscator {
	return func(value interface{}) interface{} {
		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case nil:
			return nil
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		// Hash the value
		hash := sha256.Sum256([]byte(strValue))
		return hex.EncodeToString(hash[:])
	}
}

func ValueToFixedValue(replacement string) ValueObfuscator {
	return func(_ interface{}) interface{} {
		return replacement
	}
}

func HideSensitiveValueOnKeys(matcher KeyMatcher) func(data interface{}) interface{} {
	return func(data interface{}) interface{} {
		return obfuscateWithMatcher(data, matcher)
	}
}

func obfuscateWithMatcher(data interface{}, matcher KeyMatcher) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if obfuscator, matches := matcher.Match(key); matches {
				result[key] = obfuscator(value)
			} else {
				result[key] = obfuscateWithMatcher(value, matcher)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = obfuscateWithMatcher(item, matcher)
		}
		return result
	default:
		return v
	}
}

func ByteDecoderJSONObfuscator(matcher KeyMatcher) func([]byte) string {
	obfuscator := HideSensitiveValueOnKeys(matcher)

	return func(data []byte) string {
		if len(data) == 0 {
			return ""
		}

		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return string(data)
		}

		obfuscated := obfuscator(jsonData)

		result, err := json.Marshal(obfuscated)
		if err != nil {
			return string(data)
		}

		return string(result)
	}
}
