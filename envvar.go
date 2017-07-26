package pvc

// Default mapping for this backend
const (
	DefaultEnvVarMapping = "SECRET_{{ .ID }}" // DefaultEnvVarMapping is uppercased after interpolation for convenience
)
