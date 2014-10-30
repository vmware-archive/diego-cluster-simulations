package main_test

import (
	"math"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/auction/auctionrunner"
	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/auction/simulation/visualization"
	"github.com/cloudfoundry-incubator/auction/util"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/workpool"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Ω

var _ = Describe("Two Scenarios", func() {
	var initialDistributions map[int][]auctiontypes.SimulatedInstance

	newSimulatedInstance := func(processGuid string, index int, memoryMB int) auctiontypes.SimulatedInstance {
		return auctiontypes.SimulatedInstance{
			ProcessGuid:  processGuid,
			InstanceGuid: util.NewGuid("INS"),
			Index:        index,
			MemoryMB:     memoryMB,
			DiskMB:       1,
		}
	}

	generateUniqueSimulatedInstances := func(numInstances int, index int, memoryMB int) []auctiontypes.SimulatedInstance {
		instances := []auctiontypes.SimulatedInstance{}
		for i := 0; i < numInstances; i++ {
			instances = append(instances, newSimulatedInstance(util.NewGrayscaleGuid("AAA"), index, memoryMB))
		}
		return instances
	}

	newLRPStartAuction := func(processGuid string, memoryMB int) models.LRPStartAuction {
		return models.LRPStartAuction{
			DesiredLRP: models.DesiredLRP{
				ProcessGuid: processGuid,
				MemoryMB:    memoryMB,
				DiskMB:      1,
			},

			InstanceGuid: util.NewGuid("INS"),
			Index:        0,
		}
	}

	generateUniqueLRPStartAuctions := func(numInstances int, memoryMB int) []models.LRPStartAuction {
		instances := []models.LRPStartAuction{}
		for i := 0; i < numInstances; i++ {
			instances = append(instances, newLRPStartAuction(util.NewGrayscaleGuid("BBB"), memoryMB))
		}
		return instances
	}

	generateLRPStartAuctionsWithRandomColor := func(numInstances int, memoryMB int, colors []string) []models.LRPStartAuction {
		instances := []models.LRPStartAuction{}
		for i := 0; i < numInstances; i++ {
			color := colors[util.R.Intn(len(colors))]
			instances = append(instances, newLRPStartAuction(color, memoryMB))
		}
		return instances
	}

	runStartAuction := func(startAuctions []models.LRPStartAuction, i int, j int) {
		t := time.Now()
		results := auctionDistributor.HoldStartAuctions(numAuctioneers, startAuctions, repAddresses, auctionrunner.DefaultStartAuctionRules)
		duration := time.Since(t)
		report := &visualization.Report{
			RepAddresses:    repAddresses,
			AuctionResults:  results,
			InstancesByRep:  visualization.FetchAndSortInstances(client, repAddresses),
			AuctionDuration: duration,
		}
		visualization.PrintReport(client, len(startAuctions), results, repAddresses, duration, auctionrunner.DefaultStartAuctionRules)
		svgReport.DrawReportCard(i, j, report)
		reports = append(reports, report)
	}

	setInitialDistribution := func(initialDistribution map[int][]auctiontypes.SimulatedInstance) {
		workers := workpool.NewWorkPool(50)
		wg := &sync.WaitGroup{}
		wg.Add(len(initialDistributions))
		for index, simulatedInstances := range initialDistributions {
			index := index
			simulatedInstances := simulatedInstances
			workers.Submit(func() {
				client.SetSimulatedInstances(repAddresses[index], simulatedInstances)
				wg.Done()
			})
		}
		wg.Wait()
		workers.Stop()
	}

	BeforeEach(func() {
		util.ResetGuids()
		initialDistributions = map[int][]auctiontypes.SimulatedInstance{}
	})

	JustBeforeEach(func() {
		for index, simulatedInstances := range initialDistributions {
			client.SetSimulatedInstances(repAddresses[index], simulatedInstances)
		}
	})

	It("should quickly start 10% of the capacity", func() {
		for j := 0; j < numCells; j++ {
			initialDistributions[j] = generateUniqueSimulatedInstances(util.RandomIntIn(78, 80), 0, 1)
		}
		setInitialDistribution(initialDistributions)

		numApps := numCells * 10
		instances := generateUniqueLRPStartAuctions(numApps, 1)
		runStartAuction(instances, 0, 0)
	})

	It("should eventually manage a cold start", func() {
		numOneGigApps := numCells * 40
		numTwoGigApps := numCells * 20
		numFourGigApps := numCells * 3

		instances := []models.LRPStartAuction{}
		colors := []string{"purple", "red", "orange", "teal", "gray", "blue", "pink", "green", "lime", "cyan", "lightseagreen", "brown"}

		instances = append(instances, generateUniqueLRPStartAuctions(numOneGigApps/2, 1)...)
		instances = append(instances, generateLRPStartAuctionsWithRandomColor(numOneGigApps/2, 1, colors[:4])...)
		instances = append(instances, generateUniqueLRPStartAuctions(numTwoGigApps/2, 2)...)
		instances = append(instances, generateLRPStartAuctionsWithRandomColor(numTwoGigApps/2, 2, colors[4:8])...)
		instances = append(instances, generateUniqueLRPStartAuctions(numFourGigApps/2, 4)...)
		instances = append(instances, generateLRPStartAuctionsWithRandomColor(numFourGigApps/2, 4, colors[8:12])...)

		permutedInstances := make([]models.LRPStartAuction, len(instances))
		for i, index := range util.R.Perm(len(instances)) {
			permutedInstances[i] = instances[index]
		}

		runStartAuction(instances, 1, 0)
	})

	It("should run a rolling deploy", func() {
		numEmpty := int(math.Ceil(float64(numCells) * 0.05))
		for j := 0; j < numCells-numEmpty; j++ {
			initialDistributions[j] = generateUniqueSimulatedInstances(50, 0, 1)
		}
		setInitialDistribution(initialDistributions)

		numApps := numEmpty * 100
		instances := generateUniqueLRPStartAuctions(numApps, 1)

		runStartAuction(instances, 2, 0)
	})
})
