package s3

import (
	"context"
	"os"
	"testing"
)

func TestUploadS3Text(t *testing.T) {
	conf := &Config{
		AccessKeyId:     "oh1ObY74FvJf6Zrx29pm",
		AccessKeySecret: "09X2QNXxb9228gz3jOpMvt4JyWmvqXpNGNOztrG3",
		Region:          "ap-east-1",
		Bucket:          "test",
		Endpoint:        "http://127.0.0.1:58742",
		PublicUrl:       "http://127.0.0.1:58742/test",
	}
	svc, err := NewService(conf)
	if err != nil {
		t.Fatalf("failed to create s3 service: %v", err)
	}
	data := []byte("hello s3")
	url, err := svc.UploadBytes(context.Background(), "test-folder/hello.txt", data, "text/plain")
	if err != nil {
		t.Fatalf("failed to upload bytes: %v", err)
	}
	t.Logf("uploaded file URL: %s", url)
}

func TestUploadS3Images(t *testing.T) {
	conf := &Config{
		AccessKeyId:     "oh1ObY74FvJf6Zrx29pm",
		AccessKeySecret: "09X2QNXxb9228gz3jOpMvt4JyWmvqXpNGNOztrG3",
		Region:          "ap-east-1",
		Bucket:          "test",
		Endpoint:        "http://127.0.0.1:58742",
		PublicUrl:       "http://127.0.0.1:58742/test",
	}
	svc, err := NewService(conf)
	if err != nil {
		t.Fatalf("failed to create s3 service: %v", err)
	}
	file, err := os.ReadFile("/Users/amos/Downloads/unnamed.png")
	if err != nil {
		return
	}
	url, err := svc.UploadBytes(context.Background(), "test-folder/unnamed.png", file, "image/png")
	if err != nil {
		t.Fatalf("failed to upload bytes: %v", err)
	}
	t.Logf("uploaded file URL: %s", url)
}
