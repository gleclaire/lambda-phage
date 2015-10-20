package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/lambda"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/s3"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/awserr"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "os"
import "io"
import "io/ioutil"
import "fmt"
import "bytes"

//import "github.com/aws/aws-sdk-go/aws"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"

type wtflog struct{}

func (w wtflog) Log(s ...interface{}) {
	fmt.Println(s...)
}

func init() {
	dCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploys your lambda function according to your config file or options provided",
		RunE:  deploy,
	}

	cmds = append(cmds, dCmd)
}

func deploy(c *cobra.Command, args []string) error {
	debug := debug.Debug("cmd.deploy")
	// must package first, so:
	var err error
	err = pkg(c, args)

	if err != nil {
		return err
	}

	// should have written data to this file
	binName := getArchiveName(c)

	var iamRole *string
	iamRole, err = cfg.getRoleArn()
	if err != nil {
		return err
	}

	code := &lambda.FunctionCode{}
	// try getting s3 information
	bucket, key := cfg.getS3Info(binName)

	if bucket == nil || key == nil {
		fmt.Println("Uploading function to Lambda")
		// if we couldn't get bucket or
		// key info, let's upload the data
		// ...soon
		b, err := ioutil.ReadFile(binName)
		if err != nil {
			return err
		}

		code.ZipFile = b
	} else {
		fmt.Println("Uploading function to s3")
		code.S3Bucket = bucket
		code.S3Key = key
		err = uploadS3(binName, bucket, key)
		if err != nil {
			return err
		}
	}

	debug("preparing for lambda API")
	for _, region := range cfg.Regions {
		fmt.Printf("Deploying lambda function for %s\n", *region)

		l := lambda.New(
			aws.NewConfig().
				WithRegion(*region),
		)
		//just try creating the function now
		_, err := l.CreateFunction(
			&lambda.CreateFunctionInput{
				Code:         code,
				FunctionName: cfg.Name,
				Description:  cfg.Description,
				Handler:      cfg.EntryPoint,
				MemorySize:   cfg.MemorySize,
				Runtime:      cfg.Runtime,
				Timeout:      cfg.Timeout,
				Role:         iamRole,
				Publish:      aws.Bool(true),
			},
		)

		if err != nil {
			if awe, ok := err.(awserr.Error); ok {
				if awe.Code() == "ResourceConflictException" {
					debug("function already exists, calling update")
					err = updateLambda(l, code, iamRole)
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			// if the create function succeeded,
			// we need to figure out the mapping
			// info, etc
			debug("function creation succeeded! ...we think")
		}

		if err != nil {
			return err
		}

		fmt.Printf("Function %s deployed to region %s!\n", *cfg.Name, *region)
	}

	return nil
}

// updates function code, settings in AWS lambda
func updateLambda(
	l *lambda.Lambda,
	code *lambda.FunctionCode,
	iamRole *string,
) error {
	_, err := l.UpdateFunctionConfiguration(
		&lambda.UpdateFunctionConfigurationInput{
			FunctionName: cfg.Name,
			Description:  cfg.Description,
			Handler:      cfg.EntryPoint,
			MemorySize:   cfg.MemorySize,
			Role:         iamRole,
			Timeout:      cfg.Timeout,
		},
	)
	if err != nil {
		return err
	}

	_, err = l.UpdateFunctionCode(
		&lambda.UpdateFunctionCodeInput{
			FunctionName:    cfg.Name,
			S3Bucket:        code.S3Bucket,
			S3Key:           code.S3Key,
			S3ObjectVersion: code.S3ObjectVersion,
			ZipFile:         code.ZipFile,
			Publish:         aws.Bool(true),
		},
	)

	if err != nil {
		return err
	}
	return nil
}

// should probably handle this in memory,
// but this fn will upload your file to s3
func uploadS3(fName string, bucket, key *string) error {
	debug := debug.Debug("uploadS3")
	awscfg := aws.NewConfig().WithRegion(*cfg.Location.S3Region)
	s := s3.New(awscfg)

	f, err := os.Open(fName)
	if err != nil {
		return err
	}
	debug("opened file, beginning read")

	st, err := f.Stat()
	if err != nil {
		return err
	}

	if st.Size() >= (1024 * 1024 * 5) {
		// multipart if bigger than 5MiB
		return uploadS3MPU(s, f, bucket, key)
	} else {
		_, err = s.PutObject(&s3.PutObjectInput{
			Bucket: bucket,
			Key:    key,
			Body:   f,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// multipart upload
func uploadS3MPU(
	s *s3.S3,
	f *os.File,
	bucket, key *string,
) error {
	debug := debug.Debug("uploadS3MPU")
	mpu := &s3.CreateMultipartUploadInput{
		Bucket: bucket,
		Key:    key,
	}

	canceler := func(uploadId *string) {
		debug("canceling upload")
		_, err := s.AbortMultipartUpload(
			&s3.AbortMultipartUploadInput{
				Bucket:   bucket,
				Key:      key,
				UploadId: uploadId,
			},
		)
		if err != nil {
			fmt.Printf(
				"WARNING: upload abort failed for upload ID %s",
				*uploadId,
			)
		}
	}

	cr, err := s.CreateMultipartUpload(mpu)
	if err != nil {
		return err
	}

	debug("created multipart upload")

	buf := new(bytes.Buffer)
	bslice := make([]byte, 8096)
	var pNum int64 = 1

	parts := make([]*s3.CompletedPart, 0)

	for {
		n, err := f.Read(bslice)

		isEOF := err != nil && err == io.EOF

		if err != nil && !isEOF {
			f.Close()
			canceler(cr.UploadId)
			return err
		}

		if isEOF {
			debug("reached end of file")
			// trim final content
			bslice = bslice[:n]
		}

		buf.Write(bslice)

		// drain buf on 1MiB of data
		if buf.Len() >= (1024*1024*5) || isEOF {
			debug("have file data, uploading chunk")
			var err error
			if buf.Len() > 0 {
				var p *s3.UploadPartOutput
				p, err = s.UploadPart(
					&s3.UploadPartInput{
						Bucket:     bucket,
						Key:        key,
						PartNumber: aws.Int64(pNum),
						UploadId:   cr.UploadId,
						Body:       bytes.NewReader(buf.Bytes()),
					},
				)

				if err != nil {
					// if uploading a part failed,
					// cancel the whole upload...
					f.Close()
					canceler(cr.UploadId)
					return err
				}

				parts = append(parts,
					&s3.CompletedPart{
						ETag:       p.ETag,
						PartNumber: aws.Int64(pNum),
					},
				)
			}

			pNum += 1
			buf.Reset()

			if isEOF {
				f.Close()
				debug("completing upload")
				iput := &s3.CompleteMultipartUploadInput{
					Bucket:   bucket,
					Key:      key,
					UploadId: cr.UploadId,
					MultipartUpload: &s3.CompletedMultipartUpload{
						Parts: parts,
					},
				}
				_, err := s.CompleteMultipartUpload(
					iput,
				)

				if err != nil {
					canceler(cr.UploadId)
					return err
				}

				break
			}
		}
	}
	return nil
}
