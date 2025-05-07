package main

import "github.com/HadeedTariq/p-to-p-msg-system-kademlia/algo"

func main() {
	peerA := algo.NewMessagingPeer(8001, "127.0.0.1")
	peerB := algo.NewMessagingPeer(8002, "127.0.0.1")
	peerC := algo.NewMessagingPeer(8003, "127.0.0.1")

	// manually add each other to routing tables for bootstrapping
	peerA.RoutingTable.Add(algo.Contacts{Id: peerB.ID, Address: peerB.Address})
	peerB.RoutingTable.Add(algo.Contacts{Id: peerC.ID, Address: peerC.Address})

	network := map[string]*algo.MessagingPeer{
		peerA.ID.String(): peerA,
		peerB.ID.String(): peerB,
		peerC.ID.String(): peerC,
	}

	results := peerA.IterativeFindNode(peerC.ID, network)

	select {}
}
