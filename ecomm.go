package logic

import (
	"github.com/dchest/uniuri"

	"database/sql"
	"errors"
	"fmt"
	"pleasesign/config"
	"pleasesign/ecomm"
	e "pleasesign/errlogger"
	"strconv"
	"time"
)

/*
  EcommNewCustomer will create a new customer for the current user, subscribe
  them to the plan provided, and align their token to the customer.
  This will also update the user Kind to the correct number.
*/
func (lc Lgc) EcommNewCustomer(db DataCaller, ls LogicStore, token string, plan string) error {
	userID := ls.GetCurrentUser().Id

	// Get the email for the user to save against the Stripe
	// customer.
	q := `SELECT email FROM user WHERE id = ?;`
	var em string
	err := db.Get(&em, q, userID)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when creating the new subscription.", E: err})
	}

	// Determine the user kind/plan before updating.
	// Plan is an abritrary string relating to the plans/kinds of the user.
	// Set the user quota depending on their new user kind.
	var k int
	quota := 0
	switch plan {
	case "101":
		quota = config.T1Lim
		k = 1
	case "102":
		quota = config.T2Lim
		k = 2
	case "103":
		k = 5
	}

	// Create a customer subscribed to the plan.
	cID, bEnd, err := ls.createCustomer(em, token, plan)
	if err != nil {
		// This is a friendly error returned from Stripe.
		return err
	}

	// Update the user with their customer_id and user kind.
	q = `UPDATE user SET s_customer_id=?, kind=?, next_kind=?, current_period_end=?
          WHERE id=?;`
	_, err = db.Exec(q, cID, k, k, bEnd, userID)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when creating the new subscription.", E: err})
	}

	// Update the user quota to the relevant amount if they have upgraded
	// to a 1/2 plan.
	q = `UPDATE user_quota SET quota = ? WHERE user_id = ?;`
	_, err = db.Exec(q, quota, userID)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when creating the new subscription.", E: err})
	}

	// If the user is upgrading to a bplus user, create an enterprise record
	// for their branding/sender settings information.
	if k == 5 {
		if err := createGroup(db, userID); err != nil {
			return e.ThrowError(&e.LogInput{M: "Error when creating the new subscription.", E: err})
		}
	}

	return nil
}

// The output for the EcommGetCustomerSub function.
type EcommSubscription struct {
	Trans          bool
	Next           int
	CurEnd         string
	PayingCustomer bool
	LastFour       string
	Brand          string
	ExMon          string
	ExYr           string
	FailM          string
	City           string
	Address        string
	PostCode       string
	Country        string
	FirstName      string
	LastName       string
	Email          string
	Quantity       int
}

