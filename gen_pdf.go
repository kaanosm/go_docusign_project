package logic

import (
	"database/sql"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/jung-kurt/gofpdf"
	"pleasesign/config"
	e "pleasesign/errlogger"
	"time"
)

/*
  createCertAuth is used to perform security checks as a step prior to generating
  the certificate of authenticity for a document. If any errors occur in the
  generation of the document, these will be returned as-is from the function.
*/
func (Lgc) createCertAuth(documentID string, db DataCaller, lc Lgc) error {
	// Check the document events have not been tampered since the document
	// was created, and error if so.
	if err := lc.Pvl.checkEventSecurity(documentID, db, lc); err != nil {
		// If the event check failed, email developers to inform.
		emIn := &buildSecurityFailInput{
			Db:         db,
			DocumentID: documentID,
		}
		lc.Pvl.buildSecurityFailEmail(emIn)
	}

	err := lc.genpdf(documentID, db, lc)
	return err
}

/*
  Genpdf creates the certificate of authenticity for a document, either in the
  complete or void stage.
*/
func (Lgc) genpdf(documentID string, db DataCaller, lc Lgc) error {
	// Instantiate the pdf maker to be used.
	pdf := gofpdf.New("P", "pt", "A4", ".")

	// Add a new page straight away.
	pdf.AddPage()

	////PAGE 1
	// Draw the grey background at the top.
	// Add the logo to the top left.
	lp := config.LogoPath()
	pdf.Image(lp, 15, 19, -250, -250, false, "", 0, "")
	// Create the Certificate of Authenticity and v1 text along the top.
	pdf.SetFont("Helvetica", "B", 23)
	pdf.WriteAligned(0, 39, "                                 Certificate of Authenticity", "C")
	pdf.Ln(25)
	pdf.SetFont("Helvetica", "", 20)
	pdf.WriteAligned(0, 39, "                                    v1", "C")

	// Move down to beneath the top banner and reset the font.
	pdf.SetY(122)
	pdf.SetFont("Helvetica", "", 18)
	pdf.SetTextColor(0, 0, 0)

	// Fill in the document details.
	currX := pdf.GetX()
	pdf.SetX(currX + 15)
	pdf.WriteAligned(0, 10, "Document Details", "L")
	pdf.Ln(20)

	// Get the document information from the db.
	doc, err := lc.getDocumentDetailCert(documentID, db)
	if err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})
	}

	// Construct the formatted string from the document details retrieved.
	// Ensure the status of the document is checked as this will change
	// the text that will be written (and that was returned by the db call above.)
	var txtStr string
	if doc.status != "void" {
		switch {
		case doc.docType == "" && doc.message == "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nStatus: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.status, doc.pages, doc.date)
		case doc.docType == "" && doc.message != "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nMessage to recipients: %v\nStatus: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.message, doc.status, doc.pages, doc.date)
		case doc.docType != "" && doc.message == "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nStatus: %v\nType: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.status, doc.docType, doc.pages, doc.date)
		case doc.docType != "" && doc.message != "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nMessage to recipients: %v\nStatus: %v\nType: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.message, doc.status, doc.docType, doc.pages, doc.date)
		}
	} else {
		switch {
		case doc.docType == "" && doc.message == "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nStatus: %v\nVoid Reason: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.status, doc.voidReason, doc.pages, doc.date)
		case doc.docType == "" && doc.message != "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nMessage to recipients: %v\nStatus: %v\nVoid Reason: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.message, doc.status, doc.voidReason, doc.pages, doc.date)
		case doc.docType != "" && doc.message == "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nStatus: %v\nVoid Reason: %v\nType: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.status, doc.voidReason, doc.docType, doc.pages, doc.date)
		case doc.docType != "" && doc.message != "":
			txtStr = fmt.Sprintf("ID: %v\nTitle: %v\nMessage to recipients: %v\nStatus: %v\nVoid Reason: %v\nType: %v\nPages: %v\nCreated: %v",
				doc.id, doc.title, doc.message, doc.status, doc.voidReason, doc.docType, doc.pages, doc.date)
		}
	}
	// Write the string that was formatted above.
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetX(currX + 20)
	pdf.MultiCell(0, 17, string(txtStr), "", "", false)

	// Draw the line to seperate document details from recipient details.
	pdf.Ln(20)
	currY := pdf.GetY()
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(0, currY, 595, currY)
	pdf.Ln(20)

	// Fill in the recipient details heading.
	pdf.SetFont("Helvetica", "", 18)
	pdf.SetX(currX + 15)
	pdf.WriteAligned(0, 10, "Recipient Details", "L")
	pdf.Ln(20)

	// Retrieve the recipients for the document, and loop through them
	// to write to the page. Within the next section there is a lot
	// of string formatting that handle what has been returned from the
	// database - due to the certificate being for either a completed
	// document or a voided document.
	// The thumbnail of the signature is also drawn on the page during
	// this process.
	recipients, err := lc.getRecipientDetailCert(documentID, db)
	if err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})
	}

	for i, r := range recipients {
		pdf.SetFont("Helvetica", "", 12)
		pdf.SetX(currX + 20)
		pdf.WriteAligned(0, 20, r.name, "L")
		pdf.SetFont("Helvetica", "", 10)
		// If there was no session the below information
		// is not relevant and will all be blank.
		if r.sessionid != "" {
			// Check that the thumbnail exists before drawing the thumbnail.
			if r.thumb != "" {
				// Print the signature on the page at the current position if
				// the recipient has a valid session.
				pdf.SetX(325)
				currX, currY := pdf.GetXY()
				pdf.Image(r.thumb, currX+21, currY+12, 140, 0, false, "", 0, "")
			}
			pdf.Ln(20)
			currX = pdf.GetX()
			pdf.SetX(currX + 20)
			var newTxtStr string
			if r.thumb == "" {
				if r.geolat == 0 || r.geolong == 0 {
					newTxtStr = fmt.Sprintf("ID: %v(%v)\nAuthentication: %v\nSession ID: %v\nIP address: %v\nBrowser: %v\n", r.id, r.email, r.security, r.sessionid, r.ip, r.useragent)
				} else {
					newTxtStr = fmt.Sprintf("ID: %v(%v)\nAuthentication: %v\nSession ID: %v\nIP address: %v\nLocation signed: %v, %v\nBrowser: %v",
						r.id, r.email, r.security, r.sessionid, r.ip, r.geolong, r.geolat, r.useragent)
				}
			} else {
				if r.geolat == 0 || r.geolong == 0 {
					newTxtStr = fmt.Sprintf("ID: %v(%v)\nAuthentication: %v\nDate signed: %v UTC\nSession ID: %v\nIP address: %v\nBrowser: %v\n", r.id, r.email, r.security, r.date, r.sessionid, r.ip, r.useragent)
				} else {
					newTxtStr = fmt.Sprintf("ID: %v(%v)\nAuthentication: %v\nDate signed: %v UTC\nSession ID: %v\nIP address: %v\nLocation signed: %v, %v\nBrowser: %v",
						r.id, r.email, r.security, r.date, r.sessionid, r.ip, r.geolong, r.geolat, r.useragent)
				}
			}
			pdf.MultiCell(0, 17, string(newTxtStr), "", "", false)

		} else {
			// This is the case when the recipient did not have a
			// valid session, ie the document has been voided and
			// it simply needs to list the potential recipient
			// and their email. The recipient's name has been printed
			// already so just their email is formatted and written
			// to the certificate.
			pdf.Ln(20)
			currX = pdf.GetX()
			pdf.SetX(currX + 20)
			newTxtStr := fmt.Sprintf("%v", r.email)
			pdf.MultiCell(0, 17, string(newTxtStr), "", "", false)
		}

		// If we are not at the last of the recipients, move down 30 to create room to
		// print the next recipient.
		if i+1 != len(recipients) {
			pdf.Ln(30)
		}
	}

	// Draw the line to seperate recipient details from event details.
	currY = pdf.GetY()
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(0, currY+20, 595, currY+20)
	pdf.Ln(40)

	// Fill in the event details
	pdf.SetFont("Helvetica", "", 16)
	pdf.SetX(currX + 15)
	pdf.WriteAligned(0, 10, "Document Events", "L")
	pdf.Ln(20)

	pdf.SetFont("Helvetica", "", 10)
	events, err := lc.getEventDetailCert(documentID, db)
	if err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})
	}
	for _, event := range events {
		pdf.SetX(currX + 20)
		newTxtStr := fmt.Sprintf("%v - %v", event.Date, event.Body)
		pdf.MultiCell(500, 12, string(newTxtStr), "", "", false)
		pdf.Ln(5)
	}

	pdf.Ln(20)
	pdf.SetFont("Helvetica", "", 8)
	pdf.WriteAligned(0, 10, "All times on this document are in UTC.", "L")
	pdf.Ln(10)
	newTxtStr := fmt.Sprintf("Certificate generated on %v.", time.Now().UTC().Format("2006-01-02 15:04:05"))
	pdf.MultiCell(500, 12, string(newTxtStr), "", "", false)

	///// Output the pdf to disk.
	directory := config.WorkDir() + uniuri.New() + ".pdf"

	if err = pdf.OutputFileAndClose(directory); err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})

	}

	///// Upload the certificate to s3.
	err = lc.uploadDocumentCertificate(directory, documentID, db)
	if err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})
	}

	///// Delete all of the temporary files.
	// Del the actual certificate, and then loop the recipients and delete
	// the signatures if they exist.
	if err = lc.Pvl.SecureDelete(directory); err != nil {
		return e.ThrowError(&e.LogInput{
			M: err.Error() + " " + documentID,
		})
	}
	for _, r := range recipients {
		if r.thumb != "" {
			err = lc.Pvl.SecureDelete(r.thumb)
			if err != nil {
				return e.ThrowError(&e.LogInput{
					M: err.Error() + " " + documentID,
				})
			}
		}
	}

	return nil
}

