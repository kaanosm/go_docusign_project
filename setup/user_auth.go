package setup

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/dchest/uniuri"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"os"
	"pleasesign/controller"
	"pleasesign/errlogger"
	"pleasesign/logic"
	"strconv"
	"strings"
	"time"
)

// retrieveUser pulls vital information for authenticating
// the user against the system. This information is passed
// to the AuthenticationHandler which analyses its input
// and determines if the request should be fulfilled based
// on authentication.
func retrieveUserAuth(db logic.DataCaller, r *http.Request) *logic.UserAuth {

	if r.Header.Get("X-PLEASESIGN-ID") == "" {
		return nil
	}

	type u struct {
		Id          string         `db:"id"`
		Enterprise  sql.NullString `db:"enterprise_id"`
		KeyID       string         `db:"key_id"`
		SecretKey   string         `db:"secret_key"`
		Expiry      string
		Active      bool
		Kind        int
		Verified    bool `db:"email_verified"`
		GroupActive bool `db:"group_active"`
	}

	var uq u

	q := "SELECT user.id, user.enterprise_id, `keys`.key_id, `keys`.secret_key, " +
		"`keys`.expiry, `keys`.active, user.kind, user.email_verified, user.group_active " +
		"FROM `keys` INNER JOIN user ON `keys`.user_id = user.id " +
		"WHERE `keys`.key_id = ?;"

	err := db.Get(&uq, q, r.Header.Get("X-PLEASESIGN-ID"))
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRUSERAUTH1", E: err})
		return nil
	}

	ua := logic.UserAuth{
		Id:          uq.Id,
		Enterprise:  uq.Enterprise.String,
		KeyID:       uq.KeyID,
		SecretKey:   uq.SecretKey,
		Expiry:      uq.Expiry,
		Active:      uq.Active,
		Kind:        uq.Kind,
		Verified:    uq.Verified,
		GroupActive: uq.GroupActive,
	}

	return &ua
}
func AuthenticateHandler(ua *logic.UserAuth, d logic.DataCaller, db logic.DataStore, lgc logic.Lgc) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// set header for global cors
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"X-Requested-With, X-Real-Ip, X-Forwarded-For, Content-Type, X-PLEASESIGN-KEY, X-PLEASESIGN-ID",
			)
			w.Header().Set("Access-Control-Allow-Methods", "PUT, POST, GET, DELETE")

			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Options accepted.",
				Code:      200,
			}
			controller.ErrorNow(eI)
			return
		}

		// if the URL is a public route, we want to go ahead and
		// skip this middleware as we do not require authentication.
		if r.URL.Path == "/key" ||
			r.URL.Path == "/user" ||
			r.URL.Path == "/" ||
			r.URL.Path == "/public/agree" ||
			r.URL.Path == "/status" ||
			r.URL.Path == "/public/session" ||
			r.URL.Path == "/public/signature" ||
			r.URL.Path == "/public/tabs" ||
			r.URL.Path == "/public/void" ||
			r.URL.Path == "/reminders" ||
			r.URL.Path == "/user/signature" ||
			r.URL.Path == "/password/send" ||
			r.URL.Path == "/password/new" ||
			r.URL.Path == "/resetQuotas" ||
			r.URL.Path == "/public/verification" ||
			r.URL.Path == "/email_hook" ||
			r.URL.Path == "/ecomm_hook" ||
			r.URL.Path == "/ecomm_schedule" ||
			r.URL.Path == "/document/callback" ||
			r.URL.Path == "/verify_resend" {
			Forward(d, db, lgc).ServeHTTP(w, r)
			return
		}

		// ensure that the correct headers are passed for authorisation
		if r.Header.Get("X-PLEASESIGN-ID") == "" || r.Header.Get("X-PLEASESIGN-KEY") == "" {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Please supply authorisation headers with your request.",
				Code:      401,
			}
			controller.ErrorNow(eI)
			return
		}

		if ua == nil {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "This key cannot be found",
				Code:      401,
			}
			controller.ErrorNow(eI)
			return
		}

		// If the key has expired, return error letting the user
		// know they need to log in again.
		expDate, _ := time.Parse("2006-01-02 15:04:05", ua.Expiry)
		if expDate.Before(time.Now()) {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Your token has expired, please log in again.",
				Code:      401,
			}
			controller.ErrorNow(eI)
			return
		}

		challenge := []byte(r.Header.Get("X-PLEASESIGN-KEY"))
		decryptionError := bcrypt.CompareHashAndPassword([]byte(ua.SecretKey), challenge)

		if decryptionError != nil {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Invalid security token.",
				Code:      401,
			}
			controller.ErrorNow(eI)
			return
		}
		lgc.CurrentUser = ua

		// we want to return a user object so that our routes
		// can identify who is running them.
		Forward(d, db, lgc).ServeHTTP(w, r)
	}

	return http.HandlerFunc(h)
}