/*
  EcommGetCustomerSub returns the subscription information for the current user.
  This is used on the plan page heavily to ensure the correct links are displayed
  to the user, and any card/failed payment information is returned.
*/
func (lc Lgc) EcommGetCustomerSub(db DataCaller, ls LogicStore) (*EcommSubscription, error) {
	userID := ls.GetCurrentUser().Id

	// We need the user's information before continuing.
	type c struct {
		CID            sql.NullString `db:"s_customer_id"`
		NextKind       sql.NullString `db:"next_kind"`
		Kind           string         `db:"kind"`
		FailedPayments int            `db:"failed_payments"`
		FirstName      string         `db:"first_name"`
		LastName       string         `db:"last_name"`
		Email          string         `db:"email"`
	}
	var cus c
	q := `SELECT first_name, last_name, email, s_customer_id, 
          next_kind, kind, failed_payments 
          FROM user WHERE id = ?;`
	err := db.Get(&cus, q, userID)
	if err != nil {
		return nil, e.ThrowError(&e.LogInput{M: "ERRGETCUSTOMER1", E: err})
	} else if cus.CID.String == "" {
		return &EcommSubscription{}, nil
	}

	// If the user has failed payments, retrieved the latest failure
	// message for the subscription. We will return this to the user
	// so they can action their payment. The page also has cases in
	// regarding to this information.
	var failM string
	if cus.FailedPayments != 0 {
		type m struct {
			FailureMessage sql.NullString `db:"failure_message"`
			FailureCode    sql.NullString `db:"failure_code"`
		}
		var msg m
		q = `SELECT failure_message, failure_code FROM ecomm_failures 
                WHERE s_customer_id = ? ORDER BY created DESC;`
		err = db.Get(&msg, q, cus.CID.String)
		if err != nil {
			return nil, e.ThrowError(&e.LogInput{M: "ERRGETCUSTOMER1a", E: err})
		}
		failM = fmt.Sprintf("%v - %v", msg.FailureCode.String, msg.FailureMessage.String)
	}

	// Get the subscription information.
	sub, err := ecomm.GetSub(cus.CID.String)
	if err != nil {
		return nil, e.ThrowError(&e.LogInput{M: "ERRGETCUSTOMER2", E: err})
	}

	// The page expects a number, so we need to convert it back.
	nk, _ := strconv.Atoi(cus.NextKind.String)

	// If the current kind is different to the next kind, return true as
	// the user will be moving to a different kind at the end of the
	// billing period. The front end cases this to give feedback to the user.
	tr := cus.NextKind.String != cus.Kind
	out := &EcommSubscription{
		Trans:          tr,
		CurEnd:         sub.CurEnd,
		Next:           nk,
		PayingCustomer: true,
		LastFour:       sub.LastFour,
		Brand:          sub.Brand,
		ExMon:          sub.ExMon,
		ExYr:           sub.ExYr,
		FailM:          failM,
		City:           sub.City,
		Address:        sub.Address,
		PostCode:       sub.PostCode,
		Country:        sub.Country,
		FirstName:      cus.FirstName,
		LastName:       cus.LastName,
		Email:          cus.Email,
		Quantity:       int(sub.Quantity),
	}

	return out, nil
}

// Get the customerID and kind for a user.
type ui struct {
	CID  sql.NullString `db:"s_customer_id"`
	Kind string         `db:"kind"`
}

