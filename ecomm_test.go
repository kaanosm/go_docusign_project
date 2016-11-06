package logic

import (
	"database/sql"
	"testing"
)

// Test the isProrated function to ensure the correct values are being
// returned.
func TestIsProrated(t *testing.T) {
	lc := &Lgc{}

	// When the user's kind is 0 or 1, they are always upgrading, so the
	// return should be true.
	uKind := "0"
	plan := "1"
	resp := lc.isProrated(uKind, plan)
	if !resp {
		t.Errorf("Wrong value being returned from isProrated - expected false, got %v when uKind is %v and plan is %v\n.", resp, uKind, plan)
	}
	uKind = "1"
	plan = "2"
	resp = lc.isProrated(uKind, plan)
	if !resp {
		t.Errorf("Wrong value being returned from isProrated - expected false, got %v when uKind is %v and plan is %v\n.", resp, uKind, plan)
	}

	// When the user's kind if 2, they can be upgrading or downgrading.
	uKind = "2"
	plan = "1"
	///// DOWNGRADING
	resp = lc.isProrated(uKind, plan)
	if resp {
		t.Errorf("Wrong value being returned from isProrated - expected true, got %v when uKind is %v and plan is %v\n.", resp, uKind, plan)
	}
	uKind = "2"
	plan = "5"
	///// UPGRADING
	resp = lc.isProrated(uKind, plan)
	if !resp {
		t.Errorf("Wrong value being returned from isProrated - expected false, got %v when uKind is %v and plan is %v\n.", resp, uKind, plan)
	}

	// When the user's kind is 5, they are always downgrading.
	uKind = "5"
	plan = "2"
	resp = lc.isProrated(uKind, plan)
	if resp {
		t.Errorf("Wrong value being returned from isProrated - expected false, got %v when uKind is %v and plan is %v\n.", resp, uKind, plan)
	}

}

// Test the getPlan function returns the expected plans.
func TestGetPlan(t *testing.T) {
	lc := Lgc{}
	plan := "101"
	sPlan := lc.getPlan(plan)
	if sPlan != "1" {
		t.Errorf("getPlan failed. Got %v wanted 1.", sPlan)
	}

	plan = "102"
	sPlan = lc.getPlan(plan)
	if sPlan != "2" {
		t.Errorf("getPlan failed. Got %v wanted 2.", sPlan)
	}

	plan = "103"
	sPlan = lc.getPlan(plan)
	if sPlan != "5" {
		t.Errorf("getPlan failed. Got %v wanted 5.", sPlan)
	}

}

// Test the EcommUpdateCustomerSub function returns error when the validations
// occur.
func TestEcommUpdate1(t *testing.T) {

	// We expect this to return error as there were no customers for that
	// user.
	u := ui{
		CID: sql.NullString{String: "", Valid: true},
	}
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
	}
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			*dest.(*ui) = u
			return nil
		},
	}
	lc := Lgc{}

	err := lc.EcommUpdateCustomerSub(db, "123", ls)
	if err.Error() != "Customer was not found for that user." {
		t.Error("EcommUpdateCustomerSub not validating the customer exists.")
	}
}

// Test the EcommUpdateCustomerSub function returns error when the validations
// occur.
func TestEcommUpdate2(t *testing.T) {

	// We expect this to return error as there were no customers for that
	// user.
	u := ui{
		CID:  sql.NullString{String: "abc123", Valid: true},
		Kind: "123",
	}
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		getPlanMock: func(string) string {
			return "1"
		},
	}
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			*dest.(*ui) = u
			return nil
		},
	}
	lc := Lgc{}

	err := lc.EcommUpdateCustomerSub(db, "123", ls)
	if err.Error() != "User is already subscribed to the plan." {
		t.Error("EcommUpdateCustomerSub not validating the customer plan is changing.")
	}
}

