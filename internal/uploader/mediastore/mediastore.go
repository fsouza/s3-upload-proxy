// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to Elemental MediaStore.
func New() (uploader.Uploader, error) {
	var u msUploader
	sess, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	u.client = mediastore.New(sess)
	return &u, nil
}

type msUploader struct {
	client     *mediastore.Client
	containers sync.Map
}

func (u *msUploader) Upload(options uploader.Options) error {
	client, err := u.getDataClientForContainer(options.Bucket)
	if err != nil {
		return err
	}
	input := mediastoredata.PutObjectInput{
		Path:         aws.String(options.Path),
		ContentType:  options.ContentType,
		CacheControl: options.CacheControl,
		Body:         aws.ReadSeekCloser(options.Body),
	}
	req := client.PutObjectRequest(&input)
	_, err = req.Send(options.Context)
	return err
}

func (u *msUploader) Delete(options uploader.Options) error {
	client, err := u.getDataClientForContainer(options.Bucket)
	if err != nil {
		return err
	}
	input := mediastoredata.DeleteObjectInput{Path: aws.String(options.Path)}
	req := client.DeleteObjectRequest(&input)
	_, err = req.Send(options.Context)
	return err
}
