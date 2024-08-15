package me.rego

default ruleAny = false
default ruleAdminOnly = false
default ruleUserOnly = false
default ruleAdminOrSubject = false

ruleUser := "USER"
ruleAdmin := "ADMIN"
roleAll := {ruleUser, ruleAdmin}

// ruleAny is true provided that all the rules inside following bracket are true
ruleAny {
	claim_roles := {role | role := input.roles[_]}
	input_roles := roleAll & claim_roles
	count(input_roles) > 0
}

ruleAdminOnly {
	claim_roles := {role | role := input.roles[_]}
	input_admin := {roleAdmin} & claim_roles
	count(input_admin) > 0
}

ruleUserOnly {
	claim_roles := {role | role := input.roles[_]}
	input_user := {roleUser} & claim_roles
	count(input_user) > 0
}

ruleAdminOrSubject {
	claim_roles := {role | role := input.roles[_]}
	input_admin := {roleAdmin} & claim_roles
	count(input_admin) > 0
} else {
	claim_roles := {role | role := input.roles[_]}
	input_user := {roleUser} & claim_roles
	count(input_user) > 0
	input.UserID == input.Subject
}

// else part could have been written like this(using OR rules)
// ruleAdminOrSubject {
// 	claim_roles := {role | role := input.roles[_]}
// 	input_user := {roleUser} & claim_roles
// 	count(input_user) > 0
// 	input.UserID == input.Subject
// }