// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cachecontrol

import (
	"bytes"
	"encoding/json"
	"regexp"
)

// Rules is a list of cache control rules.
type Rules []Rule

// Rule is a mapping of regular expressions to cache-control string rules.
type Rule struct {
	Regexp *jregexp `json:"regexp"`
	Value  string   `json:"value"`
}

// Set loads the list of rules as a JSON-string, allowing values of type Rules
// to be used with envconfig.
func (c *Rules) Set(value string) error {
	err := json.Unmarshal([]byte(value), c)
	return err
}

// HeaderValue returns the matching cache control rule for the given file name.
func (c Rules) HeaderValue(fileName string) *string {
	for _, rule := range c {
		if rule.Regexp.re.MatchString(fileName) {
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
