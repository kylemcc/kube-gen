package kubegen

type Config struct {
	Host          string
	Template      string
	Output        string
	Overwrite     bool
	Watch         bool
	NotifyCmd     string
	ResourceTypes []string
}

type Generator interface {
	Generate() error
}

type generator struct {
}

func (g *generator) Generate() error {
	return nil
}

func NewGenerator(c Config) (Generator, error) {
	return &generator{}, nil
}
