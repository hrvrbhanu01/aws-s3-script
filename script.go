package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

func main() {
	// Create an AWS session using the "default" profile from your AWS credentials file.
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "default", // Use the "default" profile
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	// Create an S3 client using the "us-east-1" region.
	svc := s3.New(sess, &aws.Config{
		Region: aws.String("us-east-1"), // Use "us-east-1" as the default region
	})

	// List all AWS regions.
	ec2Svc := ec2.New(sess)
	regions, err := ec2Svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		fmt.Println("Error describing regions:", err)
		return
	}

	// Iterate through each region and handle S3 buckets.
	for _, region := range regions.Regions {
		fmt.Printf("Handling S3 buckets in region: %s\n", *region.RegionName)
		err := handleBucketsInRegion(svc, *region.RegionName)
		if err != nil {
			fmt.Printf("Error handling buckets in region %s: %v\n", *region.RegionName, err)
		}
	}
}

// Helper function to handle S3 buckets in a specific region.
func handleBucketsInRegion(svc *s3.S3, region string) error {
	// Update the region for the S3 client.
	svc.Config.Region = aws.String(region)

	// List all S3 buckets in the specified region.
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}

	// Iterate through the bucket list.
	for _, bucket := range result.Buckets {
		bucketName := *bucket.Name

		// Exclude buckets with names containing "chikitsa," "echaritra," or "details."
		if !containsSubstring(bucketName, "chikitsa") && !containsSubstring(bucketName, "echaritra") && !containsSubstring(bucketName, "details") {
			// Empty the bucket.
			err := emptyBucket(svc, bucketName)
			if err != nil {
				fmt.Printf("Failed to empty bucket %s in region %s: %v\n", bucketName, region, err)
			} else {
				fmt.Printf("Emptied bucket: %s in region %s\n", bucketName, region)
				// Now delete the empty bucket.
				err := deleteBucket(svc, bucketName)
				if err != nil {
					fmt.Printf("Failed to delete bucket %s in region %s: %v\n", bucketName, region, err)
				} else {
					fmt.Printf("Deleted bucket: %s in region %s\n", bucketName, region)
				}
			}
		} else {
			fmt.Printf("Excluded bucket: %s in region %s\n", bucketName, region)
		}
	}

	return nil
}

// Helper function to empty a bucket.
func emptyBucket(svc *s3.S3, bucketName string) error {
	objects, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	for _, obj := range objects.Contents {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    obj.Key,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function to delete an empty bucket.
func deleteBucket(svc *s3.S3, bucketName string) error {
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// Helper function to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
