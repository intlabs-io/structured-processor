package routes

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lazy-lagoon/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

// Mock storage for testing
var mockStoredPages = make(map[string][]byte)

// Mock storage.StoreBytes function
func mockStoreBytes(output types.Output, content []byte) error {
	key := fmt.Sprintf("%s/%s", output.Reference.Bucket, output.Reference.Prefix)
	mockStoredPages[key] = content
	return nil
}

// Test helper to create request body
func createRequestBody(storageType, dataType, bucket, prefix string) types.RequestBodyPaginate {
	return types.RequestBodyPaginate{
		Input: types.Input{
			StorageType: storageType,
			DataType:    dataType,
			Reference: types.SourceReference{
				Bucket: bucket,
				Prefix: prefix,
				Region: "us-east-1",
			},
			Credential: types.SourceCredential{
				Secrets: types.Secrets{
					Secret: "test-secret",
				},
				Resources: types.Resources{
					Id: "test-id",
				},
			},
		},
		Output: types.Output{
			StorageType: storageType,
			DataType:    dataType,
			Reference: types.SourceReference{
				Bucket: "output-bucket",
				Prefix: "output/test",
				Region: "us-east-1",
			},
			Credential: types.SourceCredential{
				Secrets: types.Secrets{
					Secret: "test-secret",
				},
				Resources: types.Resources{
					Id: "test-id",
				},
			},
		},
	}
}

func TestRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/paginate", Paginate)

	t.Run("missing input field", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"output": createRequestBody("S3", "CSV", "test-bucket", "test.csv").Output,
		}
		body, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/paginate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing output field", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"input": createRequestBody("S3", "CSV", "test-bucket", "test.csv").Input,
		}
		body, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/paginate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/paginate", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPaginateCSV(t *testing.T) {
	t.Run("CSV pagination with header on each page", func(t *testing.T) {
		// Create CSV data with header + 100 data rows
		csvData := "id,name,email\n"
		for i := 1; i <= 100; i++ {
			csvData += fmt.Sprintf("%d,User%d,user%d@example.com\n", i, i, i)
		}

		output := types.Output{
			StorageType: "S3",
			DataType:    "CSV",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/test",
			},
		}

		// Clear previous stored pages
		mockStoredPages = make(map[string][]byte)

		// Create a custom version of PaginateCSV that uses our mock
		totalPages, err := PaginateCSV([]byte(csvData), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, totalPages) // 100 data rows with chunk size 50 (including header) = 3 pages

		// Verify page 1 content
		page1Key := "test-bucket/output/test/pages/1.csv"
		page1Data, exists := mockStoredPages[page1Key]
		assert.Equal(t, true, exists)

		reader := csv.NewReader(strings.NewReader(string(page1Data)))
		page1Lines, err := reader.ReadAll()
		assert.Equal(t, nil, err)
		assert.Equal(t, 50, len(page1Lines))                                            // header + 49 data rows
		assert.Equal(t, []string{"id", "name", "email"}, page1Lines[0])                 // header
		assert.Equal(t, []string{"1", "User1", "user1@example.com"}, page1Lines[1])     // first data row
		assert.Equal(t, []string{"49", "User49", "user49@example.com"}, page1Lines[49]) // last data row on page 1

		// Verify page 2 content
		page2Key := "test-bucket/output/test/pages/2.csv"
		page2Data, exists := mockStoredPages[page2Key]
		assert.Equal(t, true, exists)

		reader = csv.NewReader(strings.NewReader(string(page2Data)))
		page2Lines, err := reader.ReadAll()
		assert.Equal(t, nil, err)
		assert.Equal(t, 50, len(page2Lines))                                            // header + 49 data rows
		assert.Equal(t, []string{"id", "name", "email"}, page2Lines[0])                 // header
		assert.Equal(t, []string{"50", "User50", "user50@example.com"}, page2Lines[1])  // first data row on page 2
		assert.Equal(t, []string{"98", "User98", "user98@example.com"}, page2Lines[49]) // last data row on page 2

		// Verify page 3 content (remaining 2 data rows)
		page3Key := "test-bucket/output/test/pages/3.csv"
		page3Data, exists := mockStoredPages[page3Key]
		assert.Equal(t, true, exists)

		reader = csv.NewReader(strings.NewReader(string(page3Data)))
		page3Lines, err := reader.ReadAll()
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, len(page3Lines))                                               // header + 2 remaining data rows
		assert.Equal(t, []string{"id", "name", "email"}, page3Lines[0])                   // header
		assert.Equal(t, []string{"99", "User99", "user99@example.com"}, page3Lines[1])    // first remaining data row
		assert.Equal(t, []string{"100", "User100", "user100@example.com"}, page3Lines[2]) // last data row
	})

	t.Run("CSV with less than chunk size", func(t *testing.T) {
		csvData := "id,name\n1,User1\n2,User2\n"

		output := types.Output{
			StorageType: "S3",
			DataType:    "CSV",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/small",
			},
		}

		mockStoredPages = make(map[string][]byte)

		totalPages, err := PaginateCSV([]byte(csvData), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, totalPages)

		// Verify page content
		pageKey := "test-bucket/output/small/pages/1.csv"
		pageData, exists := mockStoredPages[pageKey]
		assert.Equal(t, true, exists)

		reader := csv.NewReader(strings.NewReader(string(pageData)))
		lines, err := reader.ReadAll()
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, len(lines)) // header + 2 data rows
	})

	t.Run("empty CSV", func(t *testing.T) {
		output := types.Output{
			StorageType: "S3",
			DataType:    "CSV",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/empty",
			},
		}

		totalPages, err := PaginateCSV([]byte(""), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, totalPages)
	})
}

