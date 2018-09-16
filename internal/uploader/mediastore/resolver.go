// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
)

type endpointResolver struct {
	dataEndpoint string
	resolver     aws.EndpointResolver
}

func (r *endpointResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	endpoint, err := r.resolver.ResolveEndpoint(service, region)
	if err != nil {
		return aws.Endpoint{}, err
	}
	if service != mediastoredata.ServiceName {
		return endpoint, nil
	}
	endpoint.URL = r.dataEndpoint
	endpoint.SigningName = "mediastore"
	return endpoint, nil
}
