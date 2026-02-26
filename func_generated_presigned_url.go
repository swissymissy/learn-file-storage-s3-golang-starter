package main 

import (
	"time"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// create a limited time signed url
func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {

	// create new client
	presignedClient := s3.NewPresignClient(s3Client)

	// generate a signed URL
	signedURL, err := presignedClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		fmt.Printf("Error generate signed url: %s", err)
		return "", err
	}

	return signedURL.URL, nil
}