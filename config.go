package main

import "fmt"
import "github.com/aws/aws-sdk-go/service/iam"

/*

# lambda-phage config file sample

name: my-first-lambda-function
description: provides some sample stuff
pkg:
  name: my-first-lambda-function.zip
deploy:
  type: s3
  s3-bucket: test-bucket
  use-versioning: true
*/

type Config struct {
	Name        *string
	Description *string
	Archive     *string
	EntryPoint  *string
	MemorySize  *uint
	Runtime     *string
	Timeout     *uint
	IamRole     struct {
		Arn  *string
		Name *string
	}
	Location *struct {
		S3Bucket        *string
		S3Key           *string
		S3ObjectVersion *string
	}
}

// returns the arn for the role specified
func (c Config) getRoleArn() (*string, error) {
	// if the config file has an ARN listed,
	// that takes precedence
	if c.IamRole.Arn != "" {
		return c.IamRole.Arn, nil
	} else if c.IamRole.Name != "" {
		// look up the iam role name
		iamRole, err := getIamPolicy(c.IamRole.Name)
		if err != nil {
			return nil, err
		}
		return iamRole, err
	} else {
		// TODO: create a default standard role
		// and update config file
	}

	return nil, fmt
}

// gets an IAM policy by name
func getIamPolicy(name string) (*string, error) {
	i := iam.New(nil)

	r, err := i.GetRole(&iam.GetRoleInput{
		RoleName: &name,
	})

	if err != nil {
		return aws.String(""), err
	}

	return r.Role.ARN, nil
}
