package storage

import (
	"fmt"
	"lazy-lagoon/pkg/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

/*
Store the content in the output storage type
*/
func StoreBytes(output types.Output, content []byte) error {
	switch output.StorageType {
	case "S3":
		return UploadToS3(output, content)
	// Future storage types can be added here
	default:
		return fmt.Errorf("storage type '%s' not supported", output.StorageType)
	}
}

/*
CreateMultiPartClient creates a client for multipart uploads based on the storage type
*/
func CreateMultiPartClient(output types.Output) (any, string, error) {
	var storageType = output.StorageType

	if storageType == "S3" {
		return CreateS3MultiPartClient(output)
	}

	return nil, "", fmt.Errorf("storage type %s not supported for multipart uploads", storageType)
}

/*
UploadAndCompleteChunk uploads a single chunk to the specified storage and completes the upload if it's the last chunk
*/
func UploadAndCompleteChunk(client any, output types.Output, partNumber int64, uploadId string, chunk []byte, isLastChunk bool, uploadedParts any) (any, error) {
	var storageType = output.StorageType
	var uploadedPart any
	var err error

	if storageType == "S3" {
		s3Client, ok := client.(*s3.S3)
		if !ok {
			return nil, fmt.Errorf("client is not an S3 client")
		}
		uploadedPart, err = UploadChunkToS3(s3Client, output, partNumber, uploadId, chunk)
		if err != nil {
			return nil, err
		}

		if isLastChunk {
			s3Part, ok := uploadedPart.(*s3.UploadPartOutput)
			if !ok {
				return nil, fmt.Errorf("uploaded part is not an S3 upload part")
			}
			uploadedParts = append(uploadedParts.([]*s3.CompletedPart), &s3.CompletedPart{
				ETag:       s3Part.ETag,
				PartNumber: aws.Int64(partNumber),
			})

			err = CompleteChunkS3Upload(s3Client, output, uploadId, uploadedParts.([]*s3.CompletedPart))
			if err != nil {
				return nil, err
			}
		}

		return uploadedPart, nil
	}

	return nil, fmt.Errorf("storage type %s not supported for chunk uploads", storageType)
}

/*
InitializeUploadedParts initializes the uploadedParts variable based on the storage type
*/
func InitializeUploadedParts(output types.Output) (any, error) {
	var storageType = output.StorageType

	if storageType == "S3" {
		return []*s3.CompletedPart{}, nil
	}

	return nil, fmt.Errorf("storage type %s not supported for multipart uploads", storageType)
}

/*
UpdateUploadedParts updates the uploadedParts with a new uploaded part
*/
func UpdateUploadedParts(output types.Output, uploadedParts any, uploadedPart any, partNumber int64) (any, error) {
	var storageType = output.StorageType

	if storageType == "S3" {
		s3Part := uploadedPart.(*s3.UploadPartOutput)
		return append(uploadedParts.([]*s3.CompletedPart), &s3.CompletedPart{
			ETag:       s3Part.ETag,
			PartNumber: aws.Int64(partNumber),
		}), nil
	}

	return nil, fmt.Errorf("storage type %s not supported for updating uploaded parts", storageType)
}
