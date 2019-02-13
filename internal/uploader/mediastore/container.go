// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
)

func (u *msUploader) getDataClientForContainer(name string) (*mediastoredata.MediaStoreData, error) {
	v, ok := u.containers.Load(name)
	if !ok {
		client, err := u.newDataClient(name)
		if err != nil {
			return nil, err
		}
		v = client
		u.containers.Store(name, v)
	}
	return v.(*mediastoredata.MediaStoreData), nil
}

func (u *msUploader) newDataClient(containerName string) (*mediastoredata.MediaStoreData, error) {
	req := u.client.DescribeContainerRequest(&mediastore.DescribeContainerInput{
		ContainerName: aws.String(containerName),
	})
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}
	endpoint := aws.StringValue(resp.Container.Endpoint)
	sess, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	sess.EndpointResolver = aws.ResolveWithEndpointURL(endpoint)
	client := mediastoredata.New(sess)
	return client, nil
}
