// Package main provides ...
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"encoding/csv"

	"github.com/kniren/gota/dataframe"
	"golang.org/x/sync/semaphore"
	"gonum.org/v1/gonum/stat/distuv"
)

type dist interface {
	CDF(x float64) float64
}

func readCSV(name string) dataframe.DataFrame {
	// insert the file name
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file_csv := csv.NewReader(file) // declare a new reader to read the csv file

	series := make([][]string, 0)
	// read each line until the end of file is reach
	for {
		record, err := file_csv.Read() // read the record of each line
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		series = append(series, record) // store the record into the list of series
	}

	return dataframe.LoadRecords(series) // turn the list of series into the dataframe and then return
}

func giaTriThucNghiem(length int, data dataframe.DataFrame, distribution dist) []float64 {
	type prob struct {
		data_probability []float64
		sync.Mutex
	}

	var (
		maxWorker = int64(runtime.GOMAXPROCS(0))
		sem       = semaphore.NewWeighted(maxWorker)
		ctx       = context.TODO()
		groupTest = new(sync.WaitGroup)

		data_value = make([]float64, length)
		data_range = make([]string, length)

		result           = new(prob)
		data_probability = make([]float64, length)
	)
	result.data_probability = data_probability

	for i := 0; i < length; i++ {
		data_value[i] = data.Elem(i, 1).Float()
		data_range[i] = data.Elem(i, 0).String()
	}

	groupTest.Add(len(data_range)) // adding semaphore for the for loop
	for index, value := range data_range {
		// worker wait group
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Fatal("Failed to acquire semaphore: %w", err)
			break
		}

		go func(index int, value string) {
			defer sem.Release(1) // release the worker after done
			if len(strings.Split(value, "-")) == 1 && len(strings.Split(value, "<")) == 1 && len(strings.Split(value, ">")) == 1 {
				// When the range is an actual number not range
				// example: 1, 2, 3, 4
				number, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatal(err)
				}

				probability := distribution.CDF(number) // calculate the CDF of the range

				// Use mutex lock so no data race
				result.Lock()
				result.data_probability[index] = probability
				result.Unlock()

				// signal the semaphore
				groupTest.Done()

			} else if len(strings.Split(value, "-")) == 2 {
				// Range between 2 number
				// example: 100 - 200, 312873091 - 487284432423
				str_arr := strings.Split(value, "-")

				first_range, err := strconv.ParseFloat(str_arr[0], 64) // convert the first range into float64
				if err != nil {
					log.Fatal(err)
				}

				second_range, err := strconv.ParseFloat(str_arr[1], 64) // convert the second range into float64
				if err != nil {
					log.Fatal(err)
				}

				probability := distribution.CDF(second_range) - distribution.CDF(first_range) // calculate the p

				// Use mutex lock so no data race
				result.Lock()
				result.data_probability[index] = probability
				result.Unlock()

				// signal the semaphore
				groupTest.Done()

			} else {
				// Range between inifinity and the number
				// example: >100, <200, >1000, >3218, <3721893
				var probability float64
				range_of_smaller_than := strings.Split(value, "<")

				// if when you split the string but the len is only 1 then the range is >
				if len(range_of_smaller_than) == 1 {
					// calculate p when range is >
					// example: >38219038210, >3821938
					range_of_smaller_than = strings.Split(range_of_smaller_than[0], ">")
					range_number, err := strconv.ParseFloat(range_of_smaller_than[1], 64) // convert number into float64
					if err != nil {
						log.Fatal(err)
					}

					probability = 1 - distribution.CDF(range_number) // calculate the p when the range is <

				} else {
					// calculate p when range is <
					// example: <382103921, <80821093
					range_number, err := strconv.ParseFloat(range_of_smaller_than[1], 64) // convert number into float64
					if err != nil {
						log.Fatal(err)
					}

					probability = distribution.CDF(range_number) // calculate the p when the range is >
				}

				// Use mutex lock so no data race
				result.Lock()
				result.data_probability[index] = probability
				result.Unlock()

				//signal the semaphore
				groupTest.Done()
			}
		}(index, value)
	}
	groupTest.Wait() // wait until all the go routine terminate

	return data_probability
}

func K_thuc_nghiem(data_probability []float64, data dataframe.DataFrame) float64 {
	var K float64
	var sum float64

	for i := 0; i < data.Nrow(); i++ {
		sum += data.Elem(i, 1).Float()
	}

	for index, value := range data_probability {
		K += math.Pow((data.Elem(index, 1).Float()-value*sum), 2) / (sum * value)
	}

	return K
}

func main() {
	var file_name string

	testDistribution := new(distuv.Normal)
	testDistribution.Mu = 34.93
	testDistribution.Sigma = 21.82

	file_name = "new.csv"
	df := readCSV(file_name)

	result := giaTriThucNghiem(df.Nrow(), df, testDistribution)
	result_2 := K_thuc_nghiem(result, df)

	fmt.Println(result)
	fmt.Println(result_2)
}
