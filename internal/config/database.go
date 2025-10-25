package config

type MySqlConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int64  `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	Loc      string `json:"loc" yaml:"loc"`
	Charset  string `json:"charset" yaml:"charset"`
}
