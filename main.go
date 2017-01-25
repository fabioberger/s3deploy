package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fabioberger/s3/s3util"
	"github.com/fabioberger/s3deploy/store"
	"github.com/go-errors/errors"
	"github.com/urfave/cli"
)

func main() {
	s3util.DefaultConfig.AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3util.DefaultConfig.SecretKey = os.Getenv("S3_SECRET_KEY")

	app := cli.NewApp()
	app.Name = "s3deploy"
	app.Usage = "Sync files to S3"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "bucket, b",
			Value: "",
			Usage: "the name of the bucket you want your files uploaded to",
		},
		cli.StringFlag{
			Name:  "region, r",
			Value: "s3",
			Usage: "the S3 region your bucket is located in. e.g \"s3-eu-west-1\"",
		},
	}
	app.Action = func(c *cli.Context) error {
		bucketName := c.String("bucket")
		regionName := c.String("region")
		if bucketName == "" {
			return cli.NewExitError("missing required -bucket argument", 2)
		}

		if err := deployFiles(bucketName, regionName); err != nil {
			fmt.Println(err.(*errors.Error).ErrorStack())
			return cli.NewExitError(err.Error(), 2)
		}
		return nil
	}

	app.Run(os.Args)
}

func deployFiles(bucketName, regionName string) error {
	store, err := store.New()
	if err != nil {
		return err
	}
	defer store.SaveCacheToFile()

	currentDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, 0)
	}
	amazonBucketURL := fmt.Sprintf("https://%s.amazonaws.com/%s", regionName, bucketName)
	if err = filepath.Walk(currentDirectory, func(fullPath string, f os.FileInfo, err error) error {
		relativePath := fullPath[len(currentDirectory):]
		if !strings.Contains(relativePath, ".") {
			return nil // skip
		}

		if !f.IsDir() {
			didFileChange, err := store.DidFileChange(fullPath)
			if err != nil {
				return err
			}
			if didFileChange {
				finalS3Url := amazonBucketURL + relativePath
				if err = uploadFileToS3(finalS3Url, fullPath); err != nil {
					return err
				}
			}
			store.UpdateCache(fullPath)
		}
		store.MarkAsTouched(fullPath)
		return nil
	}); err != nil {
		return errors.Wrap(err, 0)
	}

	// TODO: remove deleted files from S3

	fmt.Println("Deployed Successfully!")
	return nil
}

func uploadFileToS3(s3FilePath, localFilePath string) error {
	r, err := os.Open(localFilePath)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	w, err := createS3Uploader(s3FilePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, 0)
	}

	fmt.Println("Copied", localFilePath)
	return nil
}

func createS3Uploader(s string) (w io.WriteCloser, err error) {
	header := make(http.Header)
	header.Add("x-amz-acl", "public-read")
	w, err = s3util.Create(s, header, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	return w, nil
}
