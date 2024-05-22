package s3

type S3Requirements struct {
	BucketName      string `json:"bucket_name"`
	AccountId       string `json:"account_id"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Endpoint        string `json:"endpoint"`
}