/*
  EcommUpdateCustomerSub updates the plan for a subscription to a different
  plan.
  This does not handle downgrading to free, which is considered cancelling.
  This handles if the customer has not currently got an active subscription,
  and will make a new subscription for them.
*/
func (lc Lgc) EcommUpdateCustomerSub(db DataCaller, plan string, ls LogicStore) error {
	userID := ls.GetCurrentUser().Id

	// Create a function that will be used when the customer needs to
	// be invoiced (when they are upgrading).
	inv := func(customerID string) error {
		err := ls.invoiceCustomer(customerID)
		return err
	}

	var u ui
	q := `SELECT s_customer_id, kind FROM user WHERE id = ?;`
	err := db.Get(&u, q, userID)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error updating the subscription.", E: err})
	}

	if u.CID.String == "" {
		return errors.New("Customer was not found for that user.")
	}

	// Get the user kind from the plan name. The plan input is the
	// plan expected by the ecomm package (which is the Stripe plan id), so
	// getPlan is used to convert to the expected plan that will be
	// stored in the database.
	sPlan := ls.getPlan(plan)

	// If the user is attempting to update to a plan they are already on,
	// return an error.
	if u.Kind == plan {
		return errors.New("User is already subscribed to the plan.")
	}

	// Determine if the user needs to be prorated (in the case of an upgrade
	// this will true). This is passed to the ecomm package to ensure the
	// user is charged accordingly; and is used below to check if
	// the user is down/up grading when setting the downgrade_date.
	pro := ls.isProrated(u.Kind, sPlan)

	// Now we can update the customer plan through the ecomm package.
	cpe, err := ls.updateSub(u.CID.String, plan, pro)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when updating the subscription.", E: err})
	}

	// Check the user's kind, to determine if they should be updated now,
	// or after the end of the financial period. See each case for
	// an explanation. This changes what the kind, next_kind, and downgrade_date
	// will be set to.
	var cur string // Denotes the kind to assign to the user.
	n := sPlan     // The next plan will always be whichever the user has chosen
	switch u.Kind {
	case "0":
		// If the user is currently a 0, they will always be upgrading.
		// So the kind will become whichever plan the user is upgrading to.
		cur = sPlan
	case "1":
		// If the user is currently a 1, they will always be upgrading.
		// So the kind will become whichever plan the user is upgrading to.
		cur = sPlan
		// The user is upgrading from a paid plan to a higher rate paid
		// plan, invoice the customer to pay the prorate immediately.
		if err := inv(u.CID.String); err != nil {
			return e.ThrowError(&e.LogInput{M: "Error when updating the subscription.", E: err})
		}
	case "2":
		// If the user is currently a 2, they will either be downgrading to a
		// 1, or upgrading to a 5.
		if sPlan == "5" {
			// So the kind will become whichever plan the user is upgrading to.
			cur = sPlan
			// The user is upgrading from a paid plan to a higher rate paid
			// plan, invoice the customer to pay the prorate immediately.
			if err := inv(u.CID.String); err != nil {
				return e.ThrowError(&e.LogInput{M: "Error updating the subscription.", E: err})
			}
		} else {
			cur = u.Kind
		}
	case "5":
		// If you're a 5, you're always downgrading.
		cur = u.Kind
	}

	// This will be the new quota for the user, and will only
	// be updated when they are upgrading to a 2 or 3 (and ignored
	// if they are downgrading).
	quota := 0
	if sPlan == "1" {
		quota = config.T1Lim
	}
	if sPlan == "2" {
		quota = config.T2Lim
	}

	// If this is an upgrade, we need to remove the downgrade_date to ensure
	// the user will not be picked up by the scheduler.
	// We also need to set the new user quota for the user when they upgrade.
	// Otherwise we do so it will be picked up by the scheduling service.
	// Update the kind and next_kind of the user to the chosen plan.
	if pro {
		q = `UPDATE user SET kind=?, next_kind=?, current_period_end=?, downgrade_date=NULL
          WHERE id=?;`
		_, err = db.Exec(q, cur, n, cpe, userID)
		if err != nil {
			return e.ThrowError(&e.LogInput{M: "Error when updating the subscription.", E: err})
		}

		// Update the user quota to the relevant amount if they have upgraded
		// to a 1 or 2 plan.
		if cur != "5" {
			q = `UPDATE user_quota SET quota = ? WHERE user_id = ?;`
			_, err = db.Exec(q, quota, userID)
			if err != nil {
				return e.ThrowError(&e.LogInput{M: "Error when updating the new subscription.", E: err})
			}
		}

	} else {
		q = `UPDATE user SET kind=?, next_kind=?, current_period_end=?, downgrade_date=?
          WHERE id=?;`
		_, err = db.Exec(q, cur, n, cpe, cpe, userID)
		if err != nil {
			return e.ThrowError(&e.LogInput{M: "Error when updating the subscription.", E: err})
		}
	}

	// If the user is upgrading to a bplus user, create an enterprise record
	// for their branding/sender settings information.
	if cur == "5" {
		if err := createGroup(db, userID); err != nil {
			return e.ThrowError(&e.LogInput{M: "Error when creating the new subscription.", E: err})
		}
	}

	return nil
}

/*
  EcommCancelCustomerSub cancels the plan for a customer, which will end their
  subscription on the next billing date. This also updates their next_kind to
  0, indicating they are transitioning. This is separate to the
  EcommUpgradeCustomerSub function as it uses a different method within the
  ecomm package (the ecomm.CancelSub method causes the subscription to cancel
  at the end of the billing period).
*/
func (lc Lgc) EcommCancelCustomerSub(db DataCaller) error {
	cID, userID, err := getCustomerID(db, lc)
	if err != nil || cID == "" {
		return errors.New("Error when cancelling the subscription.")
	}

	// Cancel the subscription within the ecomm package.
	if err := ecomm.CancelSub(cID); err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when cancelling the subscription.", E: err})
	}

	// Update the next_kind of the user to the free plan.
	q := `UPDATE user SET next_kind = 0, downgrade_date = current_period_end 
          WHERE id = ?;`
	_, err = db.Exec(q, userID)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "Error when cancelling the subscription.", E: err})
	}

	return nil
}

