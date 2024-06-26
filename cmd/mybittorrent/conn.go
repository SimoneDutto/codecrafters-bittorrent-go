package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func sendHandskake(conn net.Conn, infohash []byte) []byte {
	handshakeMessage := []byte{byte(19)}
	handshakeMessage = append(handshakeMessage, []byte("BitTorrent protocol")...)
	handshakeMessage = binary.BigEndian.AppendUint64(handshakeMessage, 0)
	handshakeMessage = append(handshakeMessage, infohash...)
	handshakeMessage = append(handshakeMessage, []byte("01234512123123231123")...)
	slog.Info(fmt.Sprintf("Handshake: %#v\n", handshakeMessage))
	n, err := conn.Write(handshakeMessage)
	if err != nil || n != 68 {
		slog.Error(err.Error())
		panic("cannot send handshake")
	}
	res := make([]byte, 68)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err = conn.Read(res)
	if err != nil || n != 68 {
		slog.Error(err.Error())
		panic("cannot receive handshake")
	}
	return res
}

func getPeers(announce string, hashInfo []byte, pieceL interface{}) []string {
	query := prepareRequest(string(hashInfo), pieceL.(int64))
	URL := fmt.Sprintf("%s?%s", announce, query)
	slog.Info(URL)
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	decoded, _ := decodeBencode(string(data), []interface{}{}, 0)
	slog.Info(fmt.Sprintf("Response: %#v", decoded[0]))
	return extractPeers(decoded[0].(map[string]interface{})["peers"])
}

func unchoke(conn net.Conn) {
	slog.Warn("---------READING BITFIELD----------")
	_ = readFromConn(conn, 5)
	slog.Warn("---------SENDING INTERESTED----------")
	sendToConn(conn, 2, []byte{})
	slog.Warn("---------READING UNCHOKE----------")
	readFromConn(conn, 1)
}

func readFromConn(conn net.Conn, msgId uint8) []byte {
	h := make([]byte, 5)
	n, err := conn.Read(h)
	if err != nil || n != 5 {
		slog.Error(err.Error())
		panic("cannot read from conn")
	}
	slog.Info(fmt.Sprintf("receiving header: %#v\n", h))
	// length := binary.BigEndian.Uint32(h[:4])
	length := 100
	slog.Info(fmt.Sprintf("Payload length: %d\n", length))
	id := uint8(h[4])
	if msgId != id {
		panic(fmt.Sprintf("Message id %d not equal to expected %d", id, msgId))
	}
	if length == 0 {
		return []byte{}
	}
	slog.Info(fmt.Sprintf("Read message with %d msgid\n", msgId))
	// msg := make([]byte, int(length))
	// n, err = conn.Read(msg)
	// if err != nil || n != int(length) {
	// 	slog.Error(err.Error())
	// 	panic("cannot read from conn")
	// }
	// slog.Info(fmt.Sprintf("Received: %#v\n", msg))
	return []byte{}
}

func sendToConn(conn net.Conn, msgId uint8, payload []byte) {
	pLength := len(payload)
	h := make([]byte, 0, 5)
	h = binary.BigEndian.AppendUint32(h, uint32(pLength))
	h = append(h, byte(msgId))
	slog.Info(fmt.Sprintf("sending header: %#v\n", h))
	n, err := conn.Write(h)
	if err != nil || n != 5 {
		slog.Error(err.Error())
		panic("cannot write to conn")
	}
	if pLength == 0 {
		return
	}
}
