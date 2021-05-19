package config

type Api struct {
	Api    Http   `yaml:"api"`
	Log    Log    `yaml:"log"`
	Twitch Twitch `yaml:"twitch"`
	Secret string `yaml:"secret"`
	Repo   Repos  `yaml:"repo"`
}

type Http struct {
	Port string `yaml:"port"`
	Ssl  Ssl    `yaml:"ssl"`
	Host string `yaml:"host"`
}

type Ssl struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

type Log struct {
	Level string `yaml:"level"`
}

type Twitch struct {
	ClientID      string `yaml:"client_id"`
	ClientSecret  string `yaml:"client_secret"`
	OAuthRedirect string `yaml:"oauth_redirect"`
	OidcIssuer    string `yaml:"oidc_issuer"`
}
