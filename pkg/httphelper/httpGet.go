package httphelper

import (
	"fmt"
	"io"
	"net/http"
)

/*
	Get the content from the url
*/
func Get(url string, apiBearerToken string) ([]byte, error) {
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request with bearer token authorization header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	if apiBearerToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiBearerToken))
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	return body, nil
}
