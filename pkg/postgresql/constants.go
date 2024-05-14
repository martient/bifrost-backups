package postgresql

type PostgresqlRequirements struct {
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`
}
