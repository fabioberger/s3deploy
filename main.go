package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fabioberger/s3/s3util"
	"github.com/fabioberger/s3deploy/store"
	"github.com/go-errors/errors"
	"github.com/kr/s3"
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
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "will force all files to be uploaded, even if they haven't changed since last upload",
		},
	}
	app.Action = func(c *cli.Context) error {
		bucketName := c.String("bucket")
		regionName := c.String("region")
		isForceUploadEnabled := c.Bool("all")
		if bucketName == "" {
			return cli.NewExitError("missing required -bucket argument", 2)
		}

		if err := deployFiles(bucketName, regionName, isForceUploadEnabled); err != nil {
			fmt.Println(err.(*errors.Error).ErrorStack())
			return cli.NewExitError(err.Error(), 2)
		}
		return nil
	}

	app.Run(os.Args)
}

func deployFiles(bucketName, regionName string, isForceUploadEnabled bool) error {
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
			if didFileChange || isForceUploadEnabled {
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

	filePathsToDelete, err := store.GetRelativeFilePathsToDeleteOnRemote()
	if err != nil {
		return err
	}
	if len(filePathsToDelete) > 0 {
		if err := deleteFilesOnS3(amazonBucketURL, filePathsToDelete); err != nil {
			return err
		}
		store.RemoveDeletedFilesFromCache()
	}

	fmt.Println("Deployed Successfully!")
	return nil
}

type delete struct {
	XMLName xml.Name `xml:"Delete"`
	Objects []object
}

type object struct {
	XMLName xml.Name `xml:"Object"`
	Key     string   `xml:"Key"`
}

func deleteFilesOnS3(amazonBucketURL string, filePathsToDelete []string) error {
	deleteEndpoint := fmt.Sprintf("%s?delete", amazonBucketURL)

	// Create XML payload
	objects := []object{}
	for _, key := range filePathsToDelete {
		objects = append(objects, object{
			Key: key,
		})
	}
	deleteXMLObj := delete{
		Objects: objects,
	}
	deleteXML, err := xml.Marshal(deleteXMLObj)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	body := xml.Header + string(deleteXML)
	bodyAsBuffer := bytes.NewBuffer([]byte(body))

	// Make request
	r, err := http.NewRequest("POST", deleteEndpoint, bodyAsBuffer)
	if err != nil {
		return err
	}
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	// Calculate MD5 of request body & set as a request header
	md5Hash := md5.New()
	io.WriteString(md5Hash, body)
	contentMd5 := base64.StdEncoding.EncodeToString(md5Hash.Sum(nil))
	r.Header.Set("Content-MD5", contentMd5)

	s3.DefaultService.Sign(r, *s3util.DefaultConfig.Keys)

	// Make request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("Got non-200 response: %d", resp.StatusCode)
	}

	for _, filePath := range filePathsToDelete {
		fmt.Println("Deleted: ", filePath)
	}
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

func createS3Uploader(s3FilePath string) (w io.WriteCloser, err error) {
	header := make(http.Header)
	header.Add("x-amz-acl", "public-read")
	w, err = s3util.Create(s3FilePath, header, nil)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	return w, nil
}
