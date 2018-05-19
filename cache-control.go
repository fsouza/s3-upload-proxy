// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"regexp"
)

type cacheControlRules []cacheControlRule

type cacheControlRule struct {
	Regexp *jregexp `json:"regexp"`
	Value  string   `json:"value"`
}

func (c *cacheControlRules) Set(value string) error {
	err := json.Unmarshal([]byte(value), c)
	return err
}

func (c cacheControlRules) headerValue(file string) *string {
	for _, rule := range c {
		if rule.Regexp.re.MatchString(file) {
			cacheControl := rule.Value
			return &cacheControl
		}
	}

	return nil
}

type jregexp struct {
	re *regexp.Regexp
}

func (r *jregexp) UnmarshalJSON(data []byte) (err error) {
	expr := string(bytes.Trim(data, `"`))
	r.re, err = regexp.Compile(expr)
	return err
}
