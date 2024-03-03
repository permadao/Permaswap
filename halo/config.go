package halo

type Config struct {
	Genesis           string
	UrlPrefix         string `toml:"url_prefix"`
	DefaulHaloNodeUrl string `toml:"default_halo_node_url"`
}
