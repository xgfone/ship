package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xgfone/ship"
)

// TokenAuth returns a TokenAuth middleware.
//
// For valid key it will calls the next handler.
// For invalid key, it responds "401 Unauthorized".
// For missing key, it responds "400 Bad Request".
//
// If getAuthTokenFromRequest is missing, the default is
// GetAuthTokenFromHeader(ship.HeaderAuthorization, "Bearer").
func TokenAuth(validateToken func(token string) (bool, error),
	getAuthTokenFromRequest ...func(ctx ship.Context) (token string, err error)) ship.Middleware {
	getAuthToken := GetAuthTokenFromHeader(ship.HeaderAuthorization, "Bearer")
	if len(getAuthTokenFromRequest) > 0 && getAuthTokenFromRequest[0] != nil {
		getAuthToken = getAuthTokenFromRequest[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			token, err := getAuthToken(ctx)
			if err != nil {
				return ship.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			if valid, err := validateToken(token); err != nil {
				return err
			} else if valid {
				return next(ctx)
			}
			return ship.ErrUnauthorized
		}
	}
}

// GetAuthTokenFromHeader is used to get the CSRF token from the request header.
func GetAuthTokenFromHeader(header string, authScheme ...string) func(ship.Context) (string, error) {
	var scheme string
	if header == ship.HeaderAuthorization {
		if len(authScheme) > 0 {
			scheme = strings.TrimSpace(authScheme[0])
		}
		if scheme == "" {
			panic(errors.New("ship: TokenAuth requires authScheme for Authorization"))
		}
	}
	schemelen := len(scheme)

	return func(ctx ship.Context) (string, error) {
		auth := ctx.Request().Header.Get(header)
		if auth == "" {
			return "", errors.New("missing auth token in request header")
		}
		if schemelen > 0 {
			if len(auth) > schemelen+1 && auth[:schemelen] == scheme {
				return auth[schemelen+1:], nil
			}
			return "", errors.New("invalid auth token in the request header")
		}
		return ctx.Request().Header.Get(header), nil
	}
}

// GetAuthTokenFromForm is used to get the CSRF token from the request body FORM.
func GetAuthTokenFromForm(param string) func(ship.Context) (string, error) {
	return func(ctx ship.Context) (string, error) {
		token := ctx.FormValue(param)
		if token != "" {
			return token, nil
		}
		return "", errors.New("missing auth token in the form parameter")
	}
}

// GetAuthTokenFromQuery is used to get the CSRF token from the request URL query.
func GetAuthTokenFromQuery(param string) func(ship.Context) (string, error) {
	return func(ctx ship.Context) (string, error) {
		token := ctx.QueryParam(param)
		if token != "" {
			return token, nil
		}
		return "", errors.New("missing auth token in the url query")
	}
}
