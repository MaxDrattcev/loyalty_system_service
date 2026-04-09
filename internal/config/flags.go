package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Flags struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func ParseFlags() (*Flags, error) {
	var (
		runAddress           = flag.String("a", "", "адрес и порт сервера")
		databaseURI          = flag.String("d", "", "DSN базы данных PostgresSQL")
		accrualSystemAddress = flag.String("r", "", "адрес и порт сервера системы начислений")
	)
	flag.Parse()

	if err := checkUnknownFlags(); err != nil {
		return nil, err
	}
	return &Flags{
		RunAddress:           *runAddress,
		DatabaseURI:          *databaseURI,
		AccrualSystemAddress: *accrualSystemAddress,
	}, nil
}

func checkUnknownFlags() error {
	knownFlags := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		knownFlags[f.Name] = true
	})
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		flagName := strings.TrimPrefix(arg, "-")
		flagName = strings.TrimPrefix(flagName, "-")

		if idx := strings.Index(flagName, "="); idx != -1 {
			flagName = flagName[:idx]
		}

		if !knownFlags[flagName] {
			return fmt.Errorf("unknown flag: -%s", flagName)
		}
	}

	return nil
}
