package s3

type S3Requirements struct {
	BucketName      string `json:"bucket_name"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
}
