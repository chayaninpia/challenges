package services

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	omise "github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

type Worker struct {
	WaitGroup   sync.WaitGroup
	PublicKey   string
	SecretKey   string
	NumWorker   int
	ChannelWork chan *[]string
	Result
	PositonsCol
}

type PositonsCol struct {
	Name   int
	Amount int
	Year   int
	Month  int
	Card   int
	Ccv    int
}

func NewWorker() *Worker {
	return &Worker{
		WaitGroup:   sync.WaitGroup{},
		PublicKey:   os.Getenv(`OMISE_PUBLIC_KEY`),
		SecretKey:   os.Getenv(`OMISE_SECRET_KEY`),
		NumWorker:   15,
		ChannelWork: make(chan *[]string),
		PositonsCol: PositonsCol{
			Amount: 1,
			Name:   0,
			Card:   2,
			Ccv:    3,
			Month:  4,
			Year:   5,
		},
		Result: Result{
			NumSuccess:  0,
			TotalAmount: 0,
			TotalFaulty: 0,
			Donator:     map[string]float64{},
			TopDonator:  []TopDonator{},
		},
	}
}

func (c *Worker) Close() {
	close(c.ChannelWork)
}

func (c *Worker) Run(number int) {

	client, err := omise.NewClient(c.PublicKey, c.SecretKey)
	defer c.WaitGroup.Done()
	if err != nil {
		log.Fatalln(`Connection to Omise API failed`)
		return
	}
	for {
		row, ok := <-c.ChannelWork
		if !ok {
			break
		}
		for {
			if retry := c.charge(row, client); !retry {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func (c *Worker) charge(rowPtr *[]string, client *omise.Client) bool {

	row := *rowPtr
	log.Println(row)
	amount, _ := strconv.ParseInt(row[c.PositonsCol.Amount], 10, 0)
	year, _ := strconv.Atoi(row[c.PositonsCol.Year])
	month, _ := strconv.Atoi(row[c.PositonsCol.Month])

	token, createToken := &omise.Token{}, &operations.CreateToken{
		Name:            row[c.PositonsCol.Name],
		Number:          row[c.PositonsCol.Card],
		SecurityCode:    row[c.PositonsCol.Ccv],
		ExpirationMonth: time.Month(month),
		ExpirationYear:  year,
	}
	if err := client.Do(token, createToken); err != nil {
		retry := true
		log.Printf(`create token : %v`, err)

		if !checkToManyReq(err) {
			retry = false
			c.Result.TranFailed(amount)
		}
		return retry
	}

	// Creates a charge from the token
	charge, createCharge := &omise.Charge{}, &operations.CreateCharge{
		Amount:   amount,
		Currency: os.Getenv("CURRENCY"),
		Card:     token.ID,
	}
	if err := client.Do(charge, createCharge); err != nil {
		retry := true
		log.Printf(`create charge : %v`, err)
		if !checkToManyReq(err) {
			retry = false
			c.Result.TranFailed(amount)
		}
		return retry
	}
	c.Result.TranSuccess(amount, row[c.PositonsCol.Name])
	return false
}

func checkToManyReq(err error) bool {

	return strings.Contains(err.Error(), "too_many_requests")
}
