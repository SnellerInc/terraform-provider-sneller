package api

type MfaRequirement string

const (
	MfaOff      = MfaRequirement("off")
	MfaOptional = MfaRequirement("optional")
	MfaRequired = MfaRequirement("required")
)