// The input for the EcommUpdateCustomerInfo function.
// This input is intended to be expanded upon.
type EcommUpdateCInput struct {
	Token string
}

/*
  EcommUpdateCustomerInfo will take a new token and update this for the customer.
  This then becomes the default source for that customer, and will retry any
  failed payments. We need to offer this to users when their card has failed
  charges.
*/
func (lc Lgc) EcommUpdateCustomerInfo(db DataCaller, in *EcommUpdateCInput) error {

	cID, _, err := getCustomerID(db, lc)
	if err != nil || cID == "" {
		return errors.New("This customer does not exist.")
	}

	inp := ecomm.UpdateCusInp{CustomerID: cID, Token: in.Token}
	err = ecomm.UpdateCustomerInfo(inp)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "ERRUPDCUSTOMER2", E: err})
	}

	return nil
}

// The output from the EcommGetCustomerInv function.
type EcommInvoice struct {
	ID        string
	Total     int64
	AmountDue int64
	Date      string
	Paid      bool
	Lines     []EcommLineItem
}
type EcommLineItem struct {
	ID          string
	Amount      int64
	Proration   bool
	Description string
	Plan        string
}

/*
  EcommGetCustomerInv returns all previous invoices for the current user.
  The response from the ecomm package isn't implicitly friendly to the
  controller so we format the response before returning it.

  @return []EcommInvoice is an array formatted as expected by the controller.
  @return error is a friendly error that is returned to the user.
*/
func (lc Lgc) EcommGetCustomerInv(db DataCaller, ls LogicStore) ([]EcommInvoice, error) {
	var out []EcommInvoice

	cID, _, err := getCustomerID(db, ls)
	if err != nil {
		return out, e.ThrowError(&e.LogInput{
			M: "Error retrieving customer invoices.",
			E: err,
		})
	} else if cID == "" {
		return out, errors.New("Not an existing customer.")
	}

	inv, err := ecomm.GetCustomerInvoices(cID)
	if err != nil {
		return out, e.ThrowError(&e.LogInput{M: "Error retrieving customer invoices.", E: err})
	}

	// We need to format the invoices to the expected output for controller.
	for _, i := range inv {
		in := EcommInvoice{
			ID:        i.ID,
			Total:     i.Total,
			AmountDue: i.Amount,
			Date:      i.Date,
			Paid:      i.Paid,
		}
		for _, line := range i.Lines {
			li := EcommLineItem{
				ID:          line.ID,
				Amount:      line.Amount,
				Proration:   line.Proration,
				Description: line.Description,
				Plan:        line.Plan,
			}
			in.Lines = append(in.Lines, li)
		}
		out = append(out, in)
	}
	return out, nil
}

