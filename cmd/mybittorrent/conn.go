package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
)

func sendHandskake(conn net.Conn, infohash []byte) []byte {
	handshakeMessage := []byte{19}
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
	// conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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

func downloadPiece(conn net.Conn, filename string, pieceIdx uint32, pLength uint32, length uint32) []byte {
	var byteAcc uint32 = 0
	remaining := length - (pieceIdx * pLength)
	if remaining < pLength {
		pLength = remaining
	}
	piece := make([]byte, pLength)
	nBlocks := math.Ceil(float64(pLength) / (16 * 1024))
	for i := 0; i < int(nBlocks); i++ {
		slog.Warn(fmt.Sprintf("\n---------READING BLOCK %d tot size %d/%d----------\n", i, byteAcc, pLength))
		block := downloadBlock(conn, pieceIdx, uint32(i), pLength)
		slog.Info(fmt.Sprintf("Read block size %d\n", len(block)))
		_ = binary.BigEndian.Uint32(block[0:4])
		msgBegin := binary.BigEndian.Uint32(block[4:8])
		copy(piece[msgBegin:], block[8:])
	}
	return piece
}

func downloadBlock(conn net.Conn, pieceIdx uint32, n uint32, length uint32) []byte {
	payloadrequest := []byte{}
	var chunkSize uint32 = 16 * 1024
	begin := n * uint32(chunkSize)
	reqSize := chunkSize
	remaining := length - begin
	if remaining < reqSize {
		reqSize = remaining
	}
	payloadrequest = binary.BigEndian.AppendUint32(payloadrequest, pieceIdx) //index
	payloadrequest = binary.BigEndian.AppendUint32(payloadrequest, begin)    //begin
	payloadrequest = binary.BigEndian.AppendUint32(payloadrequest, reqSize)  // lenght
	slog.Warn("---------SENDING REQUEST----------")
	sendToConn(conn, 6, payloadrequest)
	slog.Warn("---------READING PIECE----------")
	payload := readFromConn(conn, 7)
	slog.Info(fmt.Sprintf("Read %d bytes of %d length payload\n", len(payload), length))
	return payload
}

func readFromConn(conn net.Conn, msgId uint8) []byte {
	h := make([]byte, 4)
	n, err := io.ReadFull(conn, h)
	if n != 4 {
		slog.Error(fmt.Sprintf("not reading 4 bytes but %d\n", n))
		panic(err)
	}
	if err != nil {
		slog.Error(err.Error())
		panic("cannot read from conn")
	}
	slog.Info(fmt.Sprintf("receiving header: %#v\n", h))
	length := binary.BigEndian.Uint32(h)
	slog.Info(fmt.Sprintf("Payload length: %d\n", length))
	slog.Info(fmt.Sprintf("Read message with %d msgid\n", msgId))
	msg := make([]byte, length)
	n, err = io.ReadFull(conn, msg)
	if n != int(length) {
		slog.Info(fmt.Sprintf("Read %d instead of %d\n", n, length))
		panic("not read enough data")
	}
	id := msg[0]
	if msgId != id {
		panic(fmt.Sprintf("Message id %d not equal to expected %d", id, msgId))
	}
	if err != nil {
		slog.Error(fmt.Sprint(n))
		slog.Error(err.Error())
		panic("cannot read from conn")
	}
	// slog.Info(fmt.Sprintf("Received: %#v\n", msg))
	return msg[1:]
}

func sendToConn(conn net.Conn, msgId uint8, payload []byte) {
	pLength := len(payload) + 1
	h := make([]byte, 0, 4)
	h = binary.BigEndian.AppendUint32(h, uint32(pLength))
	slog.Info(fmt.Sprintf("sending header: %#v\n", h))
	n, err := conn.Write(h)
	if n != 4 {
		slog.Error(fmt.Sprint(n))
		panic("cannot write size")
	}
	if err != nil {
		slog.Error(err.Error())
		panic("cannot write to conn")
	}
	msg := make([]byte, 1, pLength)
	msg[0] = msgId
	msg = append(msg, payload...)
	slog.Info(fmt.Sprintf("sending payload: %#v\n", msg))
	n, err = conn.Write(msg)
	if n != pLength {
		slog.Error(fmt.Sprint(n))
		panic("cannot write payload")
	}
	if err != nil {
		slog.Error(err.Error())
		panic("cannot write to conn")
	}
}
