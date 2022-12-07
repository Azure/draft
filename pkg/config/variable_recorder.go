package config

type VariableRecorder interface {
	Record(key, value string)
}
