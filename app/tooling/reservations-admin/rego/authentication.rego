package me.rego


default auth = false

# auth is true if the all rules inside are true
auth {
	jwt_valid
}

# assign jwt_valid to be `valid`
# `valid` is assigned inside the rule
jwt_valid := valid {
	[valid, header, payload] := verify_jwt
}

verify_jwt := io.jwt.decode_verify(input.Token, {
        "cert": input.Key,
	}
)
