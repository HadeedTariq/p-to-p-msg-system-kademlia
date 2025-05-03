package algo

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"sort"
	"time"
)

const IdLength = 8
const IdBits = 16

type NodeID [IdBits / IdLength]byte

func (id NodeID) String() string {
	return hex.EncodeToString(id[:])
}

func (id NodeID) XOR(otherId NodeID) *big.Int {
	var result NodeID

	for i := 0; i < len(id); i++ {
		result[i] = id[i] ^ otherId[i]
	}

	return new(big.Int).SetBytes(result[:])
}

const contactSize = 2

type Contacts struct {
	Id           NodeID
	Address      string
	last_seen_at time.Time
}

type KBucket struct {
	contacts []Contacts
}

func (kb *KBucket) Add(contact Contacts) bool {
	for _, c := range kb.contacts {
		if c.Id == contact.Id {
			return true
		}
	}
	if len(kb.contacts) < contactSize {
		kb.contacts = append(kb.contacts, contact)
		return true
	} else {
		evicted := kb.Evict()
		if evicted {
			kb.contacts = append(kb.contacts, contact)
			return true
		} else {
			fmt.Println("There is no space in this bucket")
			return false
		}
	}
}

func (kb *KBucket) Evict() bool {
	sort.Slice(kb.contacts, func(i, j int) bool {
		return kb.contacts[i].last_seen_at.Before(kb.contacts[j].last_seen_at)
	})
	contact := kb.contacts[0]
	_, err := net.Dial("tcp", contact.Address)
	if err != nil {
		kb.contacts = append(kb.contacts[:0], kb.contacts[0+1:]...)
		return true
	}
	return false
}

type RoutingTable struct {
	buckets []KBucket
	selfId  NodeID
}

func GetMSBIndex(distance *big.Int) int {
	if distance.Sign() == 0 {
		return -1
	}
	return distance.BitLen() - 1
}

func (rt *RoutingTable) Add(contact Contacts) {
	distance := rt.selfId.XOR(contact.Id)
	index := GetMSBIndex(distance)
	if index >= 0 {
		rt.buckets[index].Add(contact)
	}
}

func NewNodeId(data []byte) NodeID {
	hash := sha1.Sum(data[:])
	var newId NodeID
	copy(newId[:], hash[:len(newId)])

	return newId
}

func NewRoutingTable(selfID NodeID) *RoutingTable {
	buckets := make([]KBucket, IdBits)
	for i := 0; i < len(buckets); i++ {
		buckets[i] = KBucket{contacts: make([]Contacts, 0, contactSize)}
	}

	return &RoutingTable{buckets: buckets, selfId: selfID}
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

// ~ so now as the structure is build now my main focus is towards creating the node-id and functionalities around routing table
func NewMessagingPeer(port int, address string) *MessagingPeer {
	add := fmt.Sprintf("%s:%d", address, port)
	selfID := NewNodeId([]byte(add))
	return &MessagingPeer{
		ID:           selfID,
		Messages:     make([]Messages, 0),
		Port:         port,
		Address:      address,
		RoutingTable: NewRoutingTable(selfID),
	}
}
