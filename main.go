package main

import (
	"time"

	"github.com/HadeedTariq/p-to-p-msg-system-kademlia/algo"
)

func main() {
	// ~ so how the messaging system looks like the peers connect

	// ~ so how does the peer connects
	peerA := algo.NewMessagingPeer(8000, "localhost")
	go peerA.StartServer()

	// Peer B
	peerB := algo.NewMessagingPeer(8001, "localhost")
	go peerB.StartServer()

	// Peer C
	peerC := algo.NewMessagingPeer(8002, "localhost")
	go peerC.StartServer()

	// Peer D
	peerD := algo.NewMessagingPeer(8003, "localhost")
	go peerD.StartServer()
	time.Sleep(1 * time.Second)

	peerB.RoutingTable.Add(algo.Contacts{
		Id:      peerA.ID,
		Address: peerA.Address,
	})
	peerB.RoutingTable.Add(algo.Contacts{
		Id:      peerC.ID,
		Address: peerC.Address,
	})
	peerB.RoutingTable.Add(algo.Contacts{
		Id:      peerD.ID,
		Address: peerD.Address,
	})

	peerB.SendMessage("Hello from B to A", peerA.ID)

	select {}
}
