package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type (
	AuthMethod struct {
		AuthProviders    []AuthProvider `json:"authProviders"`
		UsernamePassword bool           `json:"usernamePassword"`
		EmailPassword    bool           `json:"emailPassword"`
		OnlyVerified     bool           `json:"onlyVerified"`
	}

	AuthProvider struct {
		Name                string `json:"name"`
		DisplayName         string `json:"displayName"`
		State               string `json:"state"`
		AuthURL             string `json:"authUrl"`
		CodeVerifier        string `json:"codeVerifier"`
		CodeChallenge       string `json:"codeChallenge"`
		CodeChallengeMethod string `json:"codeChallengeMethod"`
	}
)

// ListAuthMethods returns all available collection auth methods.
func (c *Collection[T]) ListAuthMethods() (AuthMethod, error) {
	var response AuthMethod
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.BaseCollectionPath + "/auth-methods")
	if err != nil {
		return response, fmt.Errorf("[records] can't send ListAuthMethods request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal response, err %w", err)
	}
	return response, nil
}

type (
	AuthWithPasswordResponse struct {
		Record Record `json:"record"`
		Token  string `json:"token"`
	}

	Record struct {
		Avatar          string `json:"avatar"`
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		Created         string `json:"created"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		ID              string `json:"id"`
		Name            string `json:"name"`
		Updated         string `json:"updated"`
		Username        string `json:"username"`
		Verified        bool   `json:"verified"`
	}
)

// AuthWithPassword authenticate a single auth collection record via its username/email and password.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the AuthWithPasswordResponse
// - the authenticated record model
func (c *Collection[T]) AuthWithPassword(username string, password string) (AuthWithPasswordResponse, error) {
	var response AuthWithPasswordResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"identity": username,
			"password": password,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-password")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-password request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-password: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-password-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

type AuthWithOauth2Response struct {
	Token string `json:"token"`
}

// AuthWithOAuth2Code authenticate a single auth collection record with OAuth2 code.
//
// If you don't have an OAuth2 code you may also want to check `authWithOAuth2` method.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the model
// - the authenticated record model
// - the OAuth2 account data (eg. name, email, avatar, etc.)
func (c *Collection[T]) AuthWithOAuth2Code(provider string, code string, codeVerifier string, redirectURL string) (AuthWithOauth2Response, error) {
	var response AuthWithOauth2Response
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"provider":     provider,
			"code":         code,
			"codeVerifier": codeVerifier,
			"redirectUrl":  redirectURL,
			//"createData":   createData,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-oauth2")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-oauth2 request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-oauth2: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-oauth2-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

type AuthRefreshResponse struct {
	Record struct {
		Avatar          string `json:"avatar"`
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		Created         string `json:"created"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		ID              string `json:"id"`
		Name            string `json:"name"`
		Updated         string `json:"updated"`
		Username        string `json:"username"`
		Verified        bool   `json:"verified"`
	} `json:"record"`
	Token string `json:"token"`
}

// AuthRefresh refreshes the current authenticated record instance and
// * returns a new token and record data.
func (c *Collection[T]) AuthRefresh() (AuthRefreshResponse, error) {
	var response AuthRefreshResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/auth-refresh")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-refresh request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-refresh: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-refresh-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

// RequestVerification sends auth record verification email request.
func (c *Collection[T]) RequestVerification(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send request-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmVerification confirms auth record email verification request.
//
// If the current `client.authStore.model` matches with the auth record from the token,
// then on success the `client.authStore.model.verified` will be updated to `true`.
func (c *Collection[T]) ConfirmVerification(verificationToken string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token": verificationToken,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestPasswordReset sends auth record password reset request
func (c *Collection[T]) RequestPasswordReset(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send request-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmPasswordReset confirms auth record password reset request.
func (c *Collection[T]) ConfirmPasswordReset(passwordResetToken string, password string, passwordConfirm string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":           passwordResetToken,
			"password":        password,
			"passwordConfirm": passwordConfirm,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestEmailChange sends an email change request to the authenticated record model.
func (c *Collection[T]) RequestEmailChange(newEmail string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"newEmail": newEmail,
		}).
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/request-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send request-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmEmailChange confirms auth record's new email address.
func (c *Collection[T]) ConfirmEmailChange(emailChangeToken string, password string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":    emailChangeToken,
			"password": password,
		}).
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/confirm-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

type ExternalAuthRequest struct {
	ID           string `json:"id"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	RecordID     string `json:"recordId"`
	CollectionID string `json:"collectionId"`
	Provider     string `json:"provider"`
	ProviderID   string `json:"providerId"`
}

// ListExternalAuths lists all linked external auth providers for the specified auth record.
func (c *Collection[T]) ListExternalAuths(recordID string) ([]ExternalAuthRequest, error) {
	var response []ExternalAuthRequest
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.baseCrudPath() + url.QueryEscape(recordID) + "/external-auths")
	if err != nil {
		return response, fmt.Errorf("[records] can't send list external-auths request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase request for list external-auths returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal list external-auths response, err %w", err)
	}
	return response, nil
}

// UnlinkExternalAuth unlink a single external auth provider from the specified auth record.
func (c *Collection[T]) UnlinkExternalAuth(recordID string, provider string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Delete(c.baseCrudPath() + url.QueryEscape(recordID) + "/external-auths/" + url.QueryEscape(provider))
	if err != nil {
		return fmt.Errorf("[records] can't send unlink-external-auth-request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at unlink-external-auth-: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

func (c *Collection[T]) baseCollectionPath() string {
	return c.BaseCollectionPath
}

func (c *Collection[T]) baseCrudPath() string {
	return c.BaseCollectionPath + "/records/"
}
