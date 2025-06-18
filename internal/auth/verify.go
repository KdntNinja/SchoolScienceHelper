package auth

import (
	"net/http"
)

// IsEmailVerifiedFromRequest returns true if the user's email is verified, false otherwise.
func IsEmailVerifiedFromRequest(r *http.Request) (bool, error) {
	claims, err := ValidateAndParseJWT(GetJWTFromRequest(r))
	if err != nil {
		return false, err
	}
	verified, _ := claims["email_verified"].(bool)
	return verified, nil
}
