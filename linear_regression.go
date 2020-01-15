package main

import (
	"fmt"
	"io"
	"log"
	"math"
	_ "math/rand"
	"os"

	"encoding/csv"
	"image/color"

	"github.com/kniren/gota/dataframe"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"./sample_utils"
)

func readCSV(name string) (dataframe.DataFrame, error) {
	// insert the file name
	file, err := os.Open(name)
	if err != nil {
		return dataframe.New(), fmt.Errorf("Cannot open file %s: %w", name, err)
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
			log.Printf("Bad data: %w", err)
		}

		series = append(series, record)
	}

	return dataframe.LoadRecords(series), nil
}

func turnDataFrameIntoXYer(data dataframe.DataFrame) plotter.XYs {
	points := make(plotter.XYs, data.Nrow())

	for i := range points {
		points[i].X = data.Elem(i, 0).Float()
		points[i].Y = data.Elem(i, 1).Float()
	}

	return points
}

func linear_variance(alpha, beta float64, point_x, point_y []float64) float64 {
	var variance_square float64

	for index := range point_x {
		variance_square += math.Pow(point_y[index]-beta*point_x[index]-alpha, 2)
	}
	variance_square = variance_square / float64(len(point_x)-2)

	return math.Sqrt(variance_square)
}

func sum_square(x []float64) float64 {
	var sum float64

	for _, each_element := range x {
		sum += each_element * each_element
	}

	return sum
}

func alpha_beta_interval(alpha, beta, trust_interval float64, point_x, point_y []float64) (alpha_interval, beta_interval [2]float64) {
	data_len := float64(len(point_x))
	t := new(distuv.StudentsT)
	t.Nu = float64(data_len - 2)
	t.Mu = 0
	t.Sigma = 1

	variance := sample_utils.LinearVariance(alpha, beta, point_x, point_y)
	_, x_variance := stat.MeanVariance(point_x, nil)
	x_variance *= float64(len(point_x) - 1)

	beta_interval[0] = beta - (variance * t.Quantile(trust_interval) / math.Sqrt(x_variance))
	beta_interval[1] = beta + (variance * t.Quantile(trust_interval) / math.Sqrt(x_variance))
	alpha_interval[0] = alpha - (variance * t.Quantile(trust_interval) * math.Sqrt((sample_utils.SumSquare(point_x) / data_len / x_variance)))
	alpha_interval[1] = alpha + (variance * t.Quantile(trust_interval) * math.Sqrt((sample_utils.SumSquare(point_x) / data_len / x_variance)))

	return alpha_interval, beta_interval
}

func main() {
	data, err := readCSV("new.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new graph for plotting
	graph, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	graph.X.Label.Text = "X"
	graph.Y.Label.Text = "Y"
	graph.Add(plotter.NewGrid())

	// Turn data into XYer type
	points := turnDataFrameIntoXYer(data)

	// Setup the default glyph and the default color
	plotutil.DefaultGlyphShapes[0] = draw.PlusGlyph{}
	plotutil.DefaultColors = plotutil.DarkColors
	plotutil.DefaultColors[0] = color.RGBA{157, 0, 6, 255}

	// Plot point to the plane
	if err := plotutil.AddScatters(graph, points); err != nil {
		log.Fatal(err)
	}

	// Plot linear function
	point_x := make([]float64, data.Nrow())
	point_y := make([]float64, data.Nrow())

	for i := range point_x {
		point_x[i] = data.Elem(i, 0).Float()
		point_y[i] = data.Elem(i, 1).Float()
	}

	alpha, beta := stat.LinearRegression(point_x, point_y, nil, false)
	linearFunction := plotter.NewFunction(func(x float64) float64 { return beta*x + alpha })

	if err := plotutil.AddLines(graph, linearFunction); err != nil {
		log.Fatal(err)
	}

	// Save the graph into point.png file
	if err := graph.Save(10*vg.Inch, 6*vg.Inch, "point.png"); err != nil {
		log.Fatal(err)
	}

	// Calculate the R^2 of the regression
	alpha_interval, beta_interval := alpha_beta_interval(alpha, beta, 0.95, point_x, point_y)
	r2 := stat.RSquared(point_x, point_y, nil, alpha, beta)

	// Print the result
	fmt.Println("The r2 is: ", r2)
	fmt.Println("The linear variance is: ", sample_utils.LinearVariance(alpha, beta, point_x, point_y))
	fmt.Printf("The beta interval is: (%.10f, %.10f)\n", beta_interval[0], beta_interval[1])
	fmt.Printf("The alpha interval is: (%.10f, %.10f)\n", alpha_interval[0], alpha_interval[1])
}