// we create a struct that will fufil responseWriter
type LogRecord struct {
	http.ResponseWriter
	status int
}

func (r *LogRecord) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type myReader struct {
	*bytes.Buffer
}

func (m myReader) Close() error { return nil }

// createRequest encapsulates the incoming request for audit
// purposes. Each request is logged to disk which is then rotated
// to the cloud for processing.
func CreateRequest(d logic.DataCaller, db logic.DataStore, lgc logic.Lgc, runID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// we want to time how long the request takes, so
		// lets start the timer.
		start := time.Now()

		// the incoming body of the request is in a io.ReadCloser.
		// we cant to read it, however we need the copy it into
		// two buffers before doing so ( if we don't do this
		// we will run the buffer dry and the controller will
		// not be able to access it.
		buf, _ := ioutil.ReadAll(r.Body)
		rdr2 := myReader{bytes.NewBuffer(buf)}

		// we will now set the body to one of the buffers
		// so the request can proceed as normal after this op
		r.Body = rdr2

		// Get the IP address from the X-Forwarded-For header.
		getIp := func() string {
			split := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
			ip := split[len(split)-1]
			out := strings.TrimSpace(ip)
			return out
		}

		rID := uniuri.NewLen(20)

		// we want to construct something that satisfies the HTTP
		// readWriter so that we can hijack it and store the status code
		// of what our controller returned.
		record := &LogRecord{ResponseWriter: w}

		// We will go on to call our User or Enterprise Authenticator as
		// per normal, which will in turn call the controller via the router.
		// We will pass in our HTTP writer as well so that we can listen
		// in on the status code.
		var ua *logic.UserAuth
		var ent *logic.PAppAuth

		if r.Header.Get("X-PLEASESIGN-APP") == "" {
			// Lookup the user authentication and handle the
			// http request.
			ua = retrieveUserAuth(d, r)
			hand := AuthenticateHandler(ua, d, db, lgc)
			hand.ServeHTTP(record, r)
		} else {
			// Lookup the enterprise authentication and handle
			// the http request.
			ent = retrievePAppAuth(d, r)
			hand := PAppAuthHandler(ent, d, lgc)
			hand.ServeHTTP(record, r)
		}

		// we class the call as complete, and we want to stop the timer.
		duration := time.Since(start)

		// we need to get a stringified version of the yyyy/mm/dd/hh/mm
		y, m, d := start.Date()

		year := strconv.Itoa(y)
		month := strconv.Itoa(int(m))
		day := strconv.Itoa(d)

		h, min, _ := start.Clock()
		hour := strconv.Itoa(h)
		minute := strconv.Itoa(min)

		type logBody struct {
			RequestID     string
			URI           string
			UserID        string
			ContentLength int64
			AuthID        string
			Time          string
			Method        string
			Status        int
			Agent         string
			Ip            string
			Duration      string
			Body          string
		}

		// lets create the log filename
		// we want it to be unique to the minute so the logs will collate
		// for a whole minute. We will also append the runID which denotes
		// a unique id of this instance of the API. This prevents clashes
		// if this same API is running on multiple servers, and pushing logs.
		fname := year + month + day + "_" + hour + minute + "_" + runID

		var uID string
		if ua != nil {
			uID = ua.Id
		}

		var aID string
		if r.Header.Get("X-PLEASESIGN-KEY") == "" {
			aID = r.Header.Get("X-PLEASESIGN-APP")
		} else {
			aID = r.Header.Get("X-PLEASESIGN-KEY")
		}

		// lets setup the body of the log in preperation
		// for JSON marshalling.
		logB := logBody{
			RequestID:     rID,
			URI:           r.RequestURI,
			UserID:        uID,
			AuthID:        aID,
			ContentLength: r.ContentLength,
			Time:          time.Now().String(),
			Method:        r.Method,
			Status:        record.status,
			Agent:         r.UserAgent(),
			Ip:            getIp(),
			Duration:      duration.String(),
		}

		jbod, _ := json.Marshal(logB)

		file, _ := os.OpenFile(
			"/tmp/plsapi_"+fname,
			os.O_RDWR|os.O_APPEND|os.O_CREATE,
			0666,
		)

		deli := []byte("***")

		file.Write(jbod)
		file.Write(deli)
	})
}
