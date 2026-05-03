package main

import (
	"fmt"

	"github.com/vmarlier/flowstate/internal/config"
)

func main() {
	fmt.Println(config.Load("./configs/config_validate_path_prefix.json"))
}
