package main

import "fmt"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/gopkg.in/yaml.v2"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/iam"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "strings"
import "io/ioutil"
import "os"

/*

# lambda-phage config file sample

name: my-first-lambda-function
description: provides some sample stuff
archive: my-first-lambda-function.zip
entryPoint: "index.handler"
memorySize: 128
runtime: nodejs
timeout: 5
regions:
  - us-east-1
iamPolicy:
  # arn:
  name: lambda_basic_execution
location:
  s3bucket: test-bucket
  s3key: my-first-function/
  s3region: us-east-1
  #s3ObjectVersion
*/

type IamRole struct {
	Arn  *string `yaml:"arn"`
	Name *string `yaml:"name"`
}

type Location struct {
	S3Bucket        *string
	S3Key           *string
	S3ObjectVersion *string
	S3Region        *string
}

type Config struct {
	// stuff for YAML below:
	Name         *string
	Project      string
	Arn          *string
	Description  *string
	Archive      *string
	EntryPoint   *string `yaml:"entryPoint"`
	MemorySize   *int64  `yaml:"memorySize"`
	Runtime      *string
	Timeout      *int64
	Regions      []*string `yaml:"regions"`
	IamRole      *IamRole  `yaml:"iamRole"`
	Location     *Location
	EventSources []*EventSource `yaml:"eventSources"`
}

type EventSource struct {
	Type               string
	ApiName            *string `yaml:"apiName"`
	ApiDeploymentStage *string `yaml:"apiDeploymentStage"`
	ApiSecurity        *string `yaml:"apiSecurity"`
}

// writes this config to a YAML file
func (c *Config) writeToFile(fName string) error {
	d, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fName, d, os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

// copies properties from all the config objects into
// the receiver config object
func (c *Config) merge(others ...*Config) *Config {
	for _, other := range others {
		c.Name = other.Name
		c.Project = other.Project
		c.Arn = other.Arn
		c.Description = other.Description
		c.Archive = other.Archive
		c.EntryPoint = other.EntryPoint
		c.MemorySize = other.MemorySize
		c.Runtime = other.Runtime
		c.Timeout = other.Timeout
		c.Regions = other.Regions
		c.IamRole = other.IamRole
		c.Location = other.Location
		c.EventSources = other.EventSources
	}
	return c
}

// returns the arn for the role specified
func (c *Config) getRoleArn() (*string, error) {
	if c.IamRole.Arn == nil &&
		c.IamRole.Name == nil {
		return nil, fmt.Errorf("Missing ARN config!")
	}

	// if the config file has an ARN listed,
	// that takes precedence
	if c.IamRole.Arn != nil {
		return c.IamRole.Arn, nil
	} else if c.IamRole.Name != nil {
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
	debug := debug.Debug("getIamPolicy")
	i := iam.New(nil)

	debug("getting iam role")

	r, err := i.GetRole(&iam.GetRoleInput{
		RoleName: &name,
	})

	if err != nil {
		return aws.String(""), err
	}

	return r.Role.Arn, nil
}
