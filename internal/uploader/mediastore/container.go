// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	awsint "github.com/fsouza/s3-upload-proxy/internal/aws"
)

func (u *msUploader) getDataClientForContainer(name string) (*mediastoredata.Client, error) {
	v, ok := u.containers.Load(name)
	if !ok {
		client, err := u.newDataClient(name)
		if err != nil {
			return nil, err
		}
		v = client
		u.containers.Store(name, v)
	}
	return v.(*mediastoredata.Client), nil
}

func (u *msUploader) newDataClient(containerName string) (*mediastoredata.Client, error) {
	resp, err := u.client.DescribeContainer(context.Background(), &mediastore.DescribeContainerInput{
		ContainerName: aws.String(containerName),
	})
	if err != nil {
		return nil, err
	}
	cfg, err := awsint.Config()
	if err != nil {
		return nil, err
	}
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(string, string) (aws.Endpoint, error) {
		return aws.Endpoint{URL: *resp.Container.Endpoint}, nil
	})
	client := mediastoredata.NewFromConfig(cfg)
	return client, nil
}
