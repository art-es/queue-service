package initial

import (
	"fmt"
	"os"
	"strconv"
)

type Env struct {
	Name     string
	Target   any
	Required bool
}

func ParseEnv(vars ...Env) error {
	for _, v := range vars {
		strVal := os.Getenv(v.Name)

		if v.Required && len(strVal) == 0 {
			return fmt.Errorf("required env: %s", v.Name)
		}

		switch v.Target.(type) {
		case *string:
			*v.Target.(*string) = strVal

		case *int:
			intVal, err := strconv.Atoi(strVal)
			if err != nil {
				return fmt.Errorf("env %q convert string (%s) to int: %w", v.Name, strVal, err)
			}
			*v.Target.(*int) = intVal

		default:
			return fmt.Errorf("unknown type of target: %s", v.Name)
		}
	}

	return nil
}
