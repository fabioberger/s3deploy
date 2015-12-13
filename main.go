package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fabioberger/s3/s3util"
)

func main() {
	s3util.DefaultConfig.AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3util.DefaultConfig.SecretKey = os.Getenv("S3_SECRET_KEY")

	currentDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	bucketName := "countdown-christmas" // default bucket name
	if len(args) > 0 {
		bucketName = args[0]
	}

	amazonBucketUrl := "https://" + bucketName + ".s3.amazonaws.com"
	err = filepath.Walk(currentDirectory, func(fullPath string, f os.FileInfo, err error) error {

		relativePath := fullPath[len(currentDirectory):]
		if !strings.Contains(relativePath, ".") {
			return nil
		}
		finalS3Url := amazonBucketUrl + relativePath
		copy(finalS3Url, fullPath)
		return nil
	})

	fmt.Println("Deployed Successfully!")
}

func copy(finalS3Url string, fullPath string) {

	r, err := open(fullPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w, err := create(finalS3Url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = io.Copy(w, r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = w.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Copied", fullPath)
}

func open(s string) (io.ReadCloser, error) {
	if isURL(s) {
		return s3util.Open(s, nil)
	}
	return os.Open(s)
}

func create(s string) (io.WriteCloser, error) {
	if isURL(s) {
		header := make(http.Header)
		header.Add("x-amz-acl", "public-read")
		return s3util.Create(s, header, nil)
	}
	return os.Create(s)
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