/*
  uploadDocumentCertificate will take a filepath, and upload the file to s3.
  It also updates the document_keys record to include the certificate key.
*/
func (lc Lgc) uploadDocumentCertificate(filepath string, documentID string, db DataCaller) error {
	// Get the filebytes for passing through to the uploading.
	b, err := lc.Pvl.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Generate a key, and upload the file to s3. Afterwards, update
	// the document_keys record with the certificate key.
	key := uniuri.New() + ".pdf"
	bk := config.MasterBucket()
	enc := config.MasterEncryption()
	err = lc.Pvl.StoreFile(key, b, bk, enc)
	if err != nil {
		return err
	}

	q := `UPDATE document_keys SET certificate_key = ? WHERE document_id = ?;`
	_, err = db.Exec(q, key, documentID)
	return err
}

type documentDetail struct {
	id         string
	title      string
	message    string
	status     string
	docType    string
	pages      int
	date       string
	voidReason string
}

// Setup the destination struct, and query the db for the required information.
type d struct {
	Id         string
	Title      string
	Status     string
	Pages      int
	Date       string `db:"created"`
	Message    sql.NullString
	DocType    sql.NullString `db:"kind"`
	VoidReason sql.NullString `db:"void_reason"`
}

/*
  getDocumentDetail is used to retrieve and format the document information
  as expected by the report.
*/
func (lc Lgc) getDocumentDetailCert(documentID string, db DataCaller) (documentDetail, error) {
	var out documentDetail
	var doc d

	// Get the document information from the db.
	q := `SELECT id, title, message, status, kind, pages, created, void_reason
          FROM documents WHERE id = ?;`
	err := db.Get(&doc, q, documentID)
	if err != nil {
		return out, err
	}

	// Map the result to the doc details.
	out.id = doc.Id
	out.title = doc.Title
	out.status = doc.Status
	out.pages = doc.Pages
	out.date = doc.Date

	if doc.Message.String != "" {
		out.message = doc.Message.String
	}
	if doc.DocType.String != "" {
		out.docType = doc.DocType.String
	}
	if doc.VoidReason.String != "" {
		out.voidReason = doc.VoidReason.String
	}

	return out, nil
}

