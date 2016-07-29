package accounts_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/RichardKnop/example-api/accounts"
	"github.com/RichardKnop/example-api/accounts/roles"
	"github.com/RichardKnop/example-api/oauth"
	"github.com/RichardKnop/example-api/password"
	"github.com/RichardKnop/example-api/util"
	"github.com/RichardKnop/jsonhal"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func (suite *AccountsTestSuite) TestUpdateUserPasswordFailsWithBadCurrentPassword() {
	var (
		testOauthUser   *oauth.User
		testUser        *accounts.User
		testAccessToken *oauth.AccessToken
		err             error
	)

	// Insert a test user
	testOauthUser, err = suite.service.GetOauthService().CreateUser(
		"harold@finch",
		"test_password",
	)
	assert.NoError(suite.T(), err, "Failed to insert a test oauth user")
	testUser = accounts.NewUser(
		suite.accounts[0],
		testOauthUser,
		suite.userRole,
		"", // facebook ID
		"Harold",
		"Finch",
		"",    // picture
		false, // confirmed
	)
	err = suite.db.Create(testUser).Error
	assert.NoError(suite.T(), err, "Failed to insert a test user")
	testUser.Account = suite.accounts[0]
	testUser.OauthUser = testOauthUser
	testUser.Role = suite.userRole

	// Login the test user
	testAccessToken, _, err = suite.service.GetOauthService().Login(
		suite.accounts[0].OauthClient,
		testUser.OauthUser,
		"read_write", // scope
	)
	assert.NoError(suite.T(), err, "Failed to login the test user")

	payload, err := json.Marshal(&accounts.UserRequest{
		Password:    "bogus_password",
		NewPassword: "some_new_password",
	})
	assert.NoError(suite.T(), err, "JSON marshalling failed")
	r, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("http://1.2.3.4/v1/accounts/users/%d", testUser.ID),
		bytes.NewBuffer(payload),
	)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAccessToken.Token))

	// Check the routing
	match := new(mux.RouteMatch)
	suite.router.Match(r, match)
	if assert.NotNil(suite.T(), match.Route) {
		assert.Equal(suite.T(), "update_user", match.Route.GetName())
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check that the mock object expectations were met
	suite.assertMockExpectations()

	// Check the status code
	if !assert.Equal(suite.T(), 400, w.Code) {
		log.Print(w.Body.String())
	}

	// Check the response body
	expectedJSON, err := json.Marshal(
		map[string]string{"error": oauth.ErrInvalidUserPassword.Error()})
	if assert.NoError(suite.T(), err, "JSON marshalling failed") {
		assert.Equal(
			suite.T(),
			string(expectedJSON),
			strings.TrimRight(w.Body.String(), "\n"),
			"Body should contain JSON detailing the error",
		)
	}
}

func (suite *AccountsTestSuite) TestUpdateUserPasswordFailsWithPaswordlessUser() {
	var (
		testOauthUser   *oauth.User
		testUser        *accounts.User
		testAccessToken *oauth.AccessToken
		err             error
	)

	// Insert a test user
	testOauthUser, err = suite.service.GetOauthService().CreateUser(
		"harold@finch",
		"", // empty password
	)
	assert.NoError(suite.T(), err, "Failed to insert a test oauth user")
	testUser = accounts.NewUser(
		suite.accounts[0],
		testOauthUser,
		suite.userRole,
		"some_facebook_id", // facebook ID
		"Harold",
		"Finch",
		"",    // picture
		false, // confirmed
	)
	err = suite.db.Create(testUser).Error
	assert.NoError(suite.T(), err, "Failed to insert a test user")
	testUser.Account = suite.accounts[0]
	testUser.OauthUser = testOauthUser
	testUser.Role = suite.userRole

	// Login the test user
	testAccessToken, _, err = suite.service.GetOauthService().Login(
		suite.accounts[0].OauthClient,
		testUser.OauthUser,
		"read_write", // scope
	)
	assert.NoError(suite.T(), err, "Failed to login the test user")

	payload, err := json.Marshal(&accounts.UserRequest{
		Password:    "",
		NewPassword: "some_new_password",
	})
	assert.NoError(suite.T(), err, "JSON marshalling failed")
	r, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("http://1.2.3.4/v1/accounts/users/%d", testUser.ID),
		bytes.NewBuffer(payload),
	)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAccessToken.Token))

	// Check the routing
	match := new(mux.RouteMatch)
	suite.router.Match(r, match)
	if assert.NotNil(suite.T(), match.Route) {
		assert.Equal(suite.T(), "update_user", match.Route.GetName())
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check that the mock object expectations were met
	suite.assertMockExpectations()

	// Check the status code
	if !assert.Equal(suite.T(), 400, w.Code) {
		log.Print(w.Body.String())
	}

	// Check the response body
	expectedJSON, err := json.Marshal(
		map[string]string{"error": oauth.ErrUserPasswordNotSet.Error()})
	if assert.NoError(suite.T(), err, "JSON marshalling failed") {
		assert.Equal(
			suite.T(),
			string(expectedJSON),
			strings.TrimRight(w.Body.String(), "\n"),
			"Body should contain JSON detailing the error",
		)
	}
}

