package main

import (
	"app/cipher"
	"fmt"
	"log"
	"os"
	"strings"

	"app/services"
)

func main() {

	services.LoadEnv(`.env`)
	if len(os.Args) < 2 {
		log.Fatal(`please command the relative path of CSV file `)
	}
	rd, err := cipher.NewRot128Reader(os.Args[1])
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rd.Close()

	worker := services.NewWorker()

	worker.WaitGroup.Add(worker.NumWorker)
	for i := 0; i < worker.NumWorker; i++ {
		go worker.Run(i)
	}

	rd.Scan() //skip column line
	for {

		line, ok := rd.Scan()
		if !ok {
			break
		}
		row := strings.Split(line, ",")
		worker.ChannelWork <- &row
	}

	worker.Close()
	worker.WaitGroup.Wait()
	fmt.Println(worker.Result.Response())
}
