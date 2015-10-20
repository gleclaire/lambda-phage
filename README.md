# lambda-phage
a tool for deploying to aws lambda

## Installation

```sh
go get github.com/hopkinsth/lambda-phage
```

## Getting Started

lambda-phage is a tool that aims to make deploying code to AWS Lambda easier without needing CloudFormation et al. To get started, run `init` inside the folder that has the code you want to run in Lambda:

```sh
$ lambda-phage init
```

The `init` command will ask you several questions about your AWS setup and save a configuration file (named `l-p.yml` by default) to keep the options you define. 

Once you've set configuration, all you need to do is run the `deploy` command to deploy:

```sh
$ lambda-phage deploy
```

This will package all the files in the current directory, recursively, into a ZIP file and upload it to AWS Lambda.

lambda-phage uses Amazon's official Go SDK, so you can set API credentials in any way that the Go SDK supports, including a local credential file and environment variables. If you're getting set up for the first time, please read [Amazon's guide to setting up their own AWS CLI tools](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files), which describes how to set up a local credential file.

## Project Support

lambda-phage can organize your lambda functions into projects for easier deployment. To create a project, use the `project create` command:

```sh
$ lambda-phage project create my-project-name
```

You can then add the current working directory to the project:

```sh
$ lambda-phage project add my-project-name
```

Finally, you can deploy all the functions in a project with `project deploy`:

```sh
# deploy all functions in the project
$ lambda-phage project deploy my-project-name

# deploy functions matching a pattern
$ lambda-phage project deploy --filter '.*frontend' my-project-name

# perform a project deploy dry run, only printing what would be deployed
# rather than actually deploying it
$ lambda-phage project deploy --dry-run my-project-name
```

## TODO:
- real tests (tdd be damned i guess?)
- add `project import` command to interactively load suggested project names based on whatever's in config YAML
- add support for configuring event sources through `init`
  - potential blocker is lack of API gateway API client, though we might be able to do it anyway
  - can at least support streams (dynamoDB + kinesis), s3, cloudwatch event sources

## Usage

```

     ,-^-.
     |\/\|
     '-V-'
       H
       H
  itz  H
    .-;":-.
   ,'|  '; \

Usage:
  lambda-phage [command]

Available Commands:
  deploy      deploys your lambda function according to your config file or options provided
  init        initializes a config for your function
  pkg         adds all the current folder to a zip file recursively
  project     does project stuff

Flags:
  -c, --config="l-p.yml": config file location
  -v, --verbose[=false]: verbosity

Use "lambda-phage [command] --help" for more information about a command.
```

## Configuration Format

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
