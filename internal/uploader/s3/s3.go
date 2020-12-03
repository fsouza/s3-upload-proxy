// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to S3.
func New() (uploader.Uploader, error) {
	u := s3Uploader{}
	sess, err := config.LoadDefaultConfig()
	if err != nil {
		return nil, err
	}
	u.client = s3.NewFromConfig(sess)
	u.upload = manager.NewUploader(u.client)
	return &u, nil
}

type s3Uploader struct {
	client *s3.Client
	upload *manager.Uploader
}

func (u *s3Uploader) Upload(options uploader.Options) error {
	_, err := u.upload.Upload(options.Context, &s3.PutObjectInput{
		Bucket:       aws.String(options.Bucket),
		Key:          aws.String(options.Path),
		Body:         options.Body,
		ContentType:  options.ContentType,
		CacheControl: options.CacheControl,
	})
	return err
}

func (u *s3Uploader) Delete(options uploader.Options) error {
	_, err := u.client.DeleteObject(options.Context, &s3.DeleteObjectInput{
		Bucket: aws.String(options.Bucket),
		Key:    aws.String(options.Path),
	})
	return err
}
