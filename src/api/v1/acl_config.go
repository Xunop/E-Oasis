package v1

var authenticationAllowlist = map[string]bool{
	"/api/v1/signup": true,
	"/api/v1/signin": true,
}

// isUnauthorizeAllowed returns whether the method is exempted from authentication.
func isUnauthorizeAllowed(fullMethodName string) bool {
	return authenticationAllowlist[fullMethodName]
}

var allowedPathOnlyForAdmin = map[string]bool{
}

// isOnlyForAdminAllowedPath returns true if the method is allowed to be called only by admin.
func isOnlyForAdminAllowedPath(methodName string) bool {
	return allowedPathOnlyForAdmin[methodName]
}