func (suite *AccountsTestSuite) TestUpdateUserPassword() {
	var (
		testOauthUser   *oauth.User
		testUser        *accounts.User
		testAccessToken *oauth.AccessToken
		err             error
	)

	// Insert a test user
	testOauthUser, err = suite.service.GetOauthService().CreateUser(
		"harold@finch",
		"test_password",
	)
	assert.NoError(suite.T(), err, "Failed to insert a test oauth user")
	testUser = accounts.NewUser(
		suite.accounts[0],
		testOauthUser,
		suite.userRole,
		"", // facebook ID
		"Harold",
		"Finch",
		"",    // picture
		false, // confirmed
	)
	err = suite.db.Create(testUser).Error
	assert.NoError(suite.T(), err, "Failed to insert a test user")
	testUser.Account = suite.accounts[0]
	testUser.OauthUser = testOauthUser
	testUser.Role = suite.userRole

	// Login the test user
	testAccessToken, _, err = suite.service.GetOauthService().Login(
		suite.accounts[0].OauthClient,
		testUser.OauthUser,
		"read_write", // scope
	)
	assert.NoError(suite.T(), err, "Failed to login the test user")

	payload, err := json.Marshal(&accounts.UserRequest{
		Password:    "test_password",
		NewPassword: "some_new_password",
	})
	assert.NoError(suite.T(), err, "JSON marshalling failed")
	r, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("http://1.2.3.4/v1/accounts/users/%d", testUser.ID),
		bytes.NewBuffer(payload),
	)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAccessToken.Token))

	// Check the routing
	match := new(mux.RouteMatch)
	suite.router.Match(r, match)
	if assert.NotNil(suite.T(), match.Route) {
		assert.Equal(suite.T(), "update_user", match.Route.GetName())
	}

	// Count before
	var countBefore int
	suite.db.Model(new(accounts.User)).Count(&countBefore)

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check that the mock object expectations were met
	suite.assertMockExpectations()

	// Check the status code
	if !assert.Equal(suite.T(), 200, w.Code) {
		log.Print(w.Body.String())
	}

	// Count after
	var countAfter int
	suite.db.Model(new(accounts.User)).Count(&countAfter)
	assert.Equal(suite.T(), countBefore, countAfter)

	// Fetch the updated user
	user := new(accounts.User)
	notFound := accounts.UserPreload(suite.db).First(user, testUser.ID).RecordNotFound()
	assert.False(suite.T(), notFound)

	// Check that the password has changed
	assert.Error(suite.T(), password.VerifyPassword(
		user.OauthUser.Password.String,
		"test_password",
	))
	assert.NoError(suite.T(), password.VerifyPassword(
		user.OauthUser.Password.String,
		"some_new_password",
	))

	// And the user meta data is unchanged
	assert.Equal(suite.T(), "harold@finch", user.OauthUser.Username)
	assert.Equal(suite.T(), "Harold", user.FirstName.String)
	assert.Equal(suite.T(), "Finch", user.LastName.String)
	assert.Equal(suite.T(), roles.User, user.Role.ID)
	assert.False(suite.T(), user.Confirmed)

	// Check the response body
	expected := &accounts.UserResponse{
		Hal: jsonhal.Hal{
			Links: map[string]*jsonhal.Link{
				"self": &jsonhal.Link{
					Href: fmt.Sprintf("/v1/accounts/users/%d", user.ID),
				},
			},
		},
		ID:        user.ID,
		Email:     "harold@finch",
		FirstName: "Harold",
		LastName:  "Finch",
		Role:      roles.User,
		Confirmed: false,
		CreatedAt: util.FormatTime(user.CreatedAt),
		UpdatedAt: util.FormatTime(user.UpdatedAt),
	}
	expectedJSON, err := json.Marshal(expected)
	if assert.NoError(suite.T(), err, "JSON marshalling failed") {
		assert.Equal(
			suite.T(),
			string(expectedJSON),
			strings.TrimRight(w.Body.String(), "\n"), // trim the trailing \n
		)
	}
}
