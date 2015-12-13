S3 Deploy
--------------

Do you have a static web-app you'd like to deploy to S3 with one command? S3Deploy is the tool for you!

How to install:

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

5. Enter any project directory you would like to deploy and run:
```
cd ~/path/to/my/project
```<br>
```
s3deploy S3_BUCKET_NAME
```
<br>
Where you replace S3_BUCKET_NAME with the bucket you'd like to deploy the project to. Make sure this bucket is already created on S3

**Note**: Make sure you're GOPATH bin is in your $PATH variable. Try:
```
export $PATH=$PATH:$GOPATH/bin
```<br>
If you're not sure if you do!

Dedicated to: Marlene, I hope this tool helps you learn to code faster ;)