/*
  EcommHook parses and stores the webhook events sent from the Stripe
  service, and handles user account actions as necessary.
*/
func (lc Lgc) EcommHook(db DataCaller, eventID string) {
	// Get event from the ecomm package, this implicity ensures that the
	// webhook event sent is legitimate.
	in, err := ecomm.GetHookEvent(eventID)
	if err != nil {
		e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK0", E: err})
		return
	}

	// We need a function to generate the sql null string for possible
	// null values.
	toNS := func(s string) sql.NullString {
		return sql.NullString{String: s, Valid: s != ""}
	}

	// Get the user information for the given customer as we will need this to
	// perform the required changes.
	q := `SELECT id, email, first_name FROM user WHERE s_customer_id = ?;`
	type u struct {
		ID        string `db:"id"`
		Email     string `db:"email"`
		FirstName string `db:"first_name"`
	}
	var user u
	if err := db.Get(&user, q, in.CustomerID); err != nil {
		e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK2", E: err})
		return
	}

	// Depending on the kind of the webhook, certain events need to
	// occur for the user.
	switch in.Kind {
	case "customer.subscription.deleted":
		// Update the user kind to 0 with the given s_customer_id, and remove
		// the current_period_end.
		q = `UPDATE user SET kind = 0, current_period_end = NULL WHERE id = ?;`
		if _, err := db.Exec(q, user.ID); err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK3", E: err})
			return
		}
		// Email the user informing them the account type has been
		// downgraded to a free account.
		// TODO create the email event.
	case "charge.failed":
		// Write the failure event to the database.
		q = `INSERT INTO ecomm_failures
                  (event_id, created, s_customer_id, kind, invoice, failure_message, failure_code)
                  VALUES (?,?,?,?,?,?,?);`

		if _, err := db.Exec(q,
			in.ID,
			in.Created,
			in.CustomerID,
			in.Kind,
			toNS(in.Invoice),
			toNS(in.FailureMessage),
			toNS(in.FailureCode),
		); err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK1", E: err})
			return
		}

		// Add to the payment_failed count for the user.
		q = `UPDATE user SET failed_payments = failed_payments + 1 WHERE id = ?;`
		if _, err := db.Exec(q, user.ID); err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK3", E: err})
			return
		}

		// Email the user informing them of the failed payment.
		emIn := &buildPaymentFailInput{
			firstName: user.FirstName,
			email:     user.Email,
			db:        db,
		}
		if err := lc.Pvl.buildPaymentFailEmail(emIn); err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK4", E: err})
			return
		}
	case "invoice.payment_succeeded":
		// Update the user failed_payments to 0.
		// Update the current_period_end to the appropriate end date.
		sub, err := ecomm.GetSub(in.CustomerID)
		if err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK5", E: err})
			return
		}
		q = `UPDATE user SET failed_payments = 0, current_period_end = ? WHERE id = ?;`
		if _, err := db.Exec(q, sub.CurEnd, user.ID); err != nil {
			e.ThrowError(&e.LogInput{M: "ERRECOMMHOOK6", E: err})
			return
		}
	}
	return
}

// getCustomerID returns the Ecomm customer_id for the current user, if the
// user has an existing customer.
func getCustomerID(db DataCaller, lc LogicStore) (string, string, error) {
	userID := lc.GetCurrentUser().Id

	// Get the customerID for the user.
	var cID sql.NullString
	q := `SELECT s_customer_id FROM user WHERE id = ?;`
	err := db.Get(&cID, q, userID)
	if err != nil {
		return "", userID, e.ThrowError(&e.LogInput{M: "ERRUPDCUSTOMER1", E: err})
	}

	return cID.String, userID, nil
}

// EcommSchedule handles the downgrading of user accounts when they need
// to be downgraded.
func (lc Lgc) EcommSchedule(db DataCaller) error {
	now := time.Now()

	// A user needs to be downgraded when their downgrade_date
	// has passed, and their next_kind is different
	// to their current kind. This occurs when the user has
	// downgraded their account, and has not yet be downgraded.
	// Also set the downgrade_date to NULL as this downgrade has been
	// completed.
	q := `UPDATE user SET user.kind = user.next_kind, user.downgrade_date = NULL 
        WHERE user.downgrade_date <= ? AND user.next_kind != user.kind;`
	_, err := db.Exec(q, now)
	if err != nil {
		return e.ThrowError(&e.LogInput{M: "ERRDOWNGRADE", E: err})
	}
	return nil
}

