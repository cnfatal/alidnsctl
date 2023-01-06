package alidnsctl

var gitVersion = "v0.0.0" // taged version $(git describe --tags --dirty)

func Version() string {
	return gitVersion
}
