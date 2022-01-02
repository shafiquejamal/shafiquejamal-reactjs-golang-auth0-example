package auth

import (
	"io/ioutil"
	"net/http"

	"github.com/shafiquejamal/reactjs-golang-starter/errorhandler"
)

type UserIdentity struct {
	Username      string
	Email         string
	UserId        string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
}

type Statement struct {
	Actions   []string
	Resources []string
}

type Permission struct {
	Denys  []Statement
	Allows []Statement
}

type User struct {
	Identity    UserIdentity
	Permissions []Permission
}

func OAuthUserIdentityFetcher(ep string) func(bearerToken string, w *http.ResponseWriter) ([]byte, error) {
	return func(bearerToken string, w *http.ResponseWriter) ([]byte, error) {
		client := http.Client{}
		req, err := http.NewRequest("GET", ep, nil)
		if err != nil {
			errorhandler.ReturnError(w, http.StatusInternalServerError, "Could not create GET request for client info", err)
			return []byte{}, err
		}
		req.Header.Add("Authorization", "Bearer "+bearerToken)
		resp, err := client.Do(req)
		if err != nil {
			errorhandler.ReturnError(w, http.StatusInternalServerError, "Response error", err)
			return []byte{}, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errorhandler.ReturnError(w, http.StatusInternalServerError, "Error while reading the response bytes", err)
			return []byte{}, err
		}
		return body, nil
	}
}
