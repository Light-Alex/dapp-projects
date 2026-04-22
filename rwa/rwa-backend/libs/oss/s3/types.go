package s3

type Config struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	// PublicUrl is the public URL prefix used to access uploaded objects.
	// Example: "https://cdn.example.com" — final object URL will be PublicUrl + "/" + objectKey
	PublicUrl string `json:"publicUrl"`
}
