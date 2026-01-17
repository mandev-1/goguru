package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const defaultEnvContent = `PORT=8080
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
FROM_EMAIL=noreply@camagru.local
`

func LoadEnv(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if err := os.WriteFile(filename, []byte(defaultEnvContent), 0644); err != nil {
			return err
		}
		fmt.Printf("Created %s with default values\n", filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
	return scanner.Err()
}
