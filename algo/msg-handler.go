package algo

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
)

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
	var msg Messages
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	fmt.Printf("Received message from %s: %s\n", msg.SenderId.String(), msg.MessageContent)

	mp.Messages = append(mp.Messages, msg)
}

func (mp *MessagingPeer) FindNode(targetID NodeID) []Contacts {
	return mp.RoutingTable.FindClosestContacts(targetID, contactSize)
}

func (mp *MessagingPeer) IterativeFindNode(targetId NodeID, knownPeers map[string]*MessagingPeer) []Contacts {

	visited := make(map[string]bool)

	shortList := mp.RoutingTable.FindClosestContacts(targetId, contactSize)
	closest := shortList

	for {
		newShortList := []Contacts{}

		for _, contact := range shortList {
			idStr := contact.Id.String()
			if visited[idStr] {
				continue
			}

			visited[idStr] = true
			peer := knownPeers[idStr]
			if peer == nil {
				continue
			}
			closerContacts := peer.FindNode(targetId)

			for _, c := range closerContacts {
				if !visited[c.Id.String()] {
					newShortList = append(newShortList, c)
				}
			}
		}
		all := append(closest, newShortList...)
		sort.Slice(all, func(i, j int) bool {
			return targetId.XOR(all[i].Id).Cmp(targetId.XOR(all[j].Id)) < 0
		})

		if len(all) > contactSize {
			all = all[:contactSize]
		}

		if EqualContacts(closest, all) {
			break
		}

		closest = all
		shortList = newShortList
	}

	return closest

}

func (mp *MessagingPeer) SendMessage(content string, peerId NodeID, network map[string]*MessagingPeer) (string, error) {
	if mp.ID == peerId {
		return "", fmt.Errorf("same peer can't send message to itself")
	}

	distance := mp.ID.XOR(peerId)
	index := GetMSBIndex(distance)
	peerAddress := mp.RoutingTable.buckets[index].Find(peerId)

	if peerAddress == "" {

		closest := mp.IterativeFindNode(peerId, network)

		for _, con := range closest {
			if con.Id == peerId {
				peerAddress = con.Address
				break
			}
		}
	}

	if peerAddress == "" {
		return "", fmt.Errorf("Could not find peer address")
	}

	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer conn.Close()

	msg := &Messages{
		SenderId:       mp.ID,
		ReceiverId:     peerId,
		MessageContent: content,
		MessageId:      "Random",
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize message: %w", err)
	}

	_, err = conn.Write(msgBytes)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return "message sent successfully", nil
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
