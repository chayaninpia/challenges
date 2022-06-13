package services

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type Result struct {
	Mutex       sync.Mutex
	NumSuccess  int64
	TotalAmount float64
	TotalFaulty float64
	TopDonator  []TopDonator
	Donator     map[string]float64
}

type TopDonator struct {
	Name   string
	Amount float64
}

func (r *Result) SortTopDonator() {

	keys := make([]string, 0, len(r.Donator))
	for key := range r.Donator {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return r.Donator[keys[i]] > r.Donator[keys[j]]
	})

	for _, k := range keys {
		if len(r.TopDonator) < 3 {
			r.TopDonator = append(r.TopDonator, TopDonator{
				Name:   k,
				Amount: r.Donator[k],
			})
		}
	}
}

func (r *Result) Response() string {
	currency := strings.ToUpper(os.Getenv(`CURRENCY`))
	total := r.TotalAmount + r.TotalFaulty
	average := 0.0
	if r.NumSuccess > 0 {

		average = float64(r.TotalAmount) / float64(r.NumSuccess)
	}
	r.SortTopDonator()
	line1 := fmt.Sprintf("total received: %s\t%.2f", currency, total)
	line2 := fmt.Sprintf("successfully donated: %s\t%.2f", currency, r.TotalAmount)
	line3 := fmt.Sprintf("faulty donation: %s\t%.2f", currency, r.TotalFaulty)
	line4 := fmt.Sprintf("average per person: %s\t%.2f", currency, average)
	line5 := `top donors:`

	if len(r.TopDonator) != 0 {
		if len(r.TopDonator) == 1 {
			line5 = fmt.Sprintf("      top donors: %10v\n", r.TopDonator[0].Name)
		}
		if len(r.TopDonator) == 2 {
			line5 = fmt.Sprintf("      top donors: %10v\n%40v\n", r.TopDonator[0].Name, r.TopDonator[1].Name)
		}
		if len(r.TopDonator) == 3 {
			line5 = fmt.Sprintf("      top donors: %10v\n%40v\n%40v\n", r.TopDonator[0].Name, r.TopDonator[1].Name, r.TopDonator[2].Name)
		}

	}
	return fmt.Sprintf("%40v\n%40v\n%40v\n\n%40v\n%40v", line1, line2, line3, line4, line5)
}

func (r *Result) TranSuccess(amount int64, name string) {
	r.Mutex.Lock()
	r.NumSuccess = r.NumSuccess + 1
	r.Donator[name] += float64(amount) * 0.01
	r.TotalAmount += float64(amount) * 0.01
	r.Mutex.Unlock()
}

func (r *Result) TranFailed(amount int64) {
	r.Mutex.Lock()
	r.TotalFaulty += float64(amount) * 0.01
	r.Mutex.Unlock()

}
