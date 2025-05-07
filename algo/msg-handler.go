package algo

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
)

func equalContacts(a, b []Contacts) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Id != b[i].Id {
			return false
		}
	}
	return true
}

// ~ let say this is for storing messages in memory
type Messages struct {
	SenderId       NodeID
	ReceiverId     NodeID
	MessageContent string
	MessageId      string
}
type MessagingPeer struct {
	ID           NodeID
	Messages     []Messages
	Port         int
	Address      string
	RoutingTable *RoutingTable
}

func (mp *MessagingPeer) StartServer() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", mp.Port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer ln.Close()

	fmt.Println("Listening on", mp.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go mp.handleConnection(conn)
	}
}

func (mp *MessagingPeer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	// Deserialize the message (can use JSON, gob, etc.)
	var msg Messages
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	fmt.Printf("Received message from %s: %s\n", msg.SenderId.String(), msg.MessageContent)

	// Save the message
	mp.Messages = append(mp.Messages, msg)
}

func (mp *MessagingPeer) FindNode(targetId NodeID) []Contacts {
	return mp.RoutingTable.FindClosestContacts(targetId, contactSize)
}

func (mp *MessagingPeer) IterativeFindNode(targetId NodeID, knownPeers map[string]*MessagingPeer) []Contacts {
	visited := make(map[string]bool)
	shortlist := mp.RoutingTable.FindClosestContacts(targetId, contactSize)
	closest := shortlist

	for {
		newShortlist := []Contacts{}
		for _, contact := range shortlist {
			if visited[contact.Id.String()] {
				continue
			}
			visited[contact.Id.String()] = true

			peer := knownPeers[contact.Id.String()]
			if peer == nil {
				continue
			}

			closerContacts := peer.FindNode(targetId)
			for _, c := range closerContacts {
				if !visited[c.Id.String()] {
					newShortlist = append(newShortlist, c)
				}
			}
		}

		all := append(closest, newShortlist...)
		sort.Slice(all, func(i, j int) bool {
			return targetId.XOR(all[i].Id).Cmp(targetId.XOR(all[j].Id)) < 0
		})

		if len(all) > contactSize {
			all = all[:contactSize]
		}

		if equalContacts(closest, all) {
			break
		}
		closest = all
		shortlist = newShortlist
	}

	return closest
}

func (mp *MessagingPeer) SendMessage(content string, peerId NodeID) (string, error) {

	if mp.ID == peerId {
		err := fmt.Errorf("Same peer can't send message to each other")
		return "", err
	}

	distance := mp.ID.XOR(peerId)
	index := GetMSBIndex(distance)
	isExist := mp.RoutingTable.buckets[index].Find(peerId)
	if isExist {
		return "msg send", nil // simulate
	}

	return "", nil
}

func NewMessagingPeer(port int, address string) *MessagingPeer {
	add := fmt.Sprintf("%s:%d", address, port)
	selfID := NewNodeId([]byte(add))
	return &MessagingPeer{
		ID:           selfID,
		Messages:     make([]Messages, 0),
		Port:         port,
		Address:      add,
		RoutingTable: NewRoutingTable(selfID),
	}
}
