package main

import "fmt"
import "github.com/aws/aws-sdk-go/service/iam"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/tj/go-debug"
import "strings"

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
func (c *Config) getRoleArn() (*string, error) {
	if c.IamRole.Arn == nil &&
		c.IamRole.Name == nil {
		return nil, fmt.Errorf("Missing ARN config!")
	}

	// if the config file has an ARN listed,
	// that takes precedence
	if *c.IamRole.Arn != "" {
		return c.IamRole.Arn, nil
	} else if *c.IamRole.Name != "" {
		// look up the iam role name
		iamRole, err := getIamPolicy(*c.IamRole.Name)
		if err != nil {
			return nil, err
		}
		return iamRole, err
	} else {
		// TODO: create a default standard role
		// and update config file
	}

	return nil, fmt.Errorf("how did you get here")
}

// returns a normalized S3 path to a file
// based on config information
//
// requires the file name of the archive you'll upload
//
// returns the bucket and the key
func (c *Config) getS3Info(fName string) (bucket, key *string) {
	debug := debug.Debug("config.getS3Info")
	loc := c.Location
	if loc == nil {
		debug("no upload location info found")
		return nil, nil
	}

	if loc.S3Bucket == nil {
		// TODO: make these return an error instead??
		debug("upload location info found, but s3 bucket missing")
		return nil, nil
	}

	b := *loc.S3Bucket
	var k string
	if loc.S3Key == nil {
		debug("no s3 key in location config, using file name")
		// no key in config?
		// then the key is the name of the file
		// being passed in
		k = fName
	} else if loc.S3Key != nil &&
		len(*loc.S3Key) > 0 {
		// key in config? let's see
		// if it looks like a zip file
		if strings.Index(*loc.S3Key, ".zip") > -1 {
			// great, we can use this one for the key
			k = *loc.S3Key
		} else {
			// if there's no .zip in the s3Key config
			// setting, then assume this is to
			// be the first part in a directory
			dir := *loc.S3Key
			sl := []byte("/")
			if dir[len(dir)-1] != sl[0] {
				dir += "/"
			}

			k = dir + fName
		}
	} else {
		debug("empty s3key found in config file")
		k = fName
	}

	return &b, &k
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
