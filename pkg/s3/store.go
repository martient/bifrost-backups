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
	"github.com/martient/golang-utils/utils"
)

func getS3Client(storage S3Requirements) (*s3.Client, error) {
	var resolver aws.EndpointResolverWithOptions = nil
	if len(storage.Endpoint) <= 0 {
		resolver = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: storage.Endpoint,
			}, nil
		})
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(storage.AccessKeyId, storage.AccessKeySecret, "")),
		config.WithRegion(storage.Region),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
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

func upload(client *s3.Client, bucket_name string, buffer *bytes.Buffer) error {
	if client == nil {
		return fmt.Errorf("s3 client can't be null for the upload operation")
	} else if len(bucket_name) <= 0 {
		return fmt.Errorf("the bucket need a name, can't be null at the upload")
	} else if buffer == nil || buffer.Len() <= 0 {
		return fmt.Errorf("the buffer can't be nil or empty at the bucket upload")
	}
	currentTime := time.Now().UTC()

	largeBuffer := bytes.NewReader(buffer.Bytes())
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

func StoreBackup(storage S3Requirements, buffer *bytes.Buffer) error {
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

	if hb != nil {
		return upload(client, storage.BucketName, buffer)
	} else {
		err = createBucket(client, storage.BucketName, storage.Region)
		if err != nil {
			return err
		}
		return upload(client, storage.BucketName, buffer)
	}
}
