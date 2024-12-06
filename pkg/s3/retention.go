package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/martient/golang-utils/utils"
)

// Make me a retention policies for s3 to delete old backups with a date > 21 days
func deleteOldBackups(client *s3.Client, bucket_name string, retentionDays int) error {
	if client == nil {
		return fmt.Errorf("s3 client can't be null for the list operation")
	} else if len(bucket_name) <= 0 {
		return fmt.Errorf("the bucket need a name, can't be null at the list operation")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	listObjectsOutput, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucket_name,
	})
	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %s: %v", bucket_name, err)
	}

	// Iterate over the objects and delete those older than the cutoff date
	for _, obj := range listObjectsOutput.Contents {
		if obj.LastModified.Before(cutoffDate) {
			_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: &bucket_name,
				Key:    obj.Key,
			})
			if err != nil {
				utils.LogErrorInterface("Failed to delete object %s from bucket %s: %v", "S3", *obj.Key, "", bucket_name, err)
			} else {
				utils.LogInfo("Deleted object %s from bucket %s", "S3", *obj.Key, bucket_name)
			}
		}
	}

	return nil
}

func ExecuteRetentionPolicy(storage S3Requirements, retention_days int) error {
	if storage == (S3Requirements{}) {
		return fmt.Errorf("storage can't be empty")
	}

	client, err := getS3Client(storage)
	if err != nil {
		return err
	}

	return deleteOldBackups(client, storage.BucketName, retention_days)
}
