package storage

func getContentTypeForDataType(dataType string) string {
	switch dataType {
	case "CSV":
		return "text/csv"
	case "JSON":
		return "application/json"
	case "JSONL":
		return "application/jsonl"
	case "SQL":
		return "text/csv"
	}
	return "application/octet-stream"
}
