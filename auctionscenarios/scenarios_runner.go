package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	numCells := []int{25, 50, 100, 150, 200, 250, 300, 350, 400}
	maxConcurrent := []int{1, 2, 5}
	biddingPoolFraction := []float64{0.05, 0.1, 0.2}
	algorithm := []string{"compare_to_percentile", "all_rebid"}

	for _, numCell := range numCells {
		for _, maxConc := range maxConcurrent {
			for _, bidPool := range biddingPoolFraction {
				for _, alg := range algorithm {
					cmd := exec.Command(
						"./auctionscenarios.test",
						fmt.Sprintf("--numCells=%d", numCell),
						fmt.Sprintf("--maxConcurrent=%d", maxConc),
						fmt.Sprintf("--maxBiddingPoolFraction=%.2f", bidPool),
						fmt.Sprintf("--algorithm=%s", alg),
					)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					t := time.Now()
					fmt.Printf("Running %d cells, %d maxConcurrent, %.2f maxBiddingPoolFraction, %s algorithm\n", numCell, maxConc, bidPool, alg)
					cmd.Run()
					fmt.Printf("Done in %s\n", time.Since(t))
				}
			}
		}
	}
}
