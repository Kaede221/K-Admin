package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: k-admin-system
// Property 1: Unified Response Structure
// For any API endpoint response, the JSON structure SHALL contain exactly three fields:
// code (integer), data (object), and msg (string), where code equals 0 for success
// and non-zero for errors
// Validates: Requirements 1.1, 1.2, 1.3
func TestProperty1_UnifiedResponseStructure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test success responses
	properties.Property("success responses have code=0 and correct structure", prop.ForAll(
		func(dataType int, msg string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Generate different types of data based on dataType
			var data interface{}
			switch dataType % 4 {
			case 0:
				data = nil
			case 1:
				data = "test data"
			case 2:
				data = 123
			case 3:
				data = map[string]string{"key": "value"}
			}

			// Test OkWithDetailed
			OkWithDetailed(c, data, msg)

			// Parse response
			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Logf("Failed to parse response JSON: %v", err)
				return false
			}

			// Verify structure
			if response.Code != 0 {
				t.Logf("Expected code=0 for success, got %d", response.Code)
				return false
			}

			if response.Msg != msg {
				t.Logf("Expected msg=%q, got %q", msg, response.Msg)
				return false
			}

			// Verify HTTP status is 200
			if w.Code != http.StatusOK {
				t.Logf("Expected HTTP status 200, got %d", w.Code)
				return false
			}

			// Verify response has exactly 3 fields by checking JSON structure
			var rawResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &rawResponse)
			if err != nil {
				return false
			}

			if len(rawResponse) != 3 {
				t.Logf("Expected exactly 3 fields, got %d", len(rawResponse))
				return false
			}

			// Verify required fields exist
			if _, ok := rawResponse["code"]; !ok {
				t.Logf("Missing 'code' field")
				return false
			}
			if _, ok := rawResponse["data"]; !ok {
				t.Logf("Missing 'data' field")
				return false
			}
			if _, ok := rawResponse["msg"]; !ok {
				t.Logf("Missing 'msg' field")
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
		gen.AlphaString(),
	))

	// Test failure responses
	properties.Property("failure responses have non-zero code and correct structure", prop.ForAll(
		func(code int, msg string) bool {
			// Ensure code is non-zero
			if code == 0 {
				code = 1
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Test FailWithCode
			FailWithCode(c, code, msg)

			// Parse response
			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Logf("Failed to parse response JSON: %v", err)
				return false
			}

			// Verify structure
			if response.Code == 0 {
				t.Logf("Expected non-zero code for failure, got 0")
				return false
			}

			if response.Code != code {
				t.Logf("Expected code=%d, got %d", code, response.Code)
				return false
			}

			if response.Msg != msg {
				t.Logf("Expected msg=%q, got %q", msg, response.Msg)
				return false
			}

			// Verify data is nil for error responses
			if response.Data != nil {
				t.Logf("Expected data=nil for error response, got %v", response.Data)
				return false
			}

			// Verify HTTP status is still 200 (business logic determines code field)
			if w.Code != http.StatusOK {
				t.Logf("Expected HTTP status 200, got %d", w.Code)
				return false
			}

			// Verify response has exactly 3 fields
			var rawResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &rawResponse)
			if err != nil {
				return false
			}

			if len(rawResponse) != 3 {
				t.Logf("Expected exactly 3 fields, got %d", len(rawResponse))
				return false
			}

			return true
		},
		gen.IntRange(1, 1000),
		gen.AlphaString(),
	))

	// Test Ok helper
	properties.Property("Ok helper returns valid success response", prop.ForAll(
		func() bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Ok(c)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				return false
			}

			return response.Code == 0 &&
				response.Data == nil &&
				response.Msg == "success" &&
				w.Code == http.StatusOK
		},
	))

	// Test OkWithData helper
	properties.Property("OkWithData helper returns valid success response with data", prop.ForAll(
		func(dataType int) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Generate different types of data based on dataType
			var data interface{}
			switch dataType % 5 {
			case 0:
				data = nil
			case 1:
				data = "test"
			case 2:
				data = 42
			case 3:
				data = []int{1, 2, 3}
			case 4:
				data = map[string]int{"a": 1}
			}

			OkWithData(c, data)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				return false
			}

			return response.Code == 0 &&
				response.Msg == "success" &&
				w.Code == http.StatusOK
		},
		gen.IntRange(0, 100),
	))

	// Test Fail helper
	properties.Property("Fail helper returns valid error response", prop.ForAll(
		func(msg string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Fail(c, msg)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				return false
			}

			return response.Code == 1 &&
				response.Data == nil &&
				response.Msg == msg &&
				w.Code == http.StatusOK
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}
