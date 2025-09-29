package routes

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"

	"lazy-lagoon/pkg/types"
	"lazy-lagoon/storage"

	"github.com/gin-gonic/gin"
)

/*
Paginate the file - used for preview
*/
func Paginate(c *gin.Context) {
	/*
		Request body
	*/
	var requestData types.RequestBodyPaginate

	err := bindAndValidate(c, &requestData)
	if err != nil {
		sendError(c, http.StatusBadRequest, err, nil)
		return
	}

	input := requestData.Input
	output := requestData.Output
	chunkSize := 50
	totalPages := 0

	/*
		Downloading the file from the input storage type
	*/
	bytesContent, err := storage.GetBytes(input)
	if err != nil {
		sendError(c, http.StatusBadRequest, err, nil)
		return
	}

	// Get paths of the fields from the unMutated byte content to send as attributes
	paths, err := extractPaths(bytesContent, input.DataType)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err, nil)
		return
	}

	/*
		Paginating the file based on the data type
	*/
	if input.DataType == "CSV" || input.DataType == "SQL" {
		totalPages, err = PaginateCSV(bytesContent, chunkSize, output, storage.StoreBytes)
		if err != nil {
			sendError(c, http.StatusBadRequest, err, nil)
			return
		}
	} else if input.DataType == "JSONL" {
		totalPages, err = PaginateJSONL(bytesContent, chunkSize, output, storage.StoreBytes)
		if err != nil {
			sendError(c, http.StatusBadRequest, err, nil)
			return
		}
	} else if input.DataType == "JSON" {
		log.Println("JSON is not supported for pagination, but we still return the attributes")
	}

	/*
		Return pagination result
	*/
	result := types.PaginationResult{
		Message:    fmt.Sprintf("Success: %d chunks paginated", totalPages),
		TotalPages: totalPages,
		Attributes: types.Attributes{
			Paths: paths,
		},
	}

	c.JSON(http.StatusOK, result)
}

/*
Paginate the CSV file into chunks - used for preview
*/
func PaginateCSV(bytesContent []byte, chunkSize int, output types.Output, storeBytes func(output types.Output, content []byte) error) (totalPages int, err error) {
	reader := csv.NewReader(strings.NewReader(string(bytesContent)))
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	lines, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("file was not able to process, please use a different file")
	}

	if len(lines) == 0 {
		return 0, nil
	}

	totalPages = 0
	position := 0
	header := lines[0] // Store the header row

	for position < len(lines) {
		buffer := bytes.NewBuffer(nil)
		writer := csv.NewWriter(buffer)
		linesAdded := 0

		// Always add the header as the first row of each page
		if err := writer.Write(header); err != nil {
			return 0, fmt.Errorf("error writing CSV header: %v", err)
		}
		linesAdded++

		// Process up to chunkSize lines starting from current position
		// Skip the header row in the first iteration
		startPos := position
		if position == 0 {
			startPos = 1 // Skip header row for data processing
		}

		for i := startPos; i < len(lines) && linesAdded < chunkSize; i++ {
			if err := writer.Write(lines[i]); err != nil {
				return 0, fmt.Errorf("error writing CSV line: %v", err)
			}
			linesAdded++
			position = i + 1
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return 0, fmt.Errorf("error flushing csv writer: %v", err)
		}

		// Store the page
		if buffer.Len() > 0 {
			totalPages++
			outputCopy := output
			outputCopy.Reference.Prefix = fmt.Sprintf("%s/pages/%d.csv", output.Reference.Prefix, totalPages)
			err = storeBytes(outputCopy, buffer.Bytes())
			if err != nil {
				return 0, err
			}
		}

		// Break if we've processed all lines
		if position >= len(lines) {
			break
		}
	}

	return totalPages, nil
}

/*
Paginate the JSONL file - used for preview
*/
func PaginateJSONL(byteContent []byte, chunkSize int, output types.Output, storeBytes func(output types.Output, content []byte) error) (totalPages int, err error) {
	totalPages = 0
	position := 0
	lines := bytes.Split(byteContent, []byte("\n"))

	for position < len(lines) {
		var page []byte
		linesAdded := 0

		// Process up to chunkSize lines starting from current position
		for i := position; i < len(lines) && linesAdded < chunkSize; i++ {
			// If the line is not empty, add it to the page
			if len(bytes.TrimSpace(lines[i])) > 0 {
				if len(page) > 0 {
					page = append(page, []byte("\n")...)
				}
				page = append(page, lines[i]...)
				linesAdded++
			}
			position++
		}

		// If the page has content, store it
		if len(page) > 0 {
			totalPages++
			outputCopy := output
			outputCopy.Reference.Prefix = fmt.Sprintf("%s/pages/%d.jsonl", output.Reference.Prefix, totalPages)
			err = storeBytes(outputCopy, page)
			if err != nil {
				return 0, err
			}
		}
	}

	return totalPages, nil
}
