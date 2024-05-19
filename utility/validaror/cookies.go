package validator

import "strings"

// IsCookieValid determines if _abck cookie is valid, you cannot access protected site without valid cookie
//
// Parameters:
//   - abck: _abck cookie value
//
// Returns:
//   - bool: if cookie is valid returns true
func IsCookieValid(abck string) bool {
	abckSplit := strings.Split(abck, "~")
	if len(abckSplit) < 2 {
		return false
	}

	return abckSplit[1] == "0"
}

// IsCookieValid determines if _abck cookie was set by the Akamai as invalidated
//
// Parameters:
//   - abck: _abck cookie value
//
// Returns:
//   - bool: if cookie is still valid return strue
func IsCookieNoLongerValid(abck string) bool {
	abckSplit := strings.Split(abck, "~")
	if len(abckSplit) < 4 {
		return false
	}

	return abckSplit[3] != "-1"
}
