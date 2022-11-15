// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsouza/s3-upload-proxy/internal/cachecontrol"
	"github.com/fsouza/s3-upload-proxy/internal/uploader"
	"github.com/fsouza/s3-upload-proxy/internal/uploader/mediastore"
	"github.com/fsouza/s3-upload-proxy/internal/uploader/s3"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/exp/slog"
)

// Config is the configuration of the s3-uploader.
type Config struct {
	BucketName      string             `envconfig:"BUCKET_NAME" required:"true"`
	UploadDriver    string             `envconfig:"UPLOAD_DRIVER" default:"s3"`
	HealthcheckPath string             `envconfig:"HEALTHCHECK_PATH" default:"/healthcheck"`
	HTTPPort        int                `envconfig:"HTTP_PORT" default:"80"`
	LogLevel        string             `envconfig:"LOG_LEVEL" default:"debug"`
	CacheControl    cachecontrol.Rules `envconfig:"CACHE_CONTROL_RULES"`
	ChunkedTransfer bool               `envconfig:"MEDIASTORE_CHUNKED_TRANSFER"`
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
	if cfg.ChunkedTransfer && cfg.UploadDriver != "mediastore" {
		return cfg, errors.New("MEDIASTORE_CHUNKED_TRANSFERS should only be defined for the mediastore UPLOAD_DRIVER")
	}
	return cfg, nil
}

func (c *Config) uploader() (uploader.Uploader, error) {
	if c.UploadDriver == "s3" {
		return s3.New()
	}
	if c.UploadDriver == "mediastore" {
		return mediastore.New(mediastore.DriverOptions{ChunkedTransfer: c.ChunkedTransfer})
	}
	return nil, fmt.Errorf("invalid upload driver %q", c.UploadDriver)
}

func (c *Config) logger() *slog.Logger {
	levels := map[string]slog.Level{
		"debug":   slog.DebugLevel,
		"info":    slog.InfoLevel,
		"warning": slog.WarnLevel,
		"warn":    slog.WarnLevel,
		"error":   slog.ErrorLevel,
	}
	opts := slog.HandlerOptions{Level: levels[c.LogLevel]}
	return slog.New(opts.NewTextHandler(os.Stderr))
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
		log.Fatal(err)
	}
	logger := cfg.logger()

	uper, err := cfg.uploader()
	if err != nil {
		logger.Error("failed to create uploader", err)
		os.Exit(1)
	}

	http.HandleFunc(cfg.HealthcheckPath, healthcheck)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := strings.TrimLeft(r.URL.Path, "/")
		contentType := mime.TypeByExtension(filepath.Ext(key))
		logger := logger.With(
			slog.String("bucket", cfg.BucketName),
			slog.String("objectkey", key),
			slog.String("contentType", contentType),
		)
		options := uploader.Options{
			Bucket:       cfg.BucketName,
			Path:         key,
			Body:         r.Body,
			ContentType:  stringPtr(contentType),
			Context:      r.Context(),
			CacheControl: cfg.CacheControl.HeaderValue(key),
		}
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			err = uper.Upload(options)
			if err != nil {
				logger.Error("failed to upload file", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logger.Debug(fmt.Sprintf("finished upload in %s", time.Since(start)))
		case http.MethodDelete:
			err = uper.Delete(options)
			if err != nil {
				logger.Error("failed to delete file", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logger.Debug(fmt.Sprintf("deleted in %s", time.Since(start)))
		}
		fmt.Fprintln(w, "OK")
	})

	listenAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Error("failed to start listener", err)
		os.Exit(1)
	}
	defer listener.Close()
	logger.Info(fmt.Sprintf("listening on %s", listener.Addr()))
	http.Serve(listener, nil)
}

// stringPtr makes empty strings a nil pointer.
func stringPtr(input string) *string {
	if input == "" {
		return nil
	}
	return &input
}
