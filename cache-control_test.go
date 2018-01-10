// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kelseyhightower/envconfig"
)

func TestCacheControlRulesCanBeLoadedFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("RULES", `[{"ext":".mp4","maxAge":123456},{"ext":".html","maxAge":60}]`)
	var value struct {
		Rules cacheControlRules `envconfig:"RULES"`
	}
	expectedRules := cacheControlRules{
		cacheControlRule{Extension: ".mp4", MaxAge: 123456},
		cacheControlRule{Extension: ".html", MaxAge: 60},
	}
	err := envconfig.Process("", &value)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.Rules, expectedRules) {
		t.Errorf("wrong rules returned\nwant %#v\ngot  %#v", expectedRules, value.Rules)
	}
}

func TestCacheControlRulesInvalidJSON(t *testing.T) {
	os.Clearenv()
	os.Setenv("RULES", `[{"ext":".mp4","maxAge":123456},{"ext":".html",`)
	var value struct {
		Rules cacheControlRules `envconfig:"RULES"`
	}
	err := envconfig.Process("", &value)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestCacheControlHeaderValue(t *testing.T) {
	rules := cacheControlRules{
		cacheControlRule{Extension: ".mp4", MaxAge: 123456},
		cacheControlRule{Extension: ".html", MaxAge: 60},
		cacheControlRule{Extension: ".m3u8", Private: true},
		cacheControlRule{Extension: ".webm", MaxAge: 2, SMaxAge: 123456},
		cacheControlRule{Extension: ".mp3", SMaxAge: 123456},
	}
	var tests = []struct {
		input    string
		expected *string
	}{
		{
			"https://github.com/some/file.mp4",
			aws.String("public, max-age=123456"),
		},
		{
			"file.mp4",
			aws.String("public, max-age=123456"),
		},
		{
			"some/path/index.html",
			aws.String("public, max-age=60"),
		},
		{
			"video/master.m3u8",
			aws.String("private"),
		},
		{
			"file.mp3",
			aws.String("public, s-maxage=123456"),
		},
		{
			"video.webm",
			aws.String("public, max-age=2, s-maxage=123456"),
		},
		{
			"some/path/audio.ogg",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			value := rules.headerValue(test.input)
			if !reflect.DeepEqual(value, test.expected) {
				t.Errorf("wrong value returned\nwant %#v\ngot  %#v", test.expected, value)
			}
		})
	}
}
