package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

func scanLine(path string) ([]string, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var groups []string

	for scanner.Scan() {
		groups = append(groups, scanner.Text())
	}

	return groups, nil
}

func kiemDinh(s float64, k int, n int, ap float64) (bool, float64) {
	var dt distuv.F
	dt.D1 = float64(k - 1)
	dt.D2 = float64(n - k)

	F := dt.Quantile(1 - ap)
	if s > F {
		return false, F
	}
	return true, F
}

func main() {
	// Read txt file
	path := "./nhom_benh.txt"

	groups, err := scanLine(path)
	if err != nil {
		log.Fatal(err)
	}
	data := make([][]float64, len(groups))
	groupLen := make([]int, 0)

	s := make([][]string, len(groups))
	for i, v := range groups {
		s[i] = strings.Split(v, ",")
		groupLen = append(groupLen, len(s[i]))

		for _, w := range s[i] {
			n, err := strconv.ParseFloat(strings.Join(strings.Fields(w), ""), 64)
			if err != nil {
				log.Fatal(err)
			}

			data[i] = append(data[i], n)

		}
	}

	// Calculate average of each group and S1 and S2
	average := make([]float64, len(groups))
	variance := make([]float64, len(groups))
	for i, v := range data {
		average[i], variance[i] = stat.MeanVariance(v, nil)
	}

	var sum float64
	var sumSquare float64
	for _, v := range data {
		for _, w := range v {
			sum += w
			sumSquare += w * w
		}
	}

	var S1_Square float64
	var total float64
	for i, v := range groupLen {
		S1_Square += float64(v) * average[i] * average[i]
		total += float64(v)
	}
	S1_Square -= sum * sum / total
	S1_Square /= float64(len(groupLen) - 1)

	var S2_Square float64
	for i, v := range groupLen {
		S2_Square += float64(v-1) * variance[i]
	}
	S2_Square /= (total - float64(len(groupLen)))
	fmt.Printf("S1 binh: \t %.5f\n", S1_Square)
	fmt.Printf("S2 binh: \t %.5f\n", S2_Square)

	result, F := kiemDinh(S1_Square/S2_Square, len(groups), int(total), 0.05)
	fmt.Println("K thuc nghiem: \t", S1_Square/S2_Square)
	fmt.Println("phan phoi F: \t", F)
	fmt.Println("Chap nhan: \t", result)
}
