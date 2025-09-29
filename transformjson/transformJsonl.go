package transformjson

import (
	"bytes"
	"fmt"
	"lazy-lagoon/pkg/concurrent"
	"lazy-lagoon/pkg/types"
	"lazy-lagoon/storage"
)

// ExecuteTransformJsonl processes JSONL content concurrently and handles multipart uploads
func ExecuteTransformJsonl(input types.Input, rules []types.Rule, output types.Output) *types.TransformError {
	/*
		Downloading the file from the input storage type
	*/
	byteContent, err := storage.GetBytes(input)
	if err != nil {
		return &types.TransformError{Message: err.Error()}
	}
	// Split content into lines
	lines := bytes.Split(byteContent, []byte("\n"))

	// Process lines into chunks
	const chunkSize = 5 * 1024 * 1024 // 5MB in bytes
	var chunks [][][]byte
	var currentChunk [][]byte
	var currentSize int

	for _, line := range lines {
		// Skip empty lines
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		lineSize := len(line) + 1 // +1 for newline
		if currentSize+lineSize > chunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
			currentSize = 0
		}
		currentChunk = append(currentChunk, line)
		currentSize += lineSize
	}

	// Add the last chunk if it's not empty
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	// If output is specified, transform and upload the data
	return processJsonlWithOutput(chunks, rules, output)
}

// processJsonlWithOutput transforms the JSONL lines and uploads them to the specified output
func processJsonlWithOutput(chunks [][][]byte, rules []types.Rule, output types.Output) *types.TransformError {
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

	// Process each chunk
	transformErr := concurrent.ForEachVoid(chunks, func(chunk [][]byte, chunkIndex int) *types.TransformError {
		var buffer []byte

		// Process each line in the chunk
		for lineIndex, line := range chunk {
			// Parse the JSON line
			jsonDoc, err := ToJson(line)
			if err != nil {
				return &types.TransformError{Message: fmt.Sprintf("error parsing JSON line %d: %v", lineIndex+1, err)}
			}

			// Transform the JSON document
			transformedDoc, transformErr := ExecuteRules(jsonDoc, rules)
			if transformErr != nil {
				return transformErr
			}

			// Convert back to bytes
			transformedBytes, err := FromJsonl(transformedDoc)
			if err != nil {
				return &types.TransformError{Message: fmt.Sprintf("error converting transformed JSON to bytes: %v", err)}
			}

			// Add to buffer with newline
			if len(buffer) > 0 {
				buffer = append(buffer, '\n')
			}
			buffer = append(buffer, transformedBytes...)
		}

		// Upload the transformed chunk
		var partNumber int64 = int64(chunkIndex + 1)
		isLastChunk := chunkIndex == (len(chunks) - 1)

		// Ensure the last chunk ends with a newline
		if isLastChunk && len(buffer) > 0 && buffer[len(buffer)-1] != '\n' {
			buffer = append(buffer, '\n')
		}

		uploadedPart, err := storage.UploadAndCompleteChunk(client, output, partNumber, uploadId, buffer, isLastChunk, uploadedParts)
		if err != nil {
			return &types.TransformError{Message: err.Error()}
		}

		// Update uploadedParts with the new part
		uploadedParts, err = storage.UpdateUploadedParts(output, uploadedParts, uploadedPart, partNumber)
		if err != nil {
			return &types.TransformError{Message: err.Error()}
		}

		partNumber++
		return nil
	})

	if transformErr != nil {
		return transformErr
	}

	// Return empty byte slice as the data has been uploaded
	return nil
}
