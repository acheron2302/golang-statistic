package utils

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sync"

	"encoding/csv"

	"github.com/kniren/gota/dataframe"
	"gonum.org/v1/gonum/stat/distuv"
)

func readCSV(name string) dataframe.DataFrame {
	// insert the file name
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file_csv := csv.NewReader(file)

	series := make([][]string, 0)
	for {
		record, err := file_csv.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		series = append(series, record)
	}

	return dataframe.LoadRecords(series)
}

func preCalculate(row, column int, data dataframe.DataFrame) float64 {
	// variable for the mutex and the semaphore
	groupTest := new(sync.WaitGroup)
	type final_sum struct {
		sum float64
		sync.Mutex
	}

	var sum float64          // sum of all element
	result := new(final_sum) // create a mutex lock for the K_real so no data race when calculate
	result.sum = 0.0         // set the final result to 0

	sum_row := make([]float64, row)
	sum_column := make([]float64, column)

	// calculate the sum of all the element in the data
	for i := 0; i < row; i++ {
		for x := 0; x < column; x++ {
			sum += data.Elem(i, x).Float()
		}
	}

	// Calculate the sum of row and sum of column
	for i := 0; i < row; i++ {
		for x := 0; x < column; x++ {
			sum_row[i] += data.Elem(i, x).Float()
		}
	}
	for i := 0; i < column; i++ {
		for x := 0; x < row; x++ {
			sum_column[i] += data.Elem(x, i).Float()
		}
	}

	// Calculate the final result with goroutine and mutex lock
	groupTest.Add(column * row)
	// loop each element in the data
	for i := 0; i < column; i++ {
		for x := 0; x < row; x++ {
			// create a new goroutine for each element to calculate the K_real
			go func(x int, i int) {
				local_sum := sum * math.Pow(data.Elem(x, i).Float()-sum_column[i]*sum_row[x]/sum, 2) / (sum_column[i] * sum_row[x])

				// Make sure that each goroutine can only use the result variable one at the time
				result.Lock()
				result.sum += local_sum
				result.Unlock()

				// signal back to main function
				groupTest.Done()
			}(x, i)
		}
	}
	groupTest.Wait()

	return result.sum
}

func main() {
	var file_name string
	fmt.Printf("Insert the data path: ")
	fmt.Scanln(&file_name)

	df := readCSV(file_name)
	result := preCalculate(df.Nrow(), df.Ncol(), df)

	reject := new(distuv.ChiSquared)
	reject.K = float64((df.Nrow() - 1) * (df.Ncol() - 1))

	if result > reject.Quantile(0.95) {
		fmt.Println("Doc lap")
	} else {
		fmt.Println("Khong doc lap")
	}

	fmt.Println(result)
	fmt.Println(reject.Quantile(0.95))
}
