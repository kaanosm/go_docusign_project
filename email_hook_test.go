package logic

import (
	"database/sql"
	"testing"
)

// Test email ingest function for the mandrill email webhooks.
func TestEmailHook(t *testing.T) {

	// We need to test that the EmailHook function does not send emails
	// to recipients when the state is not relevant.
	// We can tell this happens when the GetMock function is not hit,
	// as this happens after the state checks.
	count := 0
	db := &MockDb{
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			// Any hits on this should do nothing.
			return nil, nil
		},
		GetMock: func(d interface{}, q string, args ...interface{}) error {
			count++
			return nil
		},
	}
	in := &EmailhookInput{State: "received"}
	lc := Lgc{}

	lc.Emailhook(db, in)

	if count > 0 {
		t.Error("Email hook is not ignoring non-bounce emails.")
	}

	// We can also check that emails that dont align to a correspondence
	// record (ie. they are system emails) are ignored. Only emails that
	// have bounced AND align to a recipient of a document should be actioned.

	db = &MockDb{
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			// Any hits on this should do nothing.
			return nil, nil
		},
		GetMock: func(d interface{}, q string, args ...interface{}) error {
			// If this is hit, increment the counter (it should
			// only be hit once and we'll check against that).
			count++
			return sql.ErrNoRows
		},
	}
	pvl := &MockPrivateLogic{
		buildNotificationEmailMock: func(in *buildNotificationInput) error {
			// This should not be hit, and will push the count to 3
			// if it is.
			count++
			return nil
		},
	}
	lc.Pvl = pvl
	in = &EmailhookInput{State: "soft_bounce"}
	lc.Emailhook(db, in)

	if count != 1 {
		t.Error("Email hook is not ignoring emails not aligning to a correspondence.")
	}

}
