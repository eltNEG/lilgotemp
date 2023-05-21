package schema

type Environment string

var Environments = struct {
	Production  Environment
	Development Environment
}{
	Production:  "production",
	Development: "development",
}
