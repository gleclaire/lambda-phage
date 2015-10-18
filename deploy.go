package main

import "github.com/aws/aws-sdk-go/service/lambda"
import "github.com/aws/aws-sdk-go/service/s3"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/tj/go-debug"
import "os"
import "io"
import "io/ioutil"
import "fmt"
import "bytes"

//import "github.com/aws/aws-sdk-go/aws"
import "github.com/spf13/cobra"

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

	flg := dCmd.Flags()

	flg.BoolP("verbose", "v", false, "verbosity")
	cmds = append(cmds, dCmd)
}

func deploy(c *cobra.Command, args []string) error {
	// must package first, so:
	var err error
	err = pkg(c, args)

	if err != nil {
		return err
	}

	// should have written data to this file
	binName := getArchiveName(c)

	//l := lambda.New(nil)

	// var iamRole *string
	// iamRole, err = cfg.getRoleArn()
	// if err != nil {
	// 	return err
	// }

	code := &lambda.FunctionCode{}
	// try getting s3 information
	bucket, key := cfg.getS3Info(binName)
	if bucket == nil || key == nil {
		// if we couldn't get bucket or
		// key info, let's upload the data
		// ...soon
		panic("not implemented yet, sorry!")
	} else {
		code.S3Bucket = bucket
		code.S3Key = key
		err = uploadS3(binName, bucket, key)
		if err != nil {
			return err
		}
	}

	// just try creating the function now
	// r, err := l.CreateFunction(
	// 	&lambda.CreateFunctionInput{
	// 		Code: &lambda.FunctionCode{},
	// 	},
	// )

	return nil
}

// should probably handle this in memory,
// but this fn will upload your file to s3
func uploadS3(fName string, bucket, key *string) error {
	debug := debug.Debug("uploadS3")
	awscfg := aws.NewConfig().WithRegion(*cfg.Region)
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
		// just put object
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		_, err = s.PutObject(&s3.PutObjectInput{
			Bucket: bucket,
			Key:    key,
			Body:   bytes.NewReader(data),
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
