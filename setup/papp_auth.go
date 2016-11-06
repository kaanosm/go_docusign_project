package setup

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"pleasesign/controller"
	"pleasesign/logic"
)

/*
  retrievePAppAuth returns information for the authentication that
  will be used throughout the request.
*/
func retrievePAppAuth(db logic.DataCaller, r *http.Request) *logic.PAppAuth {
	type e struct {
		KeyID        string         `db:"id"`
		SecretKey    string         `db:"secret_key"`
		UserID       string         `db:"user_id"`
		EnterpriseID sql.NullString `db:"enterprise_id"`
	}
	var ent e

	q := `SELECT app_keys.id, app_keys.secret_key, app_keys.enterprise_id,
          app_keys.user_id FROM app_keys 
          WHERE app_keys.id = ? AND app_keys.active = 1;`

	if err := db.Get(&ent, q, r.Header.Get("X-PLEASESIGN-APP")); err != nil {
		return &logic.PAppAuth{}
	}

	out := &logic.PAppAuth{
		Id:           ent.KeyID,
		SecretKey:    ent.SecretKey,
		EnterpriseID: ent.EnterpriseID.String,
		UserID:       ent.UserID,
	}

	return out
}

/*
  PAppAuthHandler will take the authentication token, and
  ensure they are authenticated to continue.
*/
func PAppAuthHandler(pa *logic.PAppAuth, d logic.DataCaller, lgc logic.Lgc) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		// Return error if the Auth object has a blank id
		// (when the couldn't be found).
		if pa.Id == "" {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "This key cannot be found",
				Code:      400,
			}
			controller.ErrorNow(eI)
			return
		}

		sKey := r.Header.Get("X-PLEASESIGN-APPSECRET")
		err := bcrypt.CompareHashAndPassword([]byte(pa.SecretKey), []byte(sKey))

		if err != nil {
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Secret Key does not match",
				Code:      400,
			}
			controller.ErrorNow(eI)
			return
		}

		lgc.CurrentUser = &logic.UserAuth{
			Id:         pa.UserID,
			Enterprise: pa.EnterpriseID,
			KeyID:      pa.Id,
			SecretKey:  pa.SecretKey,
		}
		PAppForward(d, lgc).ServeHTTP(w, r)
	}

	return http.HandlerFunc(h)
}
