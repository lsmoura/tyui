package database

import "fmt"

type Params struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	SSL bool
}

func (p Params) ConnStr() string {
	host := p.Host
	if host == "" {
		host = "localhost"
	}
	port := p.Port
	if port == 0 {
		port = 5432
	}
	user := p.User
	if user == "" {
		user = "postgres"
	}
	password := p.Password
	database := p.Database

	ssl := "disable"
	if p.SSL {
		ssl = "enable"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, database, ssl)
}
