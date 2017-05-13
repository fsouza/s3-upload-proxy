package main

import (
	"fmt"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration of the s3-uploader.
type Config struct {
	BucketName      string `envconfig:"BUCKET_NAME" required:"true"`
	HealthcheckPath string `envconfig:"HEALTHCHECK_PATH" default:"/healthcheck"`
	HTTPPort        int    `envconfig:"HTTP_PORT" default:"80"`
	LogLevel        string `envconfig:"LOG_LEVEL" default:"debug"`
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
		logFields := logrus.Fields{"bucket": cfg.BucketName, "objectKey": key}
		logger.WithFields(logFields).Debug("uploading file to S3")
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(cfg.BucketName),
			Key:         aws.String(key),
			Body:        r.Body,
			ContentType: aws.String(mime.TypeByExtension(filepath.Ext(key))),
		})
		if err != nil {
			logger.WithFields(logFields).WithError(err).Error("failed to upload file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "OK")
	})

	listenAddr := fmt.Sprintf(":%d", cfg.HTTPPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to start listener")
	}
	defer listener.Close()
	http.Serve(listener, nil)
}