func TestPaginateJSONL(t *testing.T) {
	t.Run("JSONL pagination with 50 lines per page", func(t *testing.T) {
		// Create JSONL data with 120 lines
		var jsonlData strings.Builder
		for i := 1; i <= 120; i++ {
			jsonlData.WriteString(fmt.Sprintf(`{"id":%d,"name":"User%d"}`, i, i))
			if i < 120 {
				jsonlData.WriteString("\n")
			}
		}

		output := types.Output{
			StorageType: "S3",
			DataType:    "JSONL",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/test",
			},
		}

		mockStoredPages = make(map[string][]byte)

		totalPages, err := PaginateJSONL([]byte(jsonlData.String()), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 3, totalPages) // 120 lines / 50 = 3 pages (rounded up)

		// Verify page 1 content (first 50 lines)
		page1Key := "test-bucket/output/test/pages/1.jsonl"
		page1Data, exists := mockStoredPages[page1Key]
		assert.Equal(t, true, exists)

		page1Lines := bytes.Split(page1Data, []byte("\n"))
		// Filter out empty lines
		var nonEmptyLines [][]byte
		for _, line := range page1Lines {
			if len(bytes.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		assert.Equal(t, 50, len(nonEmptyLines))
		assert.Equal(t, `{"id":1,"name":"User1"}`, string(nonEmptyLines[0]))
		assert.Equal(t, `{"id":50,"name":"User50"}`, string(nonEmptyLines[49]))

		// Verify page 2 content (next 50 lines)
		page2Key := "test-bucket/output/test/pages/2.jsonl"
		page2Data, exists := mockStoredPages[page2Key]
		assert.Equal(t, true, exists)

		page2Lines := bytes.Split(page2Data, []byte("\n"))
		nonEmptyLines = nil
		for _, line := range page2Lines {
			if len(bytes.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		assert.Equal(t, 50, len(nonEmptyLines))
		assert.Equal(t, `{"id":51,"name":"User51"}`, string(nonEmptyLines[0]))
		assert.Equal(t, `{"id":100,"name":"User100"}`, string(nonEmptyLines[49]))

		// Verify page 3 content (remaining 20 lines)
		page3Key := "test-bucket/output/test/pages/3.jsonl"
		page3Data, exists := mockStoredPages[page3Key]
		assert.Equal(t, true, exists)

		page3Lines := bytes.Split(page3Data, []byte("\n"))
		nonEmptyLines = nil
		for _, line := range page3Lines {
			if len(bytes.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		assert.Equal(t, 20, len(nonEmptyLines))
		assert.Equal(t, `{"id":101,"name":"User101"}`, string(nonEmptyLines[0]))
		assert.Equal(t, `{"id":120,"name":"User120"}`, string(nonEmptyLines[19]))
	})

	t.Run("JSONL with empty lines", func(t *testing.T) {
		jsonlData := `{"id":1,"name":"User1"}

		{"id":2,"name":"User2"}
		
		{"id":3,"name":"User3"}`

		output := types.Output{
			StorageType: "S3",
			DataType:    "JSONL",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/sparse",
			},
		}

		mockStoredPages = make(map[string][]byte)

		totalPages, err := PaginateJSONL([]byte(jsonlData), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, totalPages)

		// Verify only non-empty lines are included
		pageKey := "test-bucket/output/sparse/pages/1.jsonl"
		pageData, exists := mockStoredPages[pageKey]
		assert.Equal(t, true, exists)

		pageLines := bytes.Split(pageData, []byte("\n"))
		nonEmptyLines := [][]byte{}
		for _, line := range pageLines {
			if len(bytes.TrimSpace(line)) > 0 {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		assert.Equal(t, 3, len(nonEmptyLines))
	})

	t.Run("empty JSONL", func(t *testing.T) {
		output := types.Output{
			StorageType: "S3",
			DataType:    "JSONL",
			Reference: types.SourceReference{
				Bucket: "test-bucket",
				Prefix: "output/empty",
			},
		}

		totalPages, err := PaginateJSONL([]byte(""), 50, output, mockStoreBytes)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, totalPages)
	})
}
