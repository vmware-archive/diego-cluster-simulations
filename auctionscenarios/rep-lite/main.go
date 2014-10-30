package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/cloudfoundry-incubator/auction/communication/http/routes"
	"github.com/cloudfoundry-incubator/auction/communication/nats/auction_nats_server"
	"github.com/cloudfoundry/yagnats"
	"github.com/pivotal-golang/lager"

	"github.com/tedsuo/rata"

	"github.com/cloudfoundry-incubator/auction/auctionrep"
	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/auction/communication/http/auction_http_handlers"
	"github.com/cloudfoundry-incubator/auction/simulation/simulationrepdelegate"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var repGuid = flag.String("repGuid", "", "rep-guid")

var natsUsername = flag.String("natsUsername", "", "nats username")
var natsPassword = flag.String("natsPassword", "", "nats password")
var natsAddresses = flag.String("natsAddresses", "", "nats addresses")

func main() {
	flag.Parse()

	if *repGuid == "" {
		panic("need rep-guid")
	}

	repDelegate := simulationrepdelegate.New(auctiontypes.Resources{
		MemoryMB:   100,
		DiskMB:     100,
		Containers: 100,
	})
	rep := auctionrep.New(*repGuid, repDelegate)

	go serveOverNATS(rep)

	handlers := auction_http_handlers.New(rep, lager.NewLogger("rep-lite-http"))
	router, err := rata.NewRouter(routes.Routes, handlers)
	if err != nil {
		log.Fatalln("failed to make router:", err)
	}
	httpServer := http_server.New("0.0.0.0:8080", router)

	monitor := ifrit.Envoke(sigmon.New(httpServer))

	fmt.Println("rep node listening")
	err = <-monitor.Wait()
	if err != nil {
		println("EXITED WITH ERROR: ", err.Error())
	}
}

func serveOverNATS(rep *auctionrep.AuctionRep) {
	if *natsAddresses != "" && *natsUsername != "" && *natsPassword != "" {
		natsMembers := []string{}
		for _, addr := range strings.Split(*natsAddresses, ",") {
			uri := url.URL{
				Scheme: "nats",
				Host:   addr,
				User:   url.UserPassword(*natsUsername, *natsPassword),
			}
			natsMembers = append(natsMembers, uri.String())
		}

		client, err := yagnats.Connect(natsMembers)
		if err != nil {
			log.Fatalln("no nats:", err)
		}

		natsRunner := auction_nats_server.New(client, rep, lager.NewLogger("rep-lite-nats"))
		ifrit.Envoke(sigmon.New(natsRunner))
	}
}
