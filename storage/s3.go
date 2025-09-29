package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"lazy-lagoon/pkg/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

/*
Download the content from the s3 bucket
*/
func DownloadFromS3(input types.Input) ([]byte, error) {
	var reference types.SourceReference = input.Reference
	var credential types.SourceCredential = input.Credential

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(reference.Region),
		Credentials: credentials.NewStaticCredentials(
			credential.Resources.Id,
			credential.Secrets.Secret,
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)

	downloadInput := &s3.GetObjectInput{
		Bucket: aws.String(reference.Bucket),
		Key:    aws.String(reference.Prefix),
	}

	downloadResult, err := s3Client.GetObject(downloadInput)
	if err != nil {
		return nil, err
	}
	defer downloadResult.Body.Close()

	fileContents, err := io.ReadAll(downloadResult.Body)
	if err != nil {
		return nil, err
	}

	return fileContents, nil
}

/*
Upload the content to the s3 bucket
*/
func UploadToS3(output types.Output, content []byte) error {
	var reference types.SourceReference = output.Reference
	var credential types.SourceCredential = output.Credential

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(reference.Region),
		Credentials: credentials.NewStaticCredentials(
			credential.Resources.Id,
			credential.Secrets.Secret,
			"",
		),
	})
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)

	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(reference.Bucket),
		Key:         aws.String(reference.Prefix),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(getContentTypeForDataType(output.DataType)),
	}

	_, err = s3Client.PutObject(uploadInput)
	if err != nil {
		return err
	}

	return nil
}

/*
Create a multipart client for the s3 bucket
*/
func CreateS3MultiPartClient(output types.Output) (*s3.S3, string, error) {
	var reference types.SourceReference = output.Reference
	var credential types.SourceCredential = output.Credential

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(reference.Region),
		Credentials: credentials.NewStaticCredentials(
			credential.Resources.Id,
			credential.Secrets.Secret,
			"",
		),
	})
	if err != nil {
		return nil, "", err
	}

	s3Client := s3.New(sess)

	var uploadId string
	createOutput, err := s3Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(reference.Bucket),
		Key:         aws.String(reference.Prefix),
		ContentType: aws.String(getContentTypeForDataType(output.DataType)),
	})
	if err != nil {
		return nil, "", err
	}
	if createOutput != nil {
		if createOutput.UploadId != nil {
			uploadId = *createOutput.UploadId
		}
	}
	if uploadId == "" {
		return nil, "", fmt.Errorf("no upload id found in start upload request")
	}

	return s3Client, uploadId, nil
}

/*
Upload the content chunk to the s3 bucket
*/
func UploadChunkToS3(s3Client *s3.S3, output types.Output, partNumber int64, uploadId string, content []byte) (*s3.UploadPartOutput, error) {
	var reference types.SourceReference = output.Reference
	uploadPartOutput, err := s3Client.UploadPart(&s3.UploadPartInput{
		Body:          bytes.NewReader(content),
		Bucket:        aws.String(reference.Bucket),
		Key:           aws.String(reference.Prefix),
		PartNumber:    aws.Int64(partNumber),
		UploadId:      aws.String(uploadId),
		ContentLength: aws.Int64(int64(binary.Size(content))),
	})
	if err != nil {
		abortIn := s3.AbortMultipartUploadInput{
			UploadId: &uploadId,
		}
		//ignoring any errors with aborting the copy
		s3Client.AbortMultipartUpload(&abortIn)
		return nil, fmt.Errorf("Error uploading part %d : %w", partNumber, err)
	}

	return uploadPartOutput, nil
}

/*
Complete the chunk upload to the s3 bucket
*/
func CompleteChunkS3Upload(s3Client *s3.S3, output types.Output, uploadId string, uploadedParts []*s3.CompletedPart) error {
	var reference types.SourceReference = output.Reference
	completeUploadInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(reference.Bucket),
		Key:      aws.String(reference.Prefix),
		UploadId: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: uploadedParts,
		},
	}
	_, err := s3Client.CompleteMultipartUpload(completeUploadInput)
	if err != nil {
		fmt.Println("Error CompleteMultipartUpload:", err)
		return err
	}
	return nil
}
