package logic

import (
	"crypto/md5"
	"fmt"
)

/*
  genMd5Sum generates a md5 hashed representation of the given bytes, and
  returns the sum.
*/
func (lc Lgc) genMd5Sum(in []byte) []byte {
	var out []byte
	h := md5.New()
	h.Write(in)
	return h.Sum(out)
}

/*
  genSigMD generates a md5 hashed representation of the bytes, adding the
  pepper to secure the bytes, and rendering is useless to someone viewing
  the string.
*/
func (lc Lgc) genSigMD(in []byte) []byte {
	var sum []byte
	h := md5.New()
	// Apply a pepper to the input.
	idNow, _ := lc.addPepper(in)
	h.Write(idNow)
	return h.Sum(sum)
}

/*
  getObfID returns an encrypted ID used to obfuscate the owner
  of particular records. Used through the database, DO NOT UPDATE THIS METHOD
  AS IT WILL CAUSE CATASTROPHIC CONSEQUENCES.
*/
func (lc Lgc) getObfID(in string) string {
	// Encrypt the provided session ID as the session id is a representation
	// of the original id.
	idEnc := lc.genSigMD([]byte(in))
	sId := fmt.Sprintf("%x", idEnc)
	return sId
}
