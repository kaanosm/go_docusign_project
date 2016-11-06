# DocuSign/PleaseSign API

A small portion of my docusign/pleasesign API codebase,
most of the data structures and logic API is in the
private repository.





```
docker run --publish 8080:3000 -d \
 --env database_user= \
 --env database_url= \
 --env database_name= \
 --env database_password= \
 --env thumbnail_bucket= \
 --env master_bucket= \
 --env signature_bucket= \
 --env thumbnail_encryption= \
 --env master_encryption= \
 --env signature_encryption= \
 --env front_end= \
<IMAGE> 
```

### Example Config file.
```
package config

const (
	MandrillKey         = "123456"
	DatabaseUser        = "root"
	DatabaseURL         = "127.0.0.1:3306"
	DatabaseName        = "test"
	WorkDir             = ""
	CpdfDir             = "/cpdf"
	ThumbnailBucket     = ""
	ThumbnailEncryption = ""
	MasterEncryption    = ""
	MasterBucket        = ""
	SignatureEncryption = ""
	SignatureBucket     = ""
	LogPath             = ""
	FrontEnd            = ""
	TemplatePath        = ""

)
```
