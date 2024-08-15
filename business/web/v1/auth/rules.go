package auth

import (
	_ "embed"
)

const (
	RuleAuthenticate   = "auth"
	RuleAny            = "ruleAny"
	RuleAdminOnly      = "ruleAdminOnly"
	RuleUserOnly       = "ruleUserOnly"
	RuleAdminOrSubject = "ruleAdminOrSubject"
)

const (
	opaPackage string = "me.rego"
)

var (
	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)
