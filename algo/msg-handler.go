package algo

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

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
	var msg Messages
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	fmt.Printf("Received message from %s: %s\n", msg.SenderId.String(), msg.MessageContent)

	mp.Messages = append(mp.Messages, msg)
}
func (mp *MessagingPeer) SendMessage(content string, peerId NodeID) (string, error) {
	if mp.ID == peerId {
		return "", fmt.Errorf("same peer can't send message to itself")
	}

	distance := mp.ID.XOR(peerId)
	index := GetMSBIndex(distance)
	peerAddress := mp.RoutingTable.buckets[index].Find(peerId)

	if peerAddress == "" {
		return "", fmt.Errorf("peer not found in routing table")
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
