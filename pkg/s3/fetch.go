package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getBackupKey(client *s3.Client, bucket_name string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("s3 client can't be null for the list operation")
	} else if len(bucket_name) <= 0 {
		return "", fmt.Errorf("the bucket need a name, can't be null at the list operation")
	}

	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket_name),
	}

	p := s3.NewListObjectsV2Paginator(client, listObjectsInput)

	var latestBackupKey string
	var latestBackupTime time.Time

	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return "", err
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			backupTime, err := time.Parse(time.RFC3339, key)
			if err != nil {
				// Skip keys that don't match the expected date format
				continue
			}

			if backupTime.After(latestBackupTime) {
				latestBackupKey = key
				latestBackupTime = backupTime
			}
		}
	}

	if latestBackupKey == "" {
		return "", fmt.Errorf("no backups found in bucket %s", bucket_name)
	}

	return latestBackupKey, nil
}

func PullBackup(storage S3Requirements, backup_name string) (*bytes.Buffer, error) {
	if storage == (S3Requirements{}) {
		return nil, fmt.Errorf("storage can't be empty")
	}

	client, err := getS3Client(storage)
	if err != nil {
		return nil, err
	}

	latestBackupKey := backup_name
	if latestBackupKey == "" {
		latestBackupKey, err = getBackupKey(client, storage.BucketName)
		if err != nil {
			return nil, err
		}
	}

	obj, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(storage.BucketName),
		Key:    aws.String(latestBackupKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %v", latestBackupKey, storage.BucketName, err)
	}
	defer obj.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, obj.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s from bucket %s: %v", latestBackupKey, storage.BucketName, err)
	}

	return buf, nil
}
