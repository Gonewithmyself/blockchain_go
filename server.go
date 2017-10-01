package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const dnsNodeID = 3000
const nodeVersion = 1
const commandLength = 12

var nodeAddress string

type verzion struct {
	Version  int
	AddrFrom string
}

type verack struct {
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	fmt.Printf("%x\n", data)
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendVersion(addr string) {
	payload := gobEncode(verzion{nodeVersion, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

func sendVrack(addr string) {
	payload := gobEncode(verack{})

	request := append(commandToBytes("verack"), payload...)

	sendData(addr, request)
}

func handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	sendVrack(payload.AddrFrom)
}

func handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])

	switch command {
	case "version":
		fmt.Printf("Received %s command\n", command)
		handleVersion(request)
	case "verack":
		fmt.Printf("Received %s command\n", command)
	default:
		fmt.Println("Unknown command received!")
	}

	conn.Close()
}

// StartServer starts a node
func StartServer(nodeID int) {
	nodeAddress = fmt.Sprintf("localhost:%d", nodeID)
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	if nodeID != dnsNodeID {
		sendVersion(fmt.Sprintf("localhost:%d", dnsNodeID))
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn)
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}