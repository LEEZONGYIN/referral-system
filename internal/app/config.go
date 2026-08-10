package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
}

type ServerConfig struct {
	Addr string
}

type MySQLConfig struct {
	DSN string
}

func LoadConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	parsed := parseYAMLConfig(string(content))
	cfg := Config{
		Server: ServerConfig{Addr: ":8080"},
	}
	if v := parsed["server.addr"]; v != "" {
		cfg.Server.Addr = v
	}
	if v := parsed["mysql.dsn"]; v != "" {
		cfg.MySQL.DSN = v
	}
	if cfg.MySQL.DSN == "" {
		return Config{}, fmt.Errorf("mysql.dsn is required")
	}
	return cfg, nil
}

func parseYAMLConfig(content string) map[string]string {
	result := make(map[string]string)
	var section string
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"`)
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = strings.Trim(value, "'")
		}
		if section != "" {
			key = section + "." + key
		}
		if parsed, err := strconv.Unquote("\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""); err == nil {
			value = parsed
		}
		result[key] = value
	}
	return result
}
