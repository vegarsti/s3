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

var bucketName string
var awsRegion string

func readEnvVars() error {
	awsRegion = os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return fmt.Errorf("set environment variable AWS_REGION")
	}
	bucketName = os.Getenv("AWS_BUCKET")
	if bucketName == "" {
		return fmt.Errorf("set environment variable AWS_BUCKET")
	}
	return nil
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "s3: %v\n", err)
	os.Exit(1)
}

func newSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	if err != nil {
		return nil, fmt.Errorf("session.NewSession: %v", err)
	}
	return sess, nil
}

func upload(filename string) error {
	sess, err := newSession()
	if err != nil {
		return fmt.Errorf("newSession: %v", err)
	}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("%v", err)
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
	sess, err := newSession()
	if err != nil {
		return fmt.Errorf("newSession: %v", err)
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
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("bucket not found")
			case s3.ErrCodeNoSuchKey:
				return fmt.Errorf("download %s/%s: no such file", bucketName, filename)
			default:
				return fmt.Errorf("downloader.Download: %v", err)
			}
		} else {
			return fmt.Errorf("downloader.Download: %v", err)
		}
	}
	return nil
}

func delete(filename string) error {
	sess, err := newSession()
	if err != nil {
		return fmt.Errorf("newSession: %v", err)
	}
	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	}
	deleter := s3.New(sess)
	if _, err := deleter.DeleteObject(deleteParams); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("bucket not found")
			case s3.ErrCodeNoSuchKey:
				return fmt.Errorf("delete %s/%s: no such file", bucketName, filename)
			default:
				return fmt.Errorf("deleter.DeleteObject: %v", err)
			}
		} else {
			return fmt.Errorf("deleter.DeleteObject: %v", err)
		}
	}
	return nil
}

func list() error {
	sess, err := newSession()
	if err != nil {
		return fmt.Errorf("newSession: %v", err)
	}
	client := s3.New(sess)
	input := &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)}
	output, err := client.ListObjectsV2(input)
	if err != nil {
		return fmt.Errorf("ListObjectsV2: %v", err)
	}
	for _, object := range output.Contents {
		fmt.Fprintln(os.Stdout, *object.Key)
	}
	return nil
}

func validSubcommand() bool {
	switch {
	case os.Args[1] == "upload" && len(os.Args) > 2:
		return true
	case os.Args[1] == "download" && len(os.Args) > 2:
		return true
	case os.Args[1] == "delete" && len(os.Args) > 2:
		return true
	case os.Args[1] == "list" && len(os.Args) == 2:
		return true
	default:
		return false
	}
}

func main() {
	usageString := "usage: s3 [upload|download|delete|list] [file ...]\n"
	if len(os.Args) < 2 || !validSubcommand() {
		fmt.Fprintf(os.Stderr, usageString)
		os.Exit(1)
	}
	if err := readEnvVars(); err != nil {
		die(err)
	}
	switch os.Args[1] {
	case "upload":
		{
			for _, filename := range os.Args[2:] {
				if err := upload(filename); err != nil {
					die(fmt.Errorf("upload: %v", err))
				}
			}
		}
	case "download":
		{
			for _, filename := range os.Args[2:] {
				if err := download(filename); err != nil {
					die(fmt.Errorf("download: %v", err))
				}
			}
		}
	case "list":
		{
			if len(os.Args) != 2 {
				fmt.Fprintf(os.Stderr, usageString)
				os.Exit(1)
			}
			if err := list(); err != nil {
				die(fmt.Errorf("list: %v", err))
			}
		}
	case "delete":
		{
			for _, filename := range os.Args[2:] {
				if err := delete(filename); err != nil {
					die(fmt.Errorf("download: %v", err))
				}
			}
		}
	}
}
