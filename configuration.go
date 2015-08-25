package base

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Env     string
	Secrets string
	Http    struct {
		Host string
		Port int
	}
	Database struct {
		Adapter  string
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		Extra    string
	}
	Redis struct {
		Host string
		Port int
		Db   int
	}
}

func (p *Configuration) IsProduction() bool {
	return p.Env == "production"
}

func (p *Configuration) DbCreate() (string, []string) {
	d := p.Database.Adapter
	switch d {
	case "postgres":
		return "psql", []string{
			"-h", p.Database.Host,
			"-p", strconv.Itoa(p.Database.Port),
			"-U", p.Database.User,
			"-c", fmt.Sprintf("CREATE DATABASE %s", p.Database.Name)}
	default:
		return "echo", []string{"Unknown database driver " + d}
	}
}

func (p *Configuration) DbDrop() (string, []string) {
	d := p.Database.Adapter
	switch d {
	case "postgres":
		return "psql", []string{
			"-h", p.Database.Host,
			"-p", strconv.Itoa(p.Database.Port),
			"-U", p.Database.User,
			"-c", fmt.Sprintf("DROP DATABASE %s", p.Database.Name)}
	default:
		return "echo", []string{"Unknown database driver " + d}
	}
}

func (p *Configuration) DbShell() (string, []string) {
	d := p.Database.Adapter
	switch d {
	case "postgres":
		return "psql", []string{
			"-h", p.Database.Host,
			"-p", strconv.Itoa(p.Database.Port),
			"-d", p.Database.Name,
			"-U", p.Database.User}
	default:
		return "echo", []string{"Unknown database driver " + d}
	}
}

func (p *Configuration) DbUrl() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?%s",
		p.Database.Adapter, p.Database.User, p.Database.Password, p.Database.Host,
		p.Database.Port, p.Database.Name, p.Database.Extra)
}

func (p *Configuration) RedisShell() (string, []string) {
	//todo select db
	return "telnet", []string{p.Redis.Host, strconv.Itoa(p.Redis.Port)}
}

func (p *Configuration) RedisUrl() string {
	return fmt.Sprintf("%s:%d", p.Redis.Host, p.Redis.Port)
}

//-----------------------------------------------------------------------------
func Load(file string) (*Configuration, error) {
	_, err := os.Stat(file)
	if err == nil {
		config := Configuration{}
		log.Printf("Load from config file: %s", file)

		var tmp *template.Template
		tmp, err = template.ParseFiles(file)
		if err != nil {
			return nil, err
		}

		vars := make(map[string]interface{}, 0)

		vars["Env"] = os.Getenv("ITPKG_ENV")
		vars["Secrets"] = os.Getenv("ITPKG_SECRETS")
		vars["DbPassword"] = os.Getenv("ITPKG_DATABASE_PASSWORD")

		var buf bytes.Buffer

		if err = tmp.Execute(&buf, vars); err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(buf.Bytes(), &config); err != nil {
			return nil, err
		}
		return &config, nil
	}
	return nil, err
}
