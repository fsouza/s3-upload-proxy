// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to Elemental MediaStore.
func New() (uploader.Uploader, error) {
	var u msUploader
	sess, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	u.client = mediastore.NewFromConfig(sess)
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
		Body:         options.Body,
	}
	_, err = client.PutObject(options.Context, &input)
	return err
}

func (u *msUploader) Delete(options uploader.Options) error {
	client, err := u.getDataClientForContainer(options.Bucket)
	if err != nil {
		return err
	}
	input := mediastoredata.DeleteObjectInput{Path: aws.String(options.Path)}
	_, err = client.DeleteObject(options.Context, &input)
	return err
}
