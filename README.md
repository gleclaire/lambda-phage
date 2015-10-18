# lambda-phage
a tool for deploying to aws lambda

## Installation

```sh
go get github.com/hopkinsth/lambda-phage
```

## Usage

```
Usage:
  lambda-phage [command]

Available Commands:
  deploy      deploys your lambda function according to your config file or options provided
  init        initializes a config for your function
  pkg         adds all the current folder to a zip file recursively

Flags:
  -c, --config="l-p.yml": config file location
  -v, --verbose[=false]: verbosity

Use "lambda-phage [command] --help" for more information about a command.
```
## Configuration

This tool stores its config in a YAML file named `l-p.yml` by default. Here's a quick sample:

```yaml
name: my-first-lambda-function
description: provides some sample stuff
archive: my-first-lambda-function.zip
entryPoint: index.handler
memorySize: 128
runtime: nodejs
timeout: 5
regions: [us-east-1]
iamRole:
  # if present, 
  # ARN takes precedence over name
  arn: aws
  name: lambda_basic_execution
location:
  # omit S3 configuration to upload
  # directly to Lambda
  s3bucket: test-bucket
  s3key: my-first-function/
  s3region: us-east-1
  s3ObjectVersion: myversion
```
