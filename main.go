package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	bucketName = "golang-to-s3"
	awsRegion  = "eu-west-1"
)

func createFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString("hello world\n"); err != nil {
		return err
	}
	return nil
}

func createBucketIfNotExists(sess *session.Session, bucketName string) error {
	client := s3.New(sess)
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String("eu-west-1"),
		},
	}
	result, err := client.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return fmt.Errorf("%s: %v", s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return aerr
			}
		} else {
			return err
		}
	}
	fmt.Println(result)
	return nil
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "s3upload: %v\n", err)
	os.Exit(1)
}

func upload(filename string) error {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	if err != nil {
		return fmt.Errorf("session.NewSession: %v", err)
	}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %v", err)
	}
	defer f.Close()
	basename := filepath.Base(filename)
	uploader := s3manager.NewUploader(sess)
	uploadParams := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(basename),
		Body:   f,
	}
	if _, err := uploader.Upload(uploadParams); err != nil {
		return fmt.Errorf("uploader.Upload: %v", err)
	}
	return nil
}

func download(filename string) error {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	if err != nil {
		return fmt.Errorf("session.NewSession: %v", err)
	}
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer file.Close()
	downloader := s3manager.NewDownloader(sess)
	downloadParams := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	}
	if _, err := downloader.Download(file, downloadParams); err != nil {
		return fmt.Errorf("downloader.Download: %v", err)
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: s3 [upload|download] [file ...]\n")
		os.Exit(1)
	}
	if os.Args[1] == "upload" {
		if err := upload(os.Args[2]); err != nil {
			die(fmt.Errorf("upload: %v", err))
		}
	}
	if os.Args[1] == "download" {
		if err := download(os.Args[2]); err != nil {
			die(fmt.Errorf("download: %v", err))
		}
	}
}
