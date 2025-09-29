package storage

import (
	"fmt"
	"lazy-lagoon/pkg/httphelper"
	"lazy-lagoon/pkg/types"
)

/*
	Download the file from the input storage type
*/
func GetBytes(input types.Input) ([]byte, error) {
	var storageType = input.StorageType
	var content []byte
	var err error

	if storageType == "S3" || storageType == "FILE" {
		content, err = DownloadFromS3(input)
		if err != nil {
			return nil, err
		}
	} else if storageType == "REST" {
		content, err = httphelper.Get(input.Reference.Host, input.Credential.Secrets.ApiBearerToken)
		if err != nil {
			return nil, err
		}
	} else if storageType == "ONEDRIVE" {
		content, err = httphelper.Get(
			"https://graph.microsoft.com/v1.0/me/drive/items/"+
				input.Reference.Id+"/content",
			input.Credential.Secrets.AccessToken,
		)
		if err != nil {
			return nil, err
		}
	} else if storageType == "GOOGLEDRIVE" {
		content, err = DownloadFromGoogleDrive(input)
		if err != nil {
			return nil, err
		}
	} else if storageType == "RDB" {
		content, err = DownloadFromRDB(input)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("storage type not found")
	}
	return content, nil
}
