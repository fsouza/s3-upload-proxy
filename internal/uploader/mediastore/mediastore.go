// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
)

// New returns an uploader that sends objects to Elemental MediaStore.
func New(dataEndpoint string) (uploader.Uploader, error) {
	var u msUploader
	sess, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	sess.EndpointResolver = &endpointResolver{
		dataEndpoint: dataEndpoint,
		resolver:     endpoints.NewDefaultResolver(),
	}
	u.client = mediastoredata.New(sess)
	return &u, nil
}

type msUploader struct {
	client *mediastoredata.MediaStoreData
}

func (u *msUploader) Upload(options uploader.Options) error {
	input := mediastoredata.PutObjectInput{
		Path:         aws.String(options.Path),
		Body:         nopSeeker{options.Body},
		ContentType:  aws.String(options.ContentType),
		CacheControl: aws.String(options.CacheControl),
	}
	req := u.client.PutObjectRequest(&input)
	_, err := req.Send()
	return err
}

type nopSeeker struct {
	io.Reader
}

func (nopSeeker) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}
