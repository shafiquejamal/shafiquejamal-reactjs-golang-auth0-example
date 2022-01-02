package auth_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/shafiquejamal/reactjs-golang-starter/auth"
	. "github.com/smartystreets/goconvey/convey"
)

var errorOutputForTesting = ""

func createUser(u *auth.User, allowActions1, allowActions2, denyActions1, denyActions2 []string) {
	allowStatement1 := auth.Statement{
		Actions:   allowActions1,
		Resources: []string{".*"},
	}
	denyStatement1 := auth.Statement{
		Actions:   denyActions1,
		Resources: []string{".*"},
	}
	permission1 := auth.Permission{
		Denys:  []auth.Statement{{}, denyStatement1},
		Allows: []auth.Statement{allowStatement1, {}},
	}
	allowStatement2 := auth.Statement{
		Actions:   allowActions2,
		Resources: []string{".*"},
	}
	denyStatement2 := auth.Statement{
		Actions:   denyActions2,
		Resources: []string{".*"},
	}
	permission2 := auth.Permission{
		Denys:  []auth.Statement{denyStatement2, {}},
		Allows: []auth.Statement{{}, allowStatement2},
	}
	(*u).Permissions = []auth.Permission{permission1, permission2}
}

func testBadUserIdentityDataFetcher(bearerToken string, w *http.ResponseWriter) ([]byte, error) {
	return []byte{}, nil
}

func testGoodUserIdentityDataFetcher(emailVerified bool) func(bearerToken string, w *http.ResponseWriter) ([]byte, error) {
	return func(bearerToken string, w *http.ResponseWriter) ([]byte, error) {
		u := auth.UserIdentity{
			EmailVerified: emailVerified,
		}
		return json.Marshal(u)
	}
}

func testBadUserDataFetcher(userI *auth.UserIdentity, u *auth.User, w *http.ResponseWriter) error {
	return errors.New("bad user data fetecher")
}

func testGoodUserDataFetcher(userI *auth.UserIdentity, u *auth.User, w *http.ResponseWriter) error {
	allowStatement := auth.Statement{
		Actions:   []string{"^/ping$", "^/pong$"},
		Resources: []string{".*"},
	}
	denyStatement := auth.Statement{
		Actions:   []string{"/pung"},
		Resources: []string{".*"},
	}
	permission := auth.Permission{
		Denys:  []auth.Statement{denyStatement},
		Allows: []auth.Statement{allowStatement},
	}
	(*u).Identity = *userI
	(*u).Permissions = []auth.Permission{permission}
	return nil
}

func testAlwaysDenyAuthorizationStrategy(user auth.User, r *http.Request) error {
	return errors.New("Always deny")
}