type eventDetail struct {
	Date string
	Body string
}

/*
  getEventDetailCert pulls all of the events for the document and formats
  them as expected by the report.
*/
func (lc Lgc) getEventDetailCert(documentID string, db DataCaller) ([]eventDetail, error) {
	// Retrieve the event details for the document from the db, scanned into
	// the events.
	var events []eventDetail

	q := `SELECT created AS 'date', body FROM events WHERE document_id = ? 
          ORDER BY created ASC;`

	err := db.Select(&events, q, documentID)
	if err != nil {
		return events, err
	}

	return events, nil
}

type recipientDetail struct {
	id        string
	name      string
	email     string
	security  string
	date      string
	sessionid string
	ip        string
	thumb     string
	geolat    float64
	geolong   float64
	useragent string
}

/*
  getRecipientDetailCert is used to format the expected recipient information
  for the report.
*/
func (lc Lgc) getRecipientDetailCert(documentID string, db DataCaller) ([]recipientDetail, error) {
	var out []recipientDetail

	// We need to retrieve all of the active recipients from the database
	// for the documentID as the information is used in the report.
	var recipients []Recipient
	q := `SELECT id, first_name, last_name, email, complete FROM recipients 
          WHERE document_id = ? AND active = 1;`
	err := db.Select(&recipients, q, documentID)
	if err != nil {
		return out, err
	}

	/*
	   After getting the recipients, the session that the recipient agreed
	   and signed needs to be retrieved, and added to the response
	   from the previous call.
	*/
	for _, recipient := range recipients {
		n := fmt.Sprintf("%v %v", recipient.First_name, recipient.Last_name)
		r := recipientDetail{
			id:    recipient.Id,
			name:  n,
			email: recipient.Email,
		}
		if recipient.Complete.String != "" {
			r.date = recipient.Complete.String
		}

		// Get the session, and map the returned values to the recipient's output.
		// As there is a possibility the recipient has not got an agreed
		// session (for a voided doc) we need to check all values.
		recSession, err := lc.getAgreedSessionCert(recipient.Id, db)
		if err != nil {
			return out, err
		}
		if recSession.Id != "" {
			r.sessionid = recSession.Id
		}
		if recSession.Ip_address != "" {
			r.ip = recSession.Ip_address
		}
		if recSession.Geo_lat.Valid != false {
			r.geolat = recSession.Geo_lat.Float64
		}
		if recSession.Geo_long.Valid != false {
			r.geolong = recSession.Geo_long.Float64
		}
		if recSession.User_agent != "" {
			r.useragent = recSession.User_agent
		}
		if recSession.Security != "" {
			r.security = recSession.Security
		}

		/*
		   Now we need to download and store the signature to be
		   stamped on the document. This should only be done if
		   the recipient has signed (and has a sessionid).
		*/
		if r.sessionid != "" {
			fp, err := lc.getGuestSignatureCert(r.sessionid, db)
			if err != nil {
				return out, e.ThrowError(&e.LogInput{
					M: "Error when generating document " + documentID,
					E: err,
				})
			}
			if fp != "" {
				r.thumb = fp
			}
		}
		out = append(out, r)
	}

	return out, err
}

