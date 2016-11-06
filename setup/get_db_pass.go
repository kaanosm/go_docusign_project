package setup

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"pleasesign/config"
	e "pleasesign/errlogger"
)

// Decrypts the database password and returns for connection.
func DecPass(context string) string {
	if context != "live" {
		return config.DatabasePassword()
	}
	var out string
	penc, err := base64.StdEncoding.DecodeString(config.DatabasePassword())

	svc := kms.New(session.New(), &aws.Config{Region: aws.String("ap-southeast-2")})
	params := &kms.DecryptInput{
		CiphertextBlob: penc,
		EncryptionContext: map[string]*string{
			"Key": aws.String(config.MasterEncryption()),
		},
		GrantTokens: []*string{aws.String("GrantTokenType")},
	}
	resp, err := svc.Decrypt(params)
	if err != nil {
		e.ThrowError(&e.LogInput{
			M: "Unable to decrypt the pepper.",
			E: err,
		})
		return out
	}
	out = string(resp.Plaintext)

	return out
}
