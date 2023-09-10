package models

type Config struct {
	Debug    bool   `yaml:"debug" env:"DEBUG"`
	User     string `yaml:"user" env:"USER" env-default:"user"`
	Dir      string `yaml:"dir" env:"DIR" env-default:"."`
	Token    string `yaml:"token" env:"TOKEN"`
	Compress bool   `yaml:"compress" env:"COMPRESS"`
}
