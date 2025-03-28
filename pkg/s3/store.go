package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/klauspost/compress/zstd"
	"github.com/martient/golang-utils/utils"
)

func getS3Client(storage S3Requirements) (*s3.Client, error) {
	options := []func(*config.LoadOptions) error{
		// Add credentials provider
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			storage.AccessKeyId,
			storage.AccessKeySecret,
			"",
		)),
		config.WithRegion(storage.Region),
	}

	// Load the configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, err
	}

	// Create S3 client options
	s3Options := []func(*s3.Options){}

	// Add custom endpoint if specified
	if storage.Endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(storage.Endpoint)
		})
	}

	// Create and return the S3 client with custom options
	return s3.NewFromConfig(cfg, s3Options...), nil
}

func createBucket(client *s3.Client, name string, region string) error {
	if client == nil {
		return fmt.Errorf("s3 client can't be null for the bucket creation")
	} else if len(name) <= 0 || len(region) <= 0 {
		return fmt.Errorf("the bucket need a name and a region, none of them can be null")
	}
	_, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})
	if err != nil {
		log.Printf("Couldn't create bucket %v in Region %v. Here's why: %v\n",
			name, region, err)
	}
	return err
}

func upload(client *s3.Client, bucket_name string, buffer []byte) error {
	if client == nil {
		return fmt.Errorf("s3 client can't be null for the upload operation")
	} else if len(bucket_name) <= 0 {
		return fmt.Errorf("the bucket need a name, can't be null at the upload")
	} else if len(buffer) <= 0 {
		return fmt.Errorf("the buffer can't be nil or empty at the bucket upload")
	}
	currentTime := time.Now().UTC()

	largeBuffer := bytes.NewReader(buffer)
	var partMiBs int64 = 100
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket_name),
		Key: aws.String(fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
			currentTime.Year(),
			currentTime.Month(),
			currentTime.Day(),
			currentTime.Hour(),
			currentTime.Minute(),
			currentTime.Second())),

		Body: largeBuffer,
	})
	if err != nil {
		log.Printf("Couldn't upload large object to %v:%v. Here's why: %v\n",
			bucket_name, fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day(),
				currentTime.Hour(),
				currentTime.Minute(),
				currentTime.Second()), err)
		return err
	}
	return nil
}

func StoreBackup(storage S3Requirements, buffer *bytes.Buffer, useCompression bool) error {
	if buffer == nil {
		return fmt.Errorf("buffer can't be empty")
	} else if storage == (S3Requirements{}) {
		return fmt.Errorf("storage can't be empty")
	}
	client, err := getS3Client(storage)
	if err != nil {
		return err
	}
	hb, err := client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: &storage.BucketName,
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				utils.LogWarning("The bucket %s does not exist, it gonna be created", "S3", storage.BucketName)
			default:
				return err
			}
		}
	}

	var dataToWrite []byte

	if useCompression {
		// compressed, err := zstd.Compress(nil, buffer.Bytes())
		// if err != nil {
		// 	utils.LogError("Compression failed", "Local storage", err)
		// 	return err
		// }
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			utils.LogError("Compression failed", "S3", err)
			return err
		}
		defer func() {
			if err := encoder.Close(); err != nil {
				utils.LogError("Failed to close encoder", "S3", err)
			}
		}()

		// Compress the input string
		compressed := encoder.EncodeAll([]byte(buffer.Bytes()), nil)
		dataToWrite = compressed
	} else {
		dataToWrite = buffer.Bytes()
	}

	if hb != nil {
		return upload(client, storage.BucketName, dataToWrite)
	} else {
		err = createBucket(client, storage.BucketName, storage.Region)
		if err != nil {
			return err
		}
		return upload(client, storage.BucketName, dataToWrite)
	}
}
