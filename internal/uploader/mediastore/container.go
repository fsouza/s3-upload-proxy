// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
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
	resp, err := u.client.DescribeContainer(&mediastore.DescribeContainerInput{
		ContainerName: aws.String(containerName),
	})
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(aws.NewConfig().WithEndpoint(*resp.Container.Endpoint))
	if err != nil {
		return nil, err
	}
	client := mediastoredata.New(sess)
	return client, nil
}