func TestSpec(t *testing.T) {

	Convey("PolicyAuthorizationStrategy", t, func() {
		apiPrefix := "/api"

		Convey("returns no error when", func() {
			Convey("An allow action matches the URL path, and the deny actions do not - permissions in first element", func() {
				r := http.Request{RequestURI: "/api/ping?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{"^/ping$", "^/pong$"}, []string{"^/pang$"}, []string{"/pung"}, []string{})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldBeNil)
			})

			Convey("An allow action matches the URL path, and the deny actions do not - permissions in second element", func() {
				r := http.Request{RequestURI: "/api/ping?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{"^/pang$"}, []string{"^/ping$", "^/pong$"}, []string{}, []string{"/pung"})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldBeNil)
			})
		})

		Convey("returns an error when", func() {

			Convey("An allow action matches the URL path, and a deny action also matches - first element", func() {
				r := http.Request{RequestURI: "/api/pong?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{"^/ping$", "^/pong$"}, []string{}, []string{"^/pong$"}, []string{})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldNotBeNil)
			})

			Convey("An allow action matches the URL path, and a deny action also matches - second element", func() {
				r := http.Request{RequestURI: "/api/pong?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{"^/ping$", "^/pong$"}, []string{}, []string{"^/pong$"}, []string{})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldNotBeNil)
			})

			Convey("No allow action and no deny action matches the URL path", func() {
				r := http.Request{RequestURI: "/api/pang?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{"x"}, []string{"y"}, []string{"1"}, []string{"2"})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldNotBeNil)
			})

			Convey("No allow action and no deny action matches the URL path - allows and denys are empty", func() {
				r := http.Request{RequestURI: "/api/pang?foo=bar"}
				u := auth.User{}
				createUser(&u, []string{}, []string{}, []string{}, []string{})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldNotBeNil)
			})

			Convey("The URL path cannot be extracted", func() {
				r := http.Request{RequestURI: ""}
				u := auth.User{}
				createUser(&u, []string{"^/ping$", "^/pong$"}, []string{}, []string{"/pung"}, []string{})
				err := auth.PolicyAuthorizationStrategy(apiPrefix)(u, &r)
				So(err, ShouldNotBeNil)
			})

		})

	})

	Convey("AllowAllAuthorizationStrategy", t, func() {
		Convey("never returns an error", func() {
			r := http.Request{RequestURI: "/api/ping?foo=bar"}
			u := auth.User{}
			createUser(&u, []string{}, []string{}, []string{"ping"}, []string{".*"})
			err := auth.AllowAllAuthorizationStrategy(u, &r)
			So(err, ShouldBeNil)
		})
	})

	Convey("RequireAuthentication", t, func() {
		Convey("errors if the header is malformed", func() {
			r := http.Request{}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, testGoodUserIdentityDataFetcher(true), testGoodUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "Malformed authorization header or token")
		})

		Convey("errors if the user identity cannot be formed", func() {
			header := make(map[string][]string)
			header["Authorization"] = []string{"Bearer some-token"}
			r := http.Request{
				Header: header,
			}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, testBadUserIdentityDataFetcher, testGoodUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "UserID error")
		})

		Convey("errors if the user object cannot be formed", func() {
			header := make(map[string][]string)
			header["Authorization"] = []string{"Bearer some-token"}
			r := http.Request{
				Header: header,
			}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, testGoodUserIdentityDataFetcher(true), testBadUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "UserID error")
		})

		Convey("errors if the email address is not verified", func() {
			header := make(map[string][]string)
			header["Authorization"] = []string{"Bearer some-token"}
			r := http.Request{
				Header: header,
			}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, testGoodUserIdentityDataFetcher(false), testGoodUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "Unauthorized - email not verified")
		})

		Convey("errors if the authorization strategy returns an error", func() {
			header := make(map[string][]string)
			header["Authorization"] = []string{"Bearer some-token"}
			r := http.Request{
				Header: header,
			}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(testAlwaysDenyAuthorizationStrategy, testGoodUserIdentityDataFetcher(true), testGoodUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "Unauthorized - denied by policy")
		})

		Convey("does not error if the authorization strategy returns an error", func() {
			header := make(map[string][]string)
			header["Authorization"] = []string{"Bearer some-token"}
			r := http.Request{
				Header:     header,
				RequestURI: "/api/ping?foo=bar",
			}
			w := MockResponseWriter{}
			fn := auth.RequireAuthentication(auth.AllowAllAuthorizationStrategy, testGoodUserIdentityDataFetcher(true), testGoodUserDataFetcher)
			mockNext := MockNext{}
			fn(mockNext).ServeHTTP(w, &r)
			So(errorOutputForTesting, ShouldEqual, "pass")
		})
	})

}

type MockRequest struct {
	Written string
	Request http.Request
}

type MockNext struct {
}

func (m MockNext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errorOutputForTesting = "pass"
}

type MockResponseWriter struct {
}

func (m MockResponseWriter) WriteHeader(statusCode int) {
}

func (m MockResponseWriter) Header() http.Header {
	return nil
}

func (m MockResponseWriter) Write(bs []byte) (int, error) {
	errorOutputForTesting = string(bs)
	return 0, nil
}
