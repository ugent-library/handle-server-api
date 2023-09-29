package cli

type Config struct {
	Version struct {
		Branch string `env:"SOURCE_BRANCH"`
		Commit string `env:"SOURCE_COMMIT"`
		Image  string `env:"IMAGE_NAME"`
	}
	Production   bool   `env:"HDL_PRODUCTION" envDefault:"false"`
	Host         string `env:"HDL_HOST,notEmpty" envDefault:"0.0.0.0"`
	Port         int    `env:"HDL_PORT,notEmpty" envDefault:"3000"`
	DSN          string `env:"HDL_DSN,notEmpty" envDefault:"postgres://handle:handle@localhost:5432/handle?sslmode=disable"`
	AuthUsername string `env:"HDL_AUTH_USERNAME,notEmpty" envDefault:"handle"`
	AuthPassword string `env:"HDL_AUTH_PASSWORD,notEmpty" envDefault:"handle"`
	Prefix       string `env:"HDL_PREFIX,notEmpty" envDefault:"1854"`
}
