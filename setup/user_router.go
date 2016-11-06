package setup

import (
	"net/http"
	"pleasesign/controller"
	"pleasesign/logic"
)

func Forward(d logic.DataCaller, db logic.DataStore, logicController logic.Lgc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		// set header for global cors
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, X-Real-Ip, X-Forwarded-For, Content-Type")

		switch {

		case r.URL.Path == "/key":
			controller.PostKey(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/documentRec" && r.Method == "GET":
			controller.ListDocumentsRec(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/document" && r.Method == "GET":
			controller.ListDocuments(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user" && r.Method == "POST":
			controller.CreateUser(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/update" && r.Method == "PUT":
			controller.UserPut(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/userInfo":
			controller.GetUserInfo(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/document" && r.Method == "POST":
			controller.CreateDocument(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/status" && r.Method == "GET":
			controller.GetStatus().ServeHTTP(w, r)
		case r.URL.Path == "/public/session" && r.Method == "POST":
			controller.PostSession(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/public/agree" && r.Method == "POST":
			controller.PostAgree(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/public/tabs" && r.Method == "POST":
			controller.PostComplete(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/public/signature" && r.Method == "POST":
			controller.PostSignature(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/public/void" && r.Method == "POST":
			controller.VoidDocumentPublic(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/public/verification" && r.Method == "POST":
			controller.VerifyEmail(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/void" && r.Method == "POST":
			controller.VoidDocument(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/detail" && r.Method == "GET":
			controller.GetDetail(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/resendEmail" && r.Method == "POST":
			controller.PostResendEmail(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/correctRecipient" && r.Method == "PUT":
			controller.PutRecipient(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/documentSigned" && r.Method == "GET":
			controller.GetDocumentSigned(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/documentCertificate" && r.Method == "GET":
			controller.GetDocumentCertificate(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/documentOriginal" && r.Method == "GET":
			controller.GetDocumentOriginal(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/recipient" && r.Method == "POST":
			controller.PostRecipient(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/recipient" && r.Method == "DELETE":
			controller.DeleteRecipient(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/updateDocumentDetail" && r.Method == "PUT":
			controller.PutDocumentDetail(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/pages" && r.Method == "POST":
			controller.GetPages(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/tabs" && r.Method == "POST":
			controller.PostTabs(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/reminders" && r.Method == "GET":
			controller.DocumentReminders(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/password/send" && r.Method == "POST":
			controller.SendResetPassword(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/password/new" && r.Method == "PUT":
			controller.ResetPassword(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/resetQuotas" && r.Method == "GET":
			controller.ResetQuota(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/password/change" && r.Method == "PUT":
			controller.ChangePassword(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/papp/gen_token" && r.Method == "POST":
			controller.PAppPostKey(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/email_hook":
			controller.Emailhook(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/newSub" && r.Method == "POST":
			controller.EcommNewCustomer(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/plan" && r.Method == "GET":
			controller.EcommGetCustomerSub(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/updatePlan" && r.Method == "PUT":
			controller.EcommUpdateCustomerSub(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/updatePlanInfo" && r.Method == "PUT":
			controller.EcommUpdateCustomerInfo(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/invoices" && r.Method == "GET":
			controller.EcommGetCustomerInvoices(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/signature_link" && r.Method == "GET":
			controller.UserSignatureGet(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/ecomm_hook" && r.Method == "POST":
			controller.EcommWebHook(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/ecomm_schedule" && r.Method == "GET":
			controller.EcommSchedule(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/recipient/sign_link" && r.Method == "GET":
			controller.GenSigningLink(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/branding" && r.Method == "GET":
			controller.UserBrandingGet(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/branding" && r.Method == "PUT":
			controller.UserBrandingPut(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/user/logo" && r.Method == "POST":
			controller.UserBrandingLogoPost(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/document/callback" && r.Method == "POST":
			controller.DocumentCallbackPost(d, logicController).ServeHTTP(w, r)
		case r.URL.Path == "/verify_resend" && r.Method == "GET":
			controller.VerifyDocResend(d, logicController).ServeHTTP(w, r)
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
