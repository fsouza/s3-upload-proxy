// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mediastore

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediastoredata"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestResolveForMediaStoreData(t *testing.T) {
	const expectedSigningName = "mediastore"
	r := endpointResolver{
		dataEndpoint: "https://abc123hf.data.mediastore.us-east-1.amazonaws.com",
		resolver:     endpoints.NewDefaultResolver(),
	}
	endpoint, err := r.ResolveEndpoint(mediastoredata.ServiceName, "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.URL != r.dataEndpoint {
		t.Errorf("wrong endpoint url\nwant %q\ngot  %q", r.dataEndpoint, endpoint.URL)
	}
	if endpoint.SigningName != expectedSigningName {
		t.Errorf("wrong signing name\nwant %q\ngot  %q", expectedSigningName, endpoint.SigningName)
	}
}

func TestResolveForMediaStore(t *testing.T) {
	defaultResolver := endpoints.NewDefaultResolver()
	expectedEndpoint, err := defaultResolver.ResolveEndpoint(mediastore.ServiceName, "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	r := endpointResolver{
		dataEndpoint: "https://abc123hf.data.mediastore.us-east-1.amazonaws.com",
		resolver:     defaultResolver,
	}
	endpoint, err := r.ResolveEndpoint(mediastore.ServiceName, "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(endpoint, expectedEndpoint) {
		t.Errorf("resolver shouldn't touch endpoint for services that aren't mediastoredata\nservice name: %q\nwant %#v\ngot  %#v", mediastore.ServiceName, expectedEndpoint, endpoint)
	}
}

func TestResolverForS3(t *testing.T) {
	defaultResolver := endpoints.NewDefaultResolver()
	expectedEndpoint, err := defaultResolver.ResolveEndpoint(s3.ServiceName, "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	r := endpointResolver{
		dataEndpoint: "https://abc123hf.data.mediastore.us-east-1.amazonaws.com",
		resolver:     defaultResolver,
	}
	endpoint, err := r.ResolveEndpoint(s3.ServiceName, "us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(endpoint, expectedEndpoint) {
		t.Errorf("resolver shouldn't touch endpoint for services that aren't mediastoredata\nservice name: %q\nwant %#v\ngot  %#v", s3.ServiceName, expectedEndpoint, endpoint)
	}
}
