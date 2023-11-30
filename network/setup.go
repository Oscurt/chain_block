package network

import (
    "fmt"
    "github.com/libp2p/go-libp2p"
    //"reflect"
)

/*
func discoverPeers(h host.Host, discoveryService *discovery.MdnsDiscovery) {
    ticker := time.NewTicker(time.Second * 30)
    for {
        select {
        case <-ticker.C:
            peers, err := discoveryService.FindPeers(context.Background(), "my-network")
            if err != nil {
                continue
            }
            for _, p := range peers {
                if p.ID != h.ID() {
                    h.Connect(context.Background(), p)
                }
            }
        }
    }
}*/

func Init(){

    h, err := libp2p.New()
    if err != nil {
        panic(err)
    }

    ser, err := mdns.NewMdnsService(h, "blockchain-network", nil)
    if err != nil {
        panic(err)
    }

    discoveryService := discovery.NewMdnsDiscovery(ser)


    fmt.Println(discoveryService)
    //go discoverPeers(h, discoveryService)

}