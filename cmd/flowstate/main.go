package main

import (
	"errors"
	"fmt"

	"github.com/vmarlier/flowstate/internal/config"
)

func main() {
	loadConfig()
}

func loadConfig() {
	cfg, err := config.Load("./configs/bad_config.json")
	if err != nil {
		var agg *config.AggregatedConfigErrors

		if errors.As(err, &agg) {
			fmt.Println("ERROR: " + err.Error() + ":")

			for _, cfgErr := range agg.ConfigErrors {
				if cfgErr.Cause != nil {
					fmt.Printf("- field=%s kind=%s message=%s cause=%v\n",
						cfgErr.FieldPath,
						cfgErr.Kind,
						cfgErr.Message,
						cfgErr.Cause,
					)
				} else {
					fmt.Printf("- field=%s kind=%s message=%s\n",
						cfgErr.FieldPath,
						cfgErr.Kind,
						cfgErr.Message,
					)
				}
			}

			return
		}

		fmt.Println("load failed:", err)
		return
	}

	_ = cfg
}
