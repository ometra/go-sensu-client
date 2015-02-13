/*
Copyright 2013-2014 Graham King

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

For full license details see <http://www.gnu.org/licenses/>.
*/

package metrics

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FIN byte = 1  // 00 0001
	SYN byte = 2  // 00 0010
	RST byte = 4  // 00 0100
	PSH byte = 8  // 00 1000
	ACK byte = 16 // 01 0000
	URG byte = 32 // 10 0000
)

type TCPHeader struct {
	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4 bits
	Reserved    uint8 // 3 bits
	ECN         uint8 // 3 bits
	Ctrl        uint8 // 6 bits
	Window      uint16
	Checksum    uint16 // Kernel will set this if it's 0
	Urgent      uint16
	Options     []TCPOption
}

type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte
}

// Parse packet into TCPHeader structure
func NewTCPHeader(data []byte) *TCPHeader {
	var tcp TCPHeader
	r := bytes.NewReader(data)
	binary.Read(r, binary.BigEndian, &tcp.Source)
	binary.Read(r, binary.BigEndian, &tcp.Destination)
	binary.Read(r, binary.BigEndian, &tcp.SeqNum)
	binary.Read(r, binary.BigEndian, &tcp.AckNum)

	var mix uint16
	binary.Read(r, binary.BigEndian, &mix)
	tcp.DataOffset = byte(mix >> 12)  // top 4 bits
	tcp.Reserved = byte(mix >> 9 & 7) // 3 bits
	tcp.ECN = byte(mix >> 6 & 7)      // 3 bits
	tcp.Ctrl = byte(mix & 0x3f)       // bottom 6 bits

	binary.Read(r, binary.BigEndian, &tcp.Window)
	binary.Read(r, binary.BigEndian, &tcp.Checksum)
	binary.Read(r, binary.BigEndian, &tcp.Urgent)

	return &tcp
}

func (tcp *TCPHeader) HasFlag(flagBit byte) bool {
	return tcp.Ctrl&flagBit != 0
}

func (tcp *TCPHeader) Marshal() []byte {

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, tcp.Source)
	binary.Write(buf, binary.BigEndian, tcp.Destination)
	binary.Write(buf, binary.BigEndian, tcp.SeqNum)
	binary.Write(buf, binary.BigEndian, tcp.AckNum)

	var mix uint16
	mix = uint16(tcp.DataOffset)<<12 | // top 4 bits
		uint16(tcp.Reserved)<<9 | // 3 bits
		uint16(tcp.ECN)<<6 | // 3 bits
		uint16(tcp.Ctrl) // bottom 6 bits
	binary.Write(buf, binary.BigEndian, mix)

	binary.Write(buf, binary.BigEndian, tcp.Window)
	binary.Write(buf, binary.BigEndian, tcp.Checksum)
	binary.Write(buf, binary.BigEndian, tcp.Urgent)

	for _, option := range tcp.Options {
		binary.Write(buf, binary.BigEndian, option.Kind)
		if option.Length > 1 {
			binary.Write(buf, binary.BigEndian, option.Length)
			binary.Write(buf, binary.BigEndian, option.Data)
		}
	}

	out := buf.Bytes()

	// Pad to min tcp header size, which is 20 bytes (5 32-bit words)
	pad := 20 - len(out)
	for i := 0; i < pad; i++ {
		out = append(out, 0)
	}

	return out
}

