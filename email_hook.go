package logic

import (
	"database/sql"
	"fmt"
	e "pleasesign/errlogger"
	"time"
)

type EmailLocation struct {
	Country   string
	City      string
	Latitude  float64
	Longitude float64
}

type SmtpEventInput struct {
	DestinationIP string
	Diag          string
	Type          string
	TS            time.Time
}

type ClickBody struct {
	TS   time.Time
	Kind string
}

type EmailhookInput struct {
	ID                string
	Email             string
	State             string
	TS                time.Time
	UserAgent         string
	BounceDescription string
	Diag              string
	SMTPEvents        []SmtpEventInput
	Clicks            []ClickBody
	Location          *EmailLocation
}

/*
  Emailhook ingests information sent from the Mandrill email service and stores
  it in the database, ensuring it is linked to the correspondence based on the
  id provided by the third party.
*/
func (lc Lgc) Emailhook(db DataCaller, in *EmailhookInput) {
	ua := sql.NullString{String: in.UserAgent, Valid: in.UserAgent != ""}
	bd := sql.NullString{
		String: in.BounceDescription,
		Valid:  in.BounceDescription != "",
	}
	di := sql.NullString{
		String: in.Diag,
		Valid:  in.Diag != "",
	}

	// Store the main email event information.
	q := `INSERT INTO email_events (event_id, email, state, time, user_agent, bounce_description, diag)
              VALUES (?,?,?,?,?,?,?);`
	_, err := db.Exec(q, in.ID, in.Email, in.State, in.TS, ua, bd, di)
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL1", E: err})
	}

	// Store each of the SMTP events if any were provided.
	if len(in.SMTPEvents) > 0 {
		for _, sm := range in.SMTPEvents {
			dIP := sql.NullString{
				String: sm.DestinationIP,
				Valid:  sm.DestinationIP != "",
			}
			diag := sql.NullString{
				String: sm.Diag,
				Valid:  sm.Diag != "",
			}
			kind := sql.NullString{
				String: sm.Type,
				Valid:  sm.Type != "",
			}
			q = `INSERT INTO email_smtp_events (event_id, 
                          destination_ip, diag, time, kind) VALUES (?,?,?,?,?);`
			_, err = db.Exec(q, in.ID, dIP, diag, sm.TS, kind)
			if err != nil {
				e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL2", E: err})
			}
		}
	}

	// Store each of the click events if any were provided.
	if len(in.Clicks) > 0 {
		for _, cl := range in.Clicks {
			q = `INSERT INTO email_opens (event_id, kind, time) VALUES (?,?,?);`
			_, err = db.Exec(q, in.ID, cl.Kind, cl.TS)
			if err != nil {
				e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL3", E: err})
			}
		}
	}

	// Store the location if it was included.
	if in.Location != nil {
		country := sql.NullString{
			String: in.Location.Country,
			Valid:  in.Location.Country != "",
		}
		city := sql.NullString{
			String: in.Location.City,
			Valid:  in.Location.City != "",
		}
		lat := sql.NullFloat64{
			Float64: in.Location.Latitude,
			Valid:   in.Location.Latitude != 0,
		}
		long := sql.NullFloat64{
			Float64: in.Location.Longitude,
			Valid:   in.Location.Longitude != 0,
		}

		q = `INSERT INTO email_location (event_id, country, latitude, longitude, city)
                     VALUES (?,?,?,?,?);`
		_, err = db.Exec(q, in.ID, country, lat, long, city)
		if err != nil {
			e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL3", E: err})
		}
	}

	// Return if the state was not an event that requires attention.
	if in.State != "soft_bounce" && in.State != "hard_bounce" &&
		in.State != "reject" {
		return
	}
	var eve string
	switch in.State {
	case "soft_bounce":
		eve = fmt.Sprintf("%v - %v", in.BounceDescription, in.Diag)
	case "hard_bounce":
		eve = fmt.Sprintf("%v - %v", in.BounceDescription, in.Diag)
	case "reject":
		eve = "Email rejected, please use another email."
	}

	// Depending on the event, the sender may need to be notified. Construct
	// the email information where applicable and send the email.
	//// Lookup the recipient that errored, based on the event_id.
	q = `SELECT recipients.id,recipients.first_name, recipients.last_name, recipients.email
            FROM recipients
            INNER JOIN correspondences ON correspondences.recipient_id = recipients.id
            WHERE correspondences.third_party_id = ?;`
	rec := Recipient{}
	err = db.Get(&rec, q, in.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows were returned from the previous query, it means the
			// correspondence is not linked to a recipient. In this case,
			// it is a system email and will be ignored.
			return
		} else {
			e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL4", E: err})

		}
	}

	// Construct the input, and send the email.
	//// Retrieve the sender information for the email.
	note := notificationEmail{}
	q = `SELECT user.first_name, user.last_name, user.email, documents.title,
            documents.id AS document_id FROM recipients
            INNER JOIN documents ON documents.id = recipients.document_id
            INNER JOIN user ON documents.user_id = user.id
            WHERE recipients.id = ?;`

	err = db.Get(&note, q, rec.Id)
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL5", E: err})
	}
	emailInput := &buildNotificationInput{
		info:      note,
		db:        db,
		recipient: rec,
		event:     eve,
	}

	err = lc.Pvl.buildNotificationEmail(emailInput)
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL6", E: err})
	}

	// Remove the next_reminder date from the recipient so the reminder emails
	// stop being sent.
	q = `UPDATE recipients SET next_reminder = NULL WHERE id = ?;`
	_, err = db.Exec(q, rec.Id)
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRiNGESTeMAIL7", E: err})
	}

	return
}
