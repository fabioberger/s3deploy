<a href="https://godoc.org/github.com/fabioberger/s3deploy" ><img src="http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square" /></a>


S3 Deploy
--------------

Do you have a static web-app you'd like to deploy to S3 with one command? S3Deploy will sync your local project with an S3 bucket, only uploading the files that have changed, and removing files that you've deleted locally. This is essentially a simpler version of the [AWS CLI sync command](http://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).

# How to install:

1. Download this repo:<br>
```
go get -u github.com/fabioberger/s3deploy
```

2. Enter the project root directory:<br>
```
cd $GOPATH/src/github.com/fabioberger/s3deploy
```

3. Install the binary:<br>
```
go install
```

4. Set your Amazon AWS credentials as environment variables:
```
export S3_ACCESS_KEY=YOUR_S3_ACCESS_KEY
```
```
export S3_SECRET_KEY=YOUR_S3_SECRET_KEY
```
<br>
replacing YOUR_S3_ACCESS_KEY and YOUR_S3_SECRET_KEY with your keys from [https://console.aws.amazon.com/iam/home?#security_credential](https://console.aws.amazon.com/iam/home?#security_credential)

5. Enter any project directory you would like to deploy and run s3deploy:
```
cd ~/path/to/my/project
```
```
s3deploy --bucket S3_BUCKET_NAME --region s3-eu-west-1
```
<br>
Where you replace S3_BUCKET_NAME with the bucket you'd like to deploy the project to. Make sure this bucket is already created on S3 and that you specify the buckets region.

For more options, check out:

```
s3deploy --help
```

**Note**: Make sure you're GOPATH bin is in your $PATH variable. Try:
```
export $PATH=$PATH:$GOPATH/bin
```
If you're not sure if you do!
<br><br>
Dedicated to: Marlene, I hope this tool helps you learn to code faster ;)