/*
  getGuestSignatureCert retrieves the signature from the session and writes
  the file to disk.
*/
func (lc Lgc) getGuestSignatureCert(sessionID string, db DataCaller) (string, error) {
	var out string

	// Create a function to return the filename of a successful retrieval
	// of a signature.
	getFile := func(key string) (string, error) {
		// Retrieve the bytes from s3.
		input := &GetFileInput{
			Key:    key,
			Bucket: config.SignatureBucket(),
		}
		fileBytes, err := lc.Pvl.GetFile(input)
		if err != nil {
			return out, err
		}

		// Write the bytes to disk for use in the report and return
		// the filename to the thumbnail.
		fileName := config.WorkDir() + key
		err = lc.Pvl.WriteFile(fileName, fileBytes, 1)
		if err != nil {
			return out, err
		}

		return fileName, err
	}

	// Retrieve the signature used for the session.
	sigID := lc.sessionSignatureGet(db, sessionID)
	q := `SELECT bucket_key FROM signatures WHERE id = ?;`
	var key string
	err := db.Get(&key, q, sigID)
	if err != nil {
		return out, nil
	}
	// Get the file using the function above.
	out, err = getFile(key)
	return out, err
}

/*
  getAgreedSession retrieves the session for a recipient in which they agreed
  and signed their tabs. This is used to display the ip address and geolat
  and geolong values on the report.
*/
func (lc Lgc) getAgreedSessionCert(recipientID string, db DataCaller) (Session, error) {

	// Retrieve the agreed session for a recipient.
	var out Session
	q := `SELECT sessions.created, sessions.user_agent, 
              sessions.ip_address, sessions.geo_lat, sessions.geo_long,
              sessions.id, security FROM sessions 
              WHERE sessions.recipient_id = ? AND sessions.agreed = 1
              LIMIT 1;`
	err := db.Get(&out, q, recipientID)

	// If the error is an sql ErrNoRows, it should not return an error as
	// this is acceptable in the cert generation process.
	if (err != nil) && (err != sql.ErrNoRows) {
		return out, err
	}

	return out, nil
}
