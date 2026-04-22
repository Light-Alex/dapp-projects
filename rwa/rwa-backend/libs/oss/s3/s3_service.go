package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type Service struct {
	conf   *Config
	client *s3.Client
}

func NewService(conf *Config) (*Service, error) {
	s := &Service{conf: conf}
	client, err := s.newS3Client(context.Background())
	if err != nil {
		log.ErrorZ(context.Background(), "create s3 client failed", zap.Error(err))
		return nil, err
	}
	s.client = client
	return s, nil
}

func (s *Service) newS3Client(ctx context.Context) (*s3.Client, error) {
	if s == nil || s.conf == nil {
		return nil, fmt.Errorf("s3 service or config is nil")
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(s.conf.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.conf.AccessKeyId, s.conf.AccessKeySecret, "")),
	}
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: s.conf.Endpoint, SigningRegion: s.conf.Region, HostnameImmutable: true}, nil
	})
	opts = append(opts, awsconfig.WithEndpointResolverWithOptions(resolver))

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return client, nil
}

// Build a normalized public URL by joining PublicUrl and objectKey with proper escaping of path segments.
func (s *Service) buildPublicURL(objectKey string) string {
	if s == nil || s.conf == nil || s.conf.PublicUrl == "" {
		return ""
	}
	base := strings.TrimRight(s.conf.PublicUrl, "/")
	key := strings.TrimLeft(objectKey, "/")
	// escape each path segment to preserve slashes
	segments := strings.Split(key, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return base + "/" + strings.Join(segments, "/")
}

// Upload uploads data from an io.Reader to objectKey and returns a publicly-accessible URL.
// size can be zero (not required). contentType may be empty.
func (s *Service) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	// prefer cached client
	client := s.client
	var err error
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.conf.Bucket),
		Key:    aws.String(objectKey),
		Body:   reader,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	// ContentLength is optional; only set when > 0
	if size > 0 {
		input.ContentLength = size
	}

	_, err = client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	return s.buildPublicURL(objectKey), nil
}

// UploadBytes is a convenience wrapper that uploads the provided bytes and returns the URL.
func (s *Service) UploadBytes(ctx context.Context, objectKey string, data []byte, contentType string) (string, error) {
	return s.Upload(ctx, objectKey, bytes.NewReader(data), int64(len(data)), contentType)
}

// UploadFile uploads a local file (by path) to the given objectKey and returns the public URL.
func (s *Service) UploadFile(ctx context.Context, objectKey string, filePath string, contentType string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	return s.Upload(ctx, objectKey, f, fi.Size(), contentType)
}
