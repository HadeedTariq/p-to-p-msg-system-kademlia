package main

import (
	"fmt"
	"time"

	"github.com/HadeedTariq/p-to-p-msg-system-kademlia/algo"
)

func main() {

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

	// Peer E
	peerE := algo.NewMessagingPeer(8004, "localhost")
	go peerE.StartServer()
	time.Sleep(1 * time.Second)

	peerB.RoutingTable.Add(algo.Contacts{
		Id:      peerA.ID,
		Address: peerA.Address,
	})
	peerB.RoutingTable.Add(algo.Contacts{
		Id:      peerC.ID,
		Address: peerC.Address,
	})
	peerC.RoutingTable.Add(algo.Contacts{
		Id:      peerD.ID,
		Address: peerD.Address,
	})
	peerD.RoutingTable.Add(algo.Contacts{
		Id:      peerE.ID,
		Address: peerE.Address,
	})

	// ~ in real world we don't have access to all the networks so we have to use the ping mechanism
	network := map[string]*algo.MessagingPeer{}
	network[peerA.ID.String()] = peerA
	network[peerB.ID.String()] = peerB
	network[peerC.ID.String()] = peerC
	network[peerD.ID.String()] = peerD
	network[peerE.ID.String()] = peerE

	msg, err := peerB.SendMessage("Hello from B to E", peerE.ID, network)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(msg)

	select {}
}
