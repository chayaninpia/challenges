package services

import (
	"fmt"
	"sync"

	"github.com/joho/godotenv"
)

var isLoadFile = map[string]bool{}
var lock = &sync.Mutex{}

func LoadEnv(file string) error {

	if _, ok := isLoadFile[file]; !ok {
		// load .env file
		err := godotenv.Load(file)
		if err != nil {
			return fmt.Errorf(`loading environment file [%v], %v`, file, err)
		}
		lock.Lock()
		defer lock.Unlock()
		isLoadFile[file] = true
	}

	return nil
}
