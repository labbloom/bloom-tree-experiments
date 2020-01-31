package main

import (
	"fmt"
	"strconv"
	"strings"
	"io"
    "log"
	"os"
	"math"
	"github.com/montanaflynn/stats"
	"github.com/labbloom/DBF"
	bloomtree "github.com/labbloom/bloom-tree"

)

func WriteToFile(filename string, data string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.WriteString(file, data)
    if err != nil {
        return err
    }
    return file.Sync()
}

func main() {
	fileName := "data/data.csv"
	fpr := []float64{0.1, 0.01, 0.001}
	n := []int{500, 1000, 5000, 10000}
	seed := []byte("seed")

	finalResults := make([][][]float64, len(fpr))

	for index, fprValue := range fpr {
		results := make([][]float64, len(n))
		for ind, val := range n {
			elements := make([][]byte, val)
			dbf := DBF.NewDbf(uint(len(elements)), fprValue, seed)
			for i := 0; i < len(elements); i ++ {
				a := []byte(strconv.FormatInt(int64(i), 10))
				elements[i] = a
				dbf.Add(a)
			}
			bt, err := bloomtree.NewBloomTree(dbf)
			if err != nil {
				panic(err)
			}
			proofSizes := make([]int, val)
			for i := 0; i < len(elements); i ++ {
				presenceMultiproof, _ := bt.GenerateCompactMultiProof(elements[i])
				proofSizes[i] = len(presenceMultiproof.Chunks)*8 + len(presenceMultiproof.Proof)*32 + 1
			}
			data := stats.LoadRawData(proofSizes)
			median, _:= stats.Median(data)
			roundedMedian, _ := stats.Round(median, 0)
			mean, _ := stats.Mean(data)
			roundedMean, _ := stats.Round(mean, 0)
			bloomFilterSize := float64(len(dbf.BitArray().Bytes())) * 8
			absenceSize := (math.Log2(float64(math.Exp2(math.Ceil(math.Log2(float64(len(dbf.BitArray().Bytes()))))))) * 32) + 8 + 1
			results[ind] = []float64{fprValue, float64(val), absenceSize, roundedMedian, roundedMean, bloomFilterSize}
		}
		finalResults[index] = results
	}

	csvResult := "fpr,n,absence,presence_median,presence_mean,bloomFilterSize\n"
	for _, mat := range finalResults {
		for _, row := range mat {
			st := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), ", "), "[]")
			csvResult = csvResult + st + "\n"
		}
	}
	err := WriteToFile(fileName, csvResult)
	if err != nil {
		log.Fatal(err)
	}

}
