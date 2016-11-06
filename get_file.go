package logic

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
)

type GetFileInput struct {
	Key    string // The filename of the file to get.
	Bucket string // The bucket containing the file.
	EncKey string // The encryption key.
}

// Retrieves a file from s3, and decrypts the file using the provided
// encryption key.
func (pvl PrivLogic) GetFile(in *GetFileInput) ([]byte, error) {
	client := s3.New(session.New(), &aws.Config{Region: aws.String("ap-southeast-2")})

	// Construct the input for the GetObject.
	params := &s3.GetObjectInput{
		Bucket: aws.String(in.Bucket),
		Key:    aws.String(in.Key),
	}

	// Get the file.
	output, err := client.GetObject(params)
	if err != nil {
		return nil, err
	}

	// Read the response into a slice and return.
	file, err := ioutil.ReadAll(output.Body)
	return file, err
}
