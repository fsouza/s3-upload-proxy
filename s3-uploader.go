// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration of the s3-uploader.
type Config struct {
	BucketName      string            `envconfig:"BUCKET_NAME" required:"true"`
	HealthcheckPath string            `envconfig:"HEALTHCHECK_PATH" default:"/healthcheck"`
	HTTPPort        int               `envconfig:"HTTP_PORT" default:"80"`
	LogLevel        string            `envconfig:"LOG_LEVEL" default:"debug"`
	FastlySurrogate bool              `envconfig:"FASTLY_SURROGATE_CACHE"`
	CacheControl    cacheControlRules `envconfig:"CACHE_CONTROL_RULES"`
}

func loadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}

func (c *Config) logger() *logrus.Logger {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	logger := logrus.New()
	logger.Level = level
	return logger
}

func (c *Config) addCacheMetadata(input *s3manager.UploadInput) {
	if value := c.CacheControl.headerValue(aws.StringValue(input.Key)); value != nil {
		if c.FastlySurrogate {
			input.Metadata["surrogate-control"] = value
		} else {
			input.CacheControl = value
		}
	}
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	logger := cfg.logger()

	sess, err := session.NewSession()
	if err != nil {
		logger.WithError(err).Fatal("failed to load aws auth config")
	}
	uploader := s3manager.NewUploader(sess)
	http.HandleFunc(cfg.HealthcheckPath, healthcheck)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "POST" && r.Method != "PUT" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := strings.TrimLeft(r.URL.Path, "/")
		contentType := mime.TypeByExtension(filepath.Ext(key))
		logFields := logrus.Fields{"bucket": cfg.BucketName, "objectKey": key}
		input := s3manager.UploadInput{
			Bucket:      aws.String(cfg.BucketName),
			Key:         aws.String(key),
			Body:        r.Body,
			ContentType: aws.String(contentType),
			Metadata:    make(map[string]*string),
		}
		cfg.addCacheMetadata(&input)
		start := time.Now()
		_, err = uploader.Upload(&input)
		if err != nil {
			logger.WithFields(logFields).WithError(err).Error("failed to upload file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.WithFields(logFields).WithField("contentType", contentType).Debugf("finished upload in %s", time.Since(start))
		fmt.Fprintln(w, "OK")
	})

	listenAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to start listener")
	}
	defer listener.Close()
	logger.Infof("listening on %s", listener.Addr())
	http.Serve(listener, nil)
}
