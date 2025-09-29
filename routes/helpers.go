package routes

import (
	"bytes"
	"fmt"
	"lazy-lagoon/pkg/httphelper"
	"lazy-lagoon/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"lazy-lagoon/transformjson"
	"lazy-lagoon/transformcsv"
)

func sendError(c *gin.Context, status int, err error, webhook *types.Webhook) {
	fmt.Println("Error transforming data", err.Error())
	if webhook != nil && webhook.Url != "" {
		payload := webhook.Payload
		payload.Status = "ERROR"
		payload.Msg = err.Error()
		err = httphelper.SendPostRequest(payload, webhook.Url, webhook.ResponseToken)
		if err != nil {
			fmt.Println("Error sending webhook", err.Error())
		}
	}
	c.JSON(status, gin.H{"message": err.Error()})
}

func sendTransformError(c *gin.Context, status int, transformErr *types.TransformError, webhook *types.Webhook) {
	fmt.Println("Error transforming data", transformErr.Message)
	if webhook != nil && webhook.Url != "" {
		payload := webhook.Payload
		payload.Status = "ERROR"
		payload.Msg = transformErr.Message
		err := httphelper.SendPostRequest(payload, webhook.Url, webhook.ResponseToken)
		if err != nil {
			fmt.Println("Error sending webhook", err.Error())
			transformErr = &types.TransformError{
				Message: err.Error(),
			}
		}
	}
	c.JSON(status, gin.H{"transformError": transformErr})
}

func bindAndValidate(c *gin.Context, requestData any) error {
	validate := validator.New()
	err := c.ShouldBindJSON(requestData)
	if err != nil {
		return err
	}
	err = validate.Struct(requestData)
	if err != nil {
		return err
	}
	return nil
}

// Extracts the paths of the fields in the document
func extractPaths(byteContent []byte, dataType string) ([]string, error) {
	switch dataType {
	case "JSON":
		jsonDocument, err := transformjson.ToJson(byteContent)
		if err != nil {
			return nil, err
		}
		return extractJsonPaths(jsonDocument, "", []string{}), nil
	case "CSV":
		csvDocument, err := transformcsv.ToCsv(byteContent)
		if err != nil {
			return nil, err
		}
		return extractCsvHeaders(csvDocument), nil
	case "JSONL":
		paths := []string{}
		lines := bytes.Split(byteContent, []byte("\n"))

		// loop through each line
		for _, line := range lines {
			// Filter out empty lines
			if len(bytes.TrimSpace(line)) > 0 {
				jsonDocument, err := transformjson.ToJson(line)
				if err != nil {
					return nil, err
				}
				// extractJsonPaths already adds unique paths to the provided slice
				paths = extractJsonPaths(jsonDocument, "", paths)
			}
		}
		return paths, nil
	}

	return nil, nil
}

// extractJSONPaths extracts all paths from a JSON document
func extractJsonPaths(jsonObj any, currentPath string, paths []string) []string {
	if jsonObj == nil {
		return paths
	}

	// Handle maps (JSON objects)
	if m, ok := jsonObj.(map[string]any); ok {
		for key, value := range m {
			newPath := key
			if currentPath != "" {
				newPath = currentPath + "." + key
			}

			// Add the current path
			if !contains(paths, newPath) {
				paths = append(paths, newPath)
			}

			// For nested objects/arrays, continue traversing
			if value != nil {
				paths = extractJsonPaths(value, newPath, paths)
			}
		}
		return paths
	}

	// Handle arrays
	if arr, ok := jsonObj.([]any); ok {
		// Add the array path itself
		if currentPath != "" && !contains(paths, currentPath) {
			paths = append(paths, currentPath)
		}

		// Add the array path with [*] notation
		arrayPath := "[*]"
		if currentPath != "" {
			arrayPath = currentPath + "[*]"
		}
		if !contains(paths, arrayPath) {
			paths = append(paths, arrayPath)
		}

		// Recursively process array elements
		for _, item := range arr {
			if item != nil {
				paths = extractJsonPaths(item, arrayPath, paths)
			}
		}
		return paths
	}

	return paths
}

// extractCsvHeaders extracts all headers from a CSV document
func extractCsvHeaders(csvObj [][]string) []string {
	headers := []string{}
	if csvObj == nil {
		return headers
	}
	// Handle CSV objects
	if len(csvObj) > 0 {
		// Only extract headers from the first row
		firstRow := csvObj[0]
		for _, key := range firstRow {
			if !contains(headers, key) {
				headers = append(headers, key)
			}
		}
	}
	return headers
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
