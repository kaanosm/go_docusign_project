package setup

import (
	"net/http"
	"pleasesign/controller"
	"pleasesign/logic"
)

func PAppForward(d logic.DataCaller, lc logic.Lgc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/v1/documents":
			controller.PAppListDocuments(d, lc).ServeHTTP(w, r)
		case r.URL.Path == "/v1/createDocument":
			controller.PAppCreateDocument(d, lc).ServeHTTP(w, r)
		case r.URL.Path == "/v1/document/link":
			controller.PAppGetDocumentLink(d, lc).ServeHTTP(w, r)
		default:
			eI := &controller.ErrorNowInput{
				Writer:    w,
				ErrString: "Route does not exist - " + r.Method + " " + r.URL.Path,
				Code:      404,
			}
			controller.ErrorNow(eI)
			return
		}
	})
}
