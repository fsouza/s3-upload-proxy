// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cachecontrol

import (
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/kelseyhightower/envconfig"
)

func TestCacheControlRulesCanBeLoadedFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("RULES", `[{"regexp":".mp4$","value":"public, max-age=123456"},{"regexp":".html$","value":"public, max-age=60"}]`)
	var value struct {
		Rules Rules `envconfig:"RULES"`
	}
	t.Log(os.Getenv("RULES"))
	expectedRules := map[string]Rule{
		regexp.MustCompile(`.mp4$`).String():  {Value: "public, max-age=123456"},
		regexp.MustCompile(`.html$`).String(): {Value: "public, max-age=60"},
	}
	err := envconfig.Process("", &value)
	if err != nil {
		t.Fatal(err)
	}
	gotRules := map[string]Rule{}
	for _, r := range value.Rules {
		re := r.Regexp.re
		r.Regexp = nil
		gotRules[re.String()] = r
	}
	if !reflect.DeepEqual(gotRules, expectedRules) {
		t.Errorf("wrong rules returned\nwant %#v\ngot  %#v", expectedRules, gotRules)
	}
}

func TestCacheControlRulesInvalidJSON(t *testing.T) {
	os.Clearenv()
	os.Setenv("RULES", `[{"regexp:".mp4"},{"regexp":".html",`)
	var value struct {
		Rules Rules `envconfig:"RULES"`
	}
	err := envconfig.Process("", &value)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestCacheControlHeaderValue(t *testing.T) {
	rules := Rules{
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`\.mp4$`)}, Value: "public, max-age=123456"},
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`\.html$`)}, Value: "public, max-age=60"},
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`master_.+\.m3u8$`)}, Value: "private"},
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`master\.m3u8$`)}, Value: "public, max-age=1"},
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`\.webm$`)}, Value: "public, max-age=2, s-maxage=123456"},
		Rule{Regexp: &jregexp{re: regexp.MustCompile(`\.mp3$`)}, Value: "public, s-maxage=123456"},
	}
	tests := []struct {
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
			aws.String("public, max-age=1"),
		},
		{
			"video/master_720p.m3u8",
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
		test := test
		t.Run(test.input, func(t *testing.T) {
			value := rules.HeaderValue(test.input)
			if !reflect.DeepEqual(value, test.expected) {
				t.Errorf("wrong value returned\nwant %#v\ngot  %#v", test.expected, value)
			}
		})
	}
}
