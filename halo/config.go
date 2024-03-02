package halo

type Config struct {
	Genesis   string
	UrlPrefix string `toml:"url_prefix"`
	SumbitUrl string `toml:"submit_url"`
}
