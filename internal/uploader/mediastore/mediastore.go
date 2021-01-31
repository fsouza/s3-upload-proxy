// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata/types"
	awsint "github.com/fsouza/s3-upload-proxy/internal/aws"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// DriverOptions are the set of options that can change how the mediastore
// driver behaves.
type DriverOptions struct {
	ChunkedTransfer bool
}

// New returns an uploader that sends objects to Elemental MediaStore.
func New(options DriverOptions) (uploader.Uploader, error) {
	u := msUploader{uploadAvailability: types.UploadAvailabilityStandard}
	if options.ChunkedTransfer {
		u.uploadAvailability = types.UploadAvailabilityStreaming
	}
	sess, err := awsint.Config()
	if err != nil {
		return nil, err
	}
	u.client = mediastore.NewFromConfig(sess)
	return &u, nil
}

type msUploader struct {
	client             *mediastore.Client
	containers         sync.Map
	uploadAvailability types.UploadAvailability
}

func (u *msUploader) Upload(options uploader.Options) error {
	client, err := u.getDataClientForContainer(options.Bucket)
	if err != nil {
		return err
	}
	input := mediastoredata.PutObjectInput{
		Path:               aws.String(options.Path),
		ContentType:        options.ContentType,
		CacheControl:       options.CacheControl,
		Body:               options.Body,
		UploadAvailability: u.uploadAvailability,
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
