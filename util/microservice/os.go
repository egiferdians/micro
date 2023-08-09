package microservice

import (
	"fmt"
	"os"
)

//get system env variable
func GetOsEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		fmt.Println("INFO", "The env "+name+" not set")
	}
	return value
}