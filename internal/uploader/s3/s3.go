// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to S3.
func New() (uploader.Uploader, error) {
	var u s3Uploader
	sess, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	u.uploader = s3manager.NewUploader(sess)
	return &u, nil
}

type s3Uploader struct {
	uploader *s3manager.Uploader
}

func (u *s3Uploader) Upload(options uploader.Options) error {
	input := s3manager.UploadInput{
		Bucket:       aws.String(options.BucketName),
		Key:          aws.String(options.Path),
		Body:         options.Body,
		ContentType:  aws.String(options.ContentType),
		CacheControl: aws.String(options.CacheControl),
	}
	_, err := u.uploader.Upload(&input)
	return err
}
