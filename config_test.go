package main

import "testing"

func TestGetS3InfoWithNoConfigPresent(t *testing.T) {
	fName := "my-file.zip"
	cfg := getTestConfig()

	var b, k *string

	// should return nil if the S3 info is missing
	b, k = cfg.getS3Info(fName)
	if b != nil || k != nil {
		t.Log("should have returned two nils")
		t.Fail()
	}

}

func TestGetS3InfoWithOnlyBucketName(t *testing.T) {
	var b, k *string
	fName := "my-file.zip"

	cb := "some-bucket"
	//ck := "some-key"
	// should return bucket + `fName`
	// if given a bucket
	// but no key

	cfg := getTestConfig()
	cfg.Location.S3Bucket = &cb
	b, k = cfg.getS3Info(fName)
	if b == nil || k == nil {
		t.Error("should not have returned two nils!")
	}

	if *b != cb || *k != fName {
		t.Logf(
			"Unexpected return values: bucket = %s, key = %s",
			*b, *k,
		)
		t.Fail()
	}
}

func TestGetS3InfoWithKeyAsFolderName(t *testing.T) {
	var b, k *string
	fName := "my-file.zip"

	cb := "some-bucket"
	ck := "some-key"

	cfg := getTestConfig()
	cfg.Location.S3Bucket = &cb
	cfg.Location.S3Key = &ck

	b, k = cfg.getS3Info(fName)
	if b == nil || k == nil {
		t.Error("should not have returned two nils!")
	}

	if *b != cb || *k != "some-key/my-file.zip" {
		t.Logf(
			"Unexpected return values: bucket = %s, key = %s",
			*b, *k,
		)
		t.Fail()
	}
}

func TestGetS3InfoWithFullyQualifiedKey(t *testing.T) {
	var b, k *string
	fName := "my-file.zip"

	cb := "some-bucket"
	ck := "some-key/ihaveanotherfilename.zip"

	cfg := getTestConfig()
	cfg.Location.S3Bucket = &cb
	cfg.Location.S3Key = &ck

	b, k = cfg.getS3Info(fName)
	if b == nil || k == nil {
		t.Error("should not have returned two nils!")
	}

	if *b != cb || *k != ck {
		t.Logf(
			"Unexpected return values: bucket = %s, key = %s",
			*b, *k,
		)
		t.Fail()
	}
}

func getTestConfig() *Config {
	name := "some-func"
	desc := "description"

	return &Config{
		Name:        &name,
		Description: &desc,
		Location: &struct {
			S3Bucket        *string
			S3Key           *string
			S3ObjectVersion *string
		}{},
	}
}
