package config

// Target is one [targets.<alias>] entry: where to fire and which env var holds
// its secret. Provider is an optional default provider for the alias.
type Target struct {
	URL       string `toml:"url"`
	Provider  string `toml:"provider"`
	SecretEnv string `toml:"secret_env"`
}
