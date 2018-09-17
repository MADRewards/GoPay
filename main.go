package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// BaseAPIURLGVSC ...
const BaseAPIURLGVSC = "http://devapi.globalvipservicescorporation.com/admin/"

// BaseAPIURLDubli ...
const BaseAPIURLDubli = "https://adminapi.ominto.com/"
const purchaseAmount = 1000

// Transaction is the transaction rolling counter object
type Transaction struct {
	ID           int64  `json:"id"`
	AppID        string `json:"app_id"`
	TotalAmount  string `json:"total_amount"`
	Currency     string `json:"currency"`
	RollingCount int64  `json:"rolling_count"`
	Status       string `json:"status"`
	Created      string `json:"created"`
	LastUpdated  string `json:"last_updated"`
}

// RollingCount ...
type RollingCount struct {
	AppID string `json:"app_id"`
	Total int64  `json:"total"`
}

// UpdateRollingCountModel ...
type UpdateRollingCountModel struct {
	ID           int64  `json:"id"`
	RollingCount int64  `json:"rolling_count"`
	Status       string `json:"status"`
}

// CardPay ...
type CardPay struct {
	ID           int64  `json:"id"`
	AppID        string `json:"app_id"`
	RollingCount int64  `json:"rolling_count"`
}

func main() {
	// schedule := dcron.Schedule{Month: dcron.ANY, Day: dcron.ANY, Weekday: dcron.ANY, Hour: dcron.ANY, Minute: dcron.ANY, Second: 3}

	// jobConfig := dcron.JobConfig{RetriesThreshold: 5, RetryUnit: "seconds", RetryInterval: 4}
	// j1 := dcron.Job{Task: mytask, Schedule: schedule, Config: jobConfig}
	// j2 := dcron.Job{Task: mytask, Schedule: schedule, Config: jobConfig}

	// jobs := []dcron.Job{j1, j2}

	// dcron.ScheduleJobs(jobs)

	// dcron.Start()

	mytask()

}

func mytask() {
	res, err := http.Get(ConcatenateGvscURL("getConfirmed"))
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	t, err := getTransactions([]byte(body))
	if err != nil {
		panic(err.Error())
	}

	var list []string
	for _, transaction := range t {
		list = append(list, transaction.AppID)
	}

	// Convert string slice to string.
	// ... Has comma in between strings.
	appIDs := strings.Join(list, ",")

	r, err := getRollingCountsByAppIDs(appIDs)
	if err != nil {
		panic(err)
	}
	var rollingCountMap map[string]int64
	rollingCountMap = make(map[string]int64)

	for _, s := range r {
		rollingCountMap[s.AppID] = s.Total
	}

	updateList := []UpdateRollingCountModel{}
	cardPayList := []CardPay{}
	for _, transaction := range t {
		f, err := strconv.ParseFloat(transaction.TotalAmount, 64)
		if err != nil {
			panic(err.Error())
		}
		totalCount := (f / purchaseAmount) - float64(rollingCountMap[transaction.AppID])
		//fmt.Println(transaction.ID, f, rollingCountMap[transaction.AppID], int64(totalCount))
		if totalCount >= 1 {
			updateList = append(updateList, UpdateRollingCountModel{ID: transaction.ID, RollingCount: int64(totalCount), Status: "paid"})
			cardPayList = append(cardPayList, CardPay{ID: transaction.ID, AppID: transaction.AppID, RollingCount: int64(totalCount)})
			//fmt.Println("success", int64(totalCount))
		} else {
			updateList = append(updateList, UpdateRollingCountModel{ID: transaction.ID, RollingCount: transaction.RollingCount, Status: "paid"})
		}
	}

	// for _, u := range updateList {
	// 	fmt.Println(u.ID, u.RollingCount)
	// }

	c, err := json.Marshal(cardPayList)
	if err != nil {
		panic(err.Error())
	}
	p, err := payDubli(c)
	if err != nil {
		panic(err.Error())
	} else {
		if p == 200 {
			u, err := json.Marshal(updateList)
			if err != nil {
				panic(err.Error())
			}
			updateRollingCount(u)
		}
	}
	fmt.Println(p)

}

func payDubli(cardPayList []byte) (int, error) {
	form := cardPayList
	body := bytes.NewBuffer(form)
	fmt.Println(body)
	res, err := http.Post(ConcatenateDubliURL("cardPay"), "application/json", body)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return res.StatusCode, nil
}

func updateRollingCount(updateList []byte) {
	form := updateList
	body := bytes.NewBuffer(form)
	res, err := http.Post(ConcatenateGvscURL("updateTxnRollingCounters"), "application/json", body)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func getRollingCountsByAppIDs(ids string) ([]RollingCount, error) {
	res, err := http.Get(fmt.Sprintf("%s%s", ConcatenateGvscURL("getRollingCounterByAppId/"), ids))

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	r, err := getRollingCounts([]byte(body))
	if err != nil {
		return nil, err
	}
	return r, nil
}

func getTransactions(body []byte) ([]Transaction, error) {
	var t []Transaction
	err := json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func getRollingCounts(body []byte) ([]RollingCount, error) {
	var r []RollingCount
	err := json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ConcatenateGvscURL ...
func ConcatenateGvscURL(endpoint string) string {
	return fmt.Sprintf("%s%s", BaseAPIURLGVSC, endpoint)
}

// ConcatenateDubliURL ...
func ConcatenateDubliURL(endpoint string) string {
	return fmt.Sprintf("%s%s", BaseAPIURLDubli, endpoint)
}
