package transformcsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"lazy-lagoon/pkg/concurrent"
	"lazy-lagoon/pkg/types"
	"lazy-lagoon/storage"
	"slices"
	"strings"
)

/*
Step 1: execute transform, and store in output storage
*/
func ExecuteTransform(input types.Input, rules []types.Rule, output types.Output) *types.TransformError {
	/*
		Downloading the file from the input storage type
	*/
	byteContent, err := storage.GetBytes(input)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}
	/*
		Chunking the file
	*/
	chunks, err := Chunk(byteContent)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}

	// If output is specified, transform and upload the data
	return processChunksWithOutput(chunks, rules, output)
}

// processChunksWithOutput transforms the chunks and uploads them to the specified output
func processChunksWithOutput(chunks [][]byte, rules []types.Rule, output types.Output) *types.TransformError {
	// Create a client for multipart uploads
	client, uploadId, err := storage.CreateMultiPartClient(output)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}

	// Initialize uploadedParts based on storage type
	uploadedParts, err := storage.InitializeUploadedParts(output)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}
	// Process and upload each chunk
	transformErr := concurrent.ForEachVoid(chunks, func(chunk []byte, index int) *types.TransformError {
		// Transform the chunk
		csvLines, err := ToCsv(chunk)
		if err != nil {
			return &types.TransformError{Message: err.Error()}
		}
		/*
			Transforming the CSV lines
		*/
		csvLines, transformErr := ExecuteRules(csvLines, rules)
		if transformErr != nil {
			return transformErr
		}
		/*
			Converting the CSV lines back to bytes
		*/
		transformedBytes, err := FromCsv(csvLines)
		if err != nil {
			return &types.TransformError{Message: err.Error(), Key: "upload"}
		}

		// Upload the transformed chunk
		var partNumber int64 = int64(index + 1)
		isLastChunk := index == (len(chunks) - 1)

		uploadedPart, err := storage.UploadAndCompleteChunk(client, output, partNumber, uploadId, transformedBytes, isLastChunk, uploadedParts)
		if err != nil {
			return &types.TransformError{Message: err.Error()}
		}

		// Update uploadedParts with the new part
		uploadedParts, err = storage.UpdateUploadedParts(output, uploadedParts, uploadedPart, partNumber)
		if err != nil {
			return &types.TransformError{Message: err.Error(), Key: "upload"}
		}

		return nil
	})
	if transformErr != nil {
		return transformErr
	}

	// Return empty byte slice as the data has been uploaded
	return nil
}

/*
Step 2: Transform the CSV document based on the rules
*/
func ExecuteRules(lines [][]string, rules []types.Rule) ([][]string, *types.TransformError) {
	// Create a deep copy of the original lines to preserve them
	documentCopy := make([][]string, len(lines))
	for i, line := range lines {
		documentCopy[i] = make([]string, len(line))
		copy(documentCopy[i], line)
	}

	// Apply the rules to the lines
	for ruleIndex, rule := range rules {
		for actionIndex, action := range rule.Actions {
			transformErr := ExecuteAction(lines, documentCopy, action.FieldName, rule.Expression, action.ActionType, ruleIndex, actionIndex)
			if transformErr != nil {
				return nil, transformErr
			}
		}
	}
	return lines, nil
}

/*
Step 3: Execute the actions by the actionType if the expressions are met
*/
func ExecuteAction(lines [][]string, unMutatedLines [][]string, fieldName string, expression types.Expression, actionType string, ruleIndex int, actionIndex int) *types.TransformError {
	if fieldName == "" {
		// No op if the field name is empty
		return nil
	}
	// lines[0] is the header line
	column := slices.Index(lines[0], fieldName)
	// If theres an index found - fault safety
	if column < 0 {
		return &types.TransformError{
			Message: "column not found",
			RuleIndex: &ruleIndex,
			ActionIndex: &actionIndex,
			Key: "fieldName",
		}
	}
	for index, line := range lines {
		// Skip the header
		if index == 0 {
			continue
		}
		if column >= len(line) {
			// If the column is out of range then skip the line
			continue
		}
		// Checking if the expressions are met
		met, transformErr := IsExpressionMet(expression, unMutatedLines, index, ruleIndex)
		if transformErr != nil {
			return transformErr
		}
		if !met {
			continue
		}

		// Applying the operation if the expressions are met
		if actionType == "REDACT" {
			// Don't redact the header
			if index != 0 {
				// Redact the column
				if line[column] != "" {
					line[column] = "**redacted**"
				}
			}
		} else {
			// Exclude column
			lines[index] = append(lines[index][:column], lines[index][column+1:]...)
		}
	}
	return nil
}

/*
	Helper functions
*/
// Convert the bytes content into a 2d slice of strings which is the csv content
func ToCsv(bytesContent []byte) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(bytesContent)))
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not decode your input, please upload a new csv file")
	}
	return lines, nil
}

// Convert the 2d slice of strings into bytes
func FromCsv(lines [][]string) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buffer)
	writer.WriteAll(lines)
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing csv writer: %v", err)
	}
	return buffer.Bytes(), nil
}
