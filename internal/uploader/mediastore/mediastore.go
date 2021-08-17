// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// DriverOptions are the set of options that can change how the mediastore
// driver behaves.
type DriverOptions struct {
	ChunkedTransfer bool
}

// New returns an uploader that sends objects to Elemental MediaStore.
func New(options DriverOptions) (uploader.Uploader, error) {
	u := msUploader{uploadAvailability: mediastoredata.UploadAvailabilityStandard}
	if options.ChunkedTransfer {
		u.uploadAvailability = mediastoredata.UploadAvailabilityStreaming
	}
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	u.client = mediastore.New(sess)
	return &u, nil
}

type msUploader struct {
	client             *mediastore.MediaStore
	containers         sync.Map
	uploadAvailability string
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
		Body:               aws.ReadSeekCloser(options.Body),
		UploadAvailability: aws.String(u.uploadAvailability),
	}
	_, err = client.PutObjectWithContext(options.Context, &input)
	return err
}

func (u *msUploader) Delete(options uploader.Options) error {
	client, err := u.getDataClientForContainer(options.Bucket)
	if err != nil {
		return err
	}
	input := mediastoredata.DeleteObjectInput{Path: aws.String(options.Path)}
	_, err = client.DeleteObjectWithContext(options.Context, &input)
	return err
}