// Test the EcommUpdateCustomerSub function passes the correct parameters to
// the ecomm package when upgrading a business plus user.
func TestEcommUpdate3(t *testing.T) {
	// The user is a kind 2 upgrading to the plan '103' - which is
	// a kind 5 (business plus).
	// Assert that the user will be prorated.
	// Assert that the user is invoiced.
	u := ui{
		CID:  sql.NullString{String: "abc123", Valid: true},
		Kind: "2",
	}
	invoiceC := 0
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		getPlanMock: func(string) string {
			return "5"
		},
		invoiceCustomerMock: func(string) error {
			invoiceC++
			return nil
		},
		updateSubMock: func(a, b string, c bool) (string, error) {
			if a != "abc123" {
				t.Error("updateSub is not receiving the user's customer id.")
			}
			if b != "103" {
				t.Error("updateSub is not receiving the correct code for the plan.")
			}
			if !c {
				t.Error("updateSub is not receiving the user prorate value as true when they are upgrading.")
			}
			return "now", nil
		},
		isProratedMock: func(string, string) bool {
			return true
		},
	}
	// Assert that the exec function (insert) is hit 3 times.
	// 1: When the user is updated to the new kind.
	// 2: When the user enterprise is created.
	// 3: when the user enterprise_id is assigned.
	// Assert the get function is called twice.
	// 1: When the user information is retrieved.
	// 2: When the user enterprise information is being constructed for the
	// new enterprise.
	gcount := 0
	ecount := 0
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			if gcount == 0 {
				*dest.(*ui) = u
			}
			gcount++
			return nil
		},
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			ecount++
			return nil, nil
		},
	}
	lc := Lgc{}

	err := lc.EcommUpdateCustomerSub(db, "103", ls)
	if err != nil {
		t.Error("EcommUpdateCustomerSub is unexpectedly returning an error.")
	}

	if gcount != 2 {
		t.Errorf("The EcommUpdateCustomerSub is not retrieving user details, expected 2 got %v", gcount)
	}
	if ecount != 3 {
		t.Errorf("The EcommUpdateCustomerSub function is not making all db calls. Expected 4 got %v", ecount)
	}
}

// Test the EcommUpdateCustomerSub function passes the correct parameters to
// the ecomm package when upgrading a just you to a entreprenuer.
func TestEcommUpdate4(t *testing.T) {
	// The user is a kind 1 upgrading to the plan '102' - which is
	// a kind 2 (entreprenuer).
	// Assert that the user will be prorated.
	// Assert that the user is invoiced.
	u := ui{
		CID:  sql.NullString{String: "abc123", Valid: true},
		Kind: "1",
	}
	invoiceC := 0
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		getPlanMock: func(string) string {
			return "2"
		},
		invoiceCustomerMock: func(string) error {
			invoiceC++
			return nil
		},
		updateSubMock: func(a, b string, c bool) (string, error) {
			if a != "abc123" {
				t.Error("updateSub is not receiving the user's customer id.")
			}
			if b != "102" {
				t.Error("updateSub is not receiving the correct code for the plan.")
			}
			if !c {
				t.Error("updateSub is not receiving the user prorate value as true when they are upgrading.")
			}
			return "now", nil
		},
		isProratedMock: func(string, string) bool {
			return true
		},
	}
	// Assert that the exec function (insert) is hit 2 times.
	// 1: When the user is updated to the new kind.
	// 2: When the user's quota is updated.
	// Assert the get function is called once.
	// 1: When the user information is retrieved.
	gcount := 0
	ecount := 0
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			if gcount == 0 {
				*dest.(*ui) = u
			}
			gcount++
			return nil
		},
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			ecount++
			return nil, nil
		},
	}
	lc := Lgc{}

	err := lc.EcommUpdateCustomerSub(db, "102", ls)
	if err != nil {
		t.Error("EcommUpdateCustomerSub is unexpectedly returning an error.")
	}

	if gcount != 1 {
		t.Errorf("The EcommUpdateCustomerSub is not retrieving user details, expected 1 got %v", gcount)
	}
	if ecount != 2 {
		t.Errorf("The EcommUpdateCustomerSub function is not making all db calls. Expected 2 got %v", ecount)
	}
}

// Test the EcommNewCustomer function passes the correct arguments to the
// ecomm package, and makes the expected number of db calls when the
// payments were successful.
func TestEcommNewCustomer1(t *testing.T) {
	// Assert that the db call to retrieve the email is made.
	// Assert that the expected arguments are passed to the create customer
	// function.
	// Assert the db calls to update the user info and user quota are made;
	// this implicitly tests that these calls are ignored when the plan is
	// '102' - an entreprenuer account.
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		createCustomerMock: func(a, b, c string) (string, string, error) {
			if a != "a@b.com" {
				t.Error("createCustomer is not being passed the correct email argument.")
			}
			if b != "123ABC" {
				t.Error("createCustomer is not being passed the correct token.")
			}
			if c != "102" {
				t.Error("createCustomer is not being passed the correct plan.")
			}
			return "", "", nil
		},
	}
	ecount := 0
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			email := "a@b.com"
			*dest.(*string) = email
			return nil
		},
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			ecount++
			return nil, nil
		},
	}
	lc := Lgc{}

	err := lc.EcommNewCustomer(db, ls, "123ABC", "102")
	if err != nil {
		t.Error("EcommNewCustomer is returning an error unexpectedly.")
	}
	if ecount != 2 {
		t.Errorf("EcommNewCustomer is not making expected number of db calls, expected 2 got %v.", ecount)
	}

}

