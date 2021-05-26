package sigs3scann3r

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type sigs3scann3r struct {
	Service    s3iface.S3API
	Downloader *s3manager.Downloader
}

func New(region string) (sigs3scann3r, error) {
	sigs3scann3r := sigs3scann3r{}

	// Initialize a session in us-west-2 that the SDK will use to load credentials
	// from the shared credentials file. (~/.aws/credentials).
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	// Create S3 service client
	sigs3scann3r.Service = s3.New(sess)

	// Create S3 service client
	sigs3scann3r.Downloader = s3manager.NewDownloader(sess)

	return sigs3scann3r, nil
}
