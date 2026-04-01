package formatter

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	IndentSize int
	UseTabs    bool
	MaxWidth   int
}

func DefaultConfig() Config {
	return Config{
		IndentSize: 4,
		UseTabs:    false,
		MaxWidth:   100,
	}
}

func LoadConfig(dir string) Config {
	cfg := DefaultConfig()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return cfg
	}

	for {
		path := filepath.Join(absDir, ".ccolonfmt")
		if data, err := os.Open(path); err == nil {
			scanner := bufio.NewScanner(data)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				switch key {
				case "indent_size":
					if n, err := strconv.Atoi(val); err == nil && n > 0 {
						cfg.IndentSize = n
					}
				case "use_tabs":
					cfg.UseTabs = val == "true"
				case "max_width":
					if n, err := strconv.Atoi(val); err == nil && n > 0 {
						cfg.MaxWidth = n
					}
				}
			}
			data.Close()
			return cfg
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		absDir = parent
	}

	return cfg
}