// Test the EcommNewCustomer function passes the correct arguments to the
// ecomm package, and makes the expected number of db calls when the
// payments were successful for a business plus upgrade.
func TestEcommNewCustomer2(t *testing.T) {
	// Assert that the db call to retrieve the email is made.
	// Assert that the expected arguments are passed to the create customer
	// function.
	// Assert the db calls to update the user info and user quota are made;
	// as well as the calls to make an enterprise record and assign the
	// enterprise to the user for a plan '103' - a business plus account.
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		createCustomerMock: func(a, b, c string) (string, string, error) {
			if a != "a@b.com" {
				t.Error("createCustomer is not being passed the correct email argument.")
			}
			if b != "123ABC" {
				t.Error("createCustomer is not being passed the correct token.")
			}
			if c != "103" {
				t.Error("createCustomer is not being passed the correct plan.")
			}
			return "", "", nil
		},
	}
	ecount := 0
	count := 0
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			if count == 0 {
				email := "a@b.com"
				*dest.(*string) = email
			}
			count++
			return nil
		},
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			ecount++
			return nil, nil
		},
	}
	lc := Lgc{}

	err := lc.EcommNewCustomer(db, ls, "123ABC", "103")
	if err != nil {
		t.Error("EcommNewCustomer is returning an error unexpectedly.")
	}
	if ecount != 4 {
		t.Errorf("EcommNewCustomer is not making expected number of db calls, expected 4 got %v.", ecount)
	}
	if count != 2 {
		t.Errorf("EcommNewCustomer is not making expected number of db retrieval calls, expected 2 got %v", count)
	}

}

// Test the EcommNewCustomer function passes the correct arguments to the
// ecomm package, and makes the expected number of db calls when the
// payments were successful for a business plus upgrade.
// Ensure that a new enterprise record is not created for a user that already has
// an enterprise.
func TestEcommNewCustomer3(t *testing.T) {
	// Assert that the db call to retrieve the email is made.
	// Assert that the expected arguments are passed to the create customer
	// function.
	// Assert the db calls to update the user info and user quota are made;
	// and the calls to make an enterprise record are NOT made.
	ls := &MockLogic{
		GetCurrentUserMock: func() *UserAuth {
			return &UserAuth{Id: "1"}
		},
		createCustomerMock: func(a, b, c string) (string, string, error) {
			if a != "a@b.com" {
				t.Error("createCustomer is not being passed the correct email argument.")
			}
			if b != "123ABC" {
				t.Error("createCustomer is not being passed the correct token.")
			}
			if c != "103" {
				t.Error("createCustomer is not being passed the correct plan.")
			}
			return "", "", nil
		},
	}
	ecount := 0
	count := 0
	db := &MockDb{
		GetMock: func(dest interface{}, query string, args ...interface{}) error {
			if count == 0 {
				email := "a@b.com"
				*dest.(*string) = email
			} else if count == 1 {
				*dest.(*User) = User{
					EnterpriseID: sql.NullString{String: "FF", Valid: true},
				}
			}
			count++
			return nil
		},
		ExecMock: func(q string, args ...interface{}) (sql.Result, error) {
			ecount++
			return nil, nil
		},
	}
	lc := Lgc{}
	err := lc.EcommNewCustomer(db, ls, "123ABC", "103")
	if err != nil {
		t.Error("EcommNewCustomer is returning an error unexpectedly.")
	}
	// There should only be 3 exec calls made, as the user already had
	// an enterprise record assigned to them. The last request will be updating
	// the user to have a group_active of 1.
	if ecount != 3 {
		t.Errorf("EcommNewCustomer is not making expected number of db calls, expected 3 got %v.", ecount)
	}
	if count != 2 {
		t.Errorf("EcommNewCustomer is not making expected number of db retrieval calls, expected 2 got %v", count)
	}

}
