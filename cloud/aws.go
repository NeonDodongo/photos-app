package cloud

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AWSClient struct {
	S3         *s3.S3
	S3Uploader *s3manager.Uploader
}

func InitAWSClient(name string) AWSClient {
	awsCli := AWSClient{}
	s, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		fmt.Println("Error creating AWS session, ", err)
	}
	awsCli.S3 = s3.New(s)
	awsCli.S3Uploader = s3manager.NewUploader(s)

	return awsCli
}