/*
  isProrated returns a boolean indicating if the user should be prorated or
  not.
  If they are upgrading, we want to prorate them.
  If they are downgrading, we do NOT want to prorate them (they get
  the benefits of a subscription until the end of the billing period).
  A return value of true indicates the user should be prorated.
*/
func (lc Lgc) isProrated(uKind string, plan string) bool {
	// Determine if the user is upgrading or downgrading.
	var upg bool

	// If the user is 0 or 1, they are always upgrading.
	if uKind == "0" || uKind == "1" {
		upg = true
	} else if uKind == "5" {
		// If the user is a 5, they are always downgrading.
		upg = false
	} else if uKind == "2" {
		// If the user is a 2, they could either be upgrading to a 5,
		// or downgrading to a 1.
		if plan == "5" {
			upg = true
		} else {
			upg = false
		}

	}

	return upg
}

// getPlan returns a string of the plan kind that is stored in the db based
// on the stripe plan id.
func (lc Lgc) getPlan(plan string) string {
	var sPlan string
	switch plan {
	case "101":
		sPlan = "1"
	case "102":
		sPlan = "2"
	case "103":
		sPlan = "5"
	}
	return sPlan
}

// updateSub wraps the ecomm package update function, and returns any errors or
// information that is returned from stripe.
func (lc Lgc) updateSub(customerID string, plan string, pro bool) (string, error) {
	if config.Context() != "test" {
		cpe, err := ecomm.UpdateSub(customerID, plan, pro)
		return cpe, err
	} else {
		return "2020-01-01 00:00:00", nil
	}
}

// invoiceCustomer wraps the ecomm package invoice function, and returns any
// errors that are returned from stripe.
func (lc Lgc) invoiceCustomer(customerID string) error {
	if config.Context() != "test" {
		err := ecomm.InvoiceCustomer(customerID)
		return err
	} else {
		return nil
	}
}

// createCustomer wraps the ecomm package create function, returning any
// data that is returned from ecomm.
func (lc Lgc) createCustomer(email string, token string, plan string) (string, string, error) {
	if config.Context() != "test" {
		customerID, billingEnd, err := ecomm.CreateCustomer(email, token, plan)
		return customerID, billingEnd, err
	} else {
		return "ABC123", "2020-01-01 00:00:00", nil
	}
}

// createGroup will create a group with default information for the provided
// userID. This is used when the user is upgrading to to a business plus.
func createGroup(db DataCaller, userID string) error {
	// Gather the user information to build the default business
	// information for the user.
	q := `SELECT first_name, last_name, email, enterprise_id FROM user WHERE id = ?;`
	var bplus User
	if err := db.Get(&bplus, q, userID); err != nil {
		return err
	}

	// If the user already has a group record, skip creating
	// one, as they will assume the previous group information.
	// Update them to be active within their group - there must always be
	// an active group member.
	if bplus.EnterpriseID.String != "" {
		q = `UPDATE user SET group_active=1 WHERE id=?;`
		if _, err := db.Exec(q, userID); err != nil {
			return err
		}
		return nil
	}

	// Put together the default info for the enterprise record.
	// The name of the enterprise record is the first and last
	// names.
	// The address is their email, and the contact information
	// is their email and name.
	// All of this can be updated once the record is created.
	// The enterprise record is created with a default of 1
	// seat. This is able to be upgraded when the user purchases
	// more seats for their enterprise.
	id := uniuri.New()
	name := fmt.Sprintf("%v %v", bplus.FirstName, bplus.LastName)
	address := bplus.Email
	contact := fmt.Sprintf("%v - %v", name, bplus.Email)
	q = `INSERT INTO enterprises (id, name, address, contact, seats) VALUES (?,?,?,?,?);`
	if _, err := db.Exec(q, id, name, address, contact, 1); err != nil {
		return err
	}

	// Update the user record to align it with the new enterprise_id.
	// Update the user record to be active within their group. This ensures
	// the group always has an active member.
	q = `UPDATE user SET enterprise_id=?, group_active=1 WHERE id=?;`
	if _, err := db.Exec(q, id, userID); err != nil {
		return err
	}
	return nil
}
