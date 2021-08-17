// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to S3.
func New() (uploader.Uploader, error) {
	u := s3Uploader{}
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	u.client = s3.New(sess)
	u.upload = s3manager.NewUploader(sess)
	return &u, nil
}

type s3Uploader struct {
	client *s3.S3
	upload *s3manager.Uploader
}

func (u *s3Uploader) Upload(options uploader.Options) error {
	_, err := u.upload.UploadWithContext(options.Context, &s3manager.UploadInput{
		Bucket:       aws.String(options.Bucket),
		Key:          aws.String(options.Path),
		Body:         options.Body,
		ContentType:  options.ContentType,
		CacheControl: options.CacheControl,
	})
	return err
}

func (u *s3Uploader) Delete(options uploader.Options) error {
	_, err := u.client.DeleteObjectWithContext(options.Context, &s3.DeleteObjectInput{
		Bucket: aws.String(options.Bucket),
		Key:    aws.String(options.Path),
	})
	return err
}