// TCP Checksum
func csum(data []byte, srcip, dstip [4]byte) uint16 {

	pseudoHeader := []byte{
		srcip[0], srcip[1], srcip[2], srcip[3],
		dstip[0], dstip[1], dstip[2], dstip[3],
		0,                  // zero
		6,                  // protocol number (6 == TCP)
		0, byte(len(data)), // TCP length (16 bits), not inc pseudo header
	}

	sumThis := make([]byte, 0, len(pseudoHeader)+len(data))
	sumThis = append(sumThis, pseudoHeader...)
	sumThis = append(sumThis, data...)
	//fmt.Printf("% x\n", sumThis)

	lenSumThis := len(sumThis)
	var nextWord uint16
	var sum uint32
	for i := 0; i+1 < lenSumThis; i += 2 {
		nextWord = uint16(sumThis[i])<<8 | uint16(sumThis[i+1])
		sum += uint32(nextWord)
	}
	if lenSumThis%2 != 0 {
		//fmt.Println("Odd byte")
		sum += uint32(sumThis[len(sumThis)-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}

func latency(localAddr string, remoteHost string, port uint16) (time.Duration, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	var receiveTime time.Time

	remoteAddr, err := getRemoteAddress(remoteHost)
	if err != nil {
		return time.Duration(0), err
	}

	go func() {
		receiveTime, err = receiveSynAck(localAddr, remoteAddr)
		wg.Done()
	}()

	time.Sleep(1 * time.Millisecond)
	sendTime := sendSyn(localAddr, remoteAddr, port)

	wg.Wait()
	return receiveTime.Sub(sendTime), err
}

/**
Used in a couple of places. It works around sometimes net.LookupHost borking on IP addresses
*/
func getRemoteAddress(address string) (string, error) {
	remoteAddr := ""
	if ip := net.ParseIP(address); ip != nil {
		remoteAddr = ip.String()
	} else {
		addrs, err := net.LookupHost(address)
		if err != nil {
			return "", fmt.Errorf("Error resolving %s. %s", address, err)
		} else {
			remoteAddr = addrs[0]
		}
	}
	return remoteAddr, nil
}

func interfaceAddress(ifaceName string) (net.Addr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	var dummy net.Addr
	if err != nil {
		return dummy, fmt.Errorf("net.InterfaceByName for %s. %s", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return dummy, fmt.Errorf("iface.Addrs: %s", err)
	}
	if len(addrs) == 0 {
		return dummy, fmt.Errorf("iface.Addrs: No Addresses")
	}

	// look for an IPv4 Addr
	ip, _, err := net.ParseCIDR(addrs[0].String())
	if err != nil {
		return dummy, err
	}

	if nil == ip.To4() {
		return dummy, fmt.Errorf("No IPv4 Addresses for interface %s", ifaceName)
	}

	return addrs[0], nil
}

func sendSyn(laddr, raddr string, port uint16) time.Time {
	var sendTime time.Time
	packet := TCPHeader{
		Source:      0xaa47, // Random ephemeral port
		Destination: port,
		SeqNum:      rand.Uint32(),
		AckNum:      0,
		DataOffset:  5,      // 4 bits
		Reserved:    0,      // 3 bits
		ECN:         0,      // 3 bits
		Ctrl:        2,      // 6 bits (000010, SYN bit set)
		Window:      0xaaaa, // 43690, dunno, copied it
		Checksum:    0,      // Kernel will set this if it's 0
		Urgent:      0,
		Options:     []TCPOption{},
	}

	data := packet.Marshal()
	packet.Checksum = csum(data, to4byte(laddr), to4byte(raddr))

	data = packet.Marshal()

	//fmt.Printf("% x\n", data)

	conn, err := net.Dial("ip4:tcp", raddr)
	if err != nil {
		log.Printf("Dial: %s\n", err)
		return sendTime
	}

	sendTime = time.Now()

	numWrote, err := conn.Write(data)
	if err != nil {
		log.Printf("Write: %s\n", err)
		return sendTime
	}
	if numWrote != len(data) {
		log.Printf("Short write. Wrote %d/%d bytes\n", numWrote, len(data))
		return sendTime
	}

	conn.Close()

	return sendTime
}

func to4byte(addr string) [4]byte {
	parts := strings.Split(addr, ".")
	b0, _ := strconv.Atoi(parts[0])
	b1, _ := strconv.Atoi(parts[1])
	b2, _ := strconv.Atoi(parts[2])
	b3, _ := strconv.Atoi(parts[3])
	return [4]byte{byte(b0), byte(b1), byte(b2), byte(b3)}
}

func receiveSynAck(localAddress, remoteIp string) (time.Time, error) {
	var receiveTime time.Time
	netaddr, err := net.ResolveIPAddr("ip4", localAddress)
	if err != nil {
		log.Printf("net.ResolveIPAddr: %s. %s\n", localAddress, netaddr)
		return receiveTime, err
	}

	conn, err := net.ListenIP("ip4:tcp", netaddr)
	if err != nil {
		log.Printf("ListenIP: %s\n", err)
		return receiveTime, err
	}
	for {
		buf := make([]byte, 1024)
		numRead, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("ReadFrom: %s\n", err)
			break
		}
		if remoteIp != raddr.String() {
			// this is not the packet (droid?) we are looking for
			continue
		}
		//fmt.Printf("Received: % x\n", buf[:numRead])
		tcp := NewTCPHeader(buf[:numRead])
		// Closed port gets RST, open port gets SYN ACK
		if tcp.HasFlag(RST) || (tcp.HasFlag(SYN) && tcp.HasFlag(ACK)) {
			receiveTime = time.Now()
			break
		}
	}
	return receiveTime, nil
}
