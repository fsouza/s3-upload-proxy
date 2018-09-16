// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsouza/s3-upload-proxy/internal/cachecontrol"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
	"github.com/fsouza/s3-upload-proxy/internal/uploader/mediastore"
	"github.com/fsouza/s3-upload-proxy/internal/uploader/s3"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config is the configuration of the s3-uploader.
type Config struct {
	BucketName      string             `envconfig:"BUCKET_NAME"`
	DataEndpoint    string             `envconfig:"MEDIASTORE_DATA_ENDPOINT"`
	UploadDriver    string             `envconfig:"UPLOAD_DRIVER" default:"s3"`
	HealthcheckPath string             `envconfig:"HEALTHCHECK_PATH" default:"/healthcheck"`
	HTTPPort        int                `envconfig:"HTTP_PORT" default:"80"`
	LogLevel        string             `envconfig:"LOG_LEVEL" default:"debug"`
	CacheControl    cachecontrol.Rules `envconfig:"CACHE_CONTROL_RULES"`
}

func loadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, err
	}
	if cfg.UploadDriver != "s3" && cfg.UploadDriver != "mediastore" {
		return cfg, errors.New(`invalid UPLOAD_DRIVER, valid options are "s3" and "mediastore"`)
	}
	if cfg.UploadDriver == "s3" && cfg.BucketName == "" {
		return cfg, errors.New("s3 upload driver requires the BUCKET_NAME")
	}
	if cfg.UploadDriver == "mediastore" && cfg.DataEndpoint == "" {
		return cfg, errors.New("mediastore upload driver requires the MEDIASTORE_DATA_ENDPOINT")
	}
	return cfg, nil
}

func (c *Config) uploader() (uploader.Uploader, error) {
	if c.UploadDriver == "s3" {
		return s3.New(c.BucketName)
	}
	if c.UploadDriver == "mediastore" {
		return mediastore.New(c.DataEndpoint)
	}
	return nil, fmt.Errorf("invalid upload driver %q", c.UploadDriver)
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

func (c *Config) addCacheMetadata(options *uploader.Options) {
	if value := c.CacheControl.HeaderValue(options.Path); value != nil {
		options.CacheControl = *value
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

	uper, err := cfg.uploader()
	if err != nil {
		logger.WithError(err).Fatal("failed to create uploader")
	}

	http.HandleFunc(cfg.HealthcheckPath, healthcheck)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		if r.Method != "POST" && r.Method != "PUT" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := strings.TrimLeft(r.URL.Path, "/")
		contentType := mime.TypeByExtension(filepath.Ext(key))
		logFields := logrus.Fields{"bucket": cfg.BucketName, "objectKey": key}
		options := uploader.Options{
			Path:        key,
			Body:        r.Body,
			ContentType: contentType,
		}
		cfg.addCacheMetadata(&options)
		err = uper.Upload(options)
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
