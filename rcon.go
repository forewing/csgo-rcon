// Package rcon provides a golang interface of Source Remote Console (RCON) client, let server operators to administer
// and interact with their servers remotely in the same manner as the console provided by srcds.
// Based on http://developer.valvesoftware.com/wiki/Source_RCON_Protocol
package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	serverdataAuth         = 3
	serverdataAuthResponse = 2

	serverdataExecCommand   = 2
	serverdataResponseValue = 0

	// found by trial & error
	maxCommandLength = 510

	// command (4), id (4), string1 (1), string2 (1)
	minMessageLength = 4 + 4 + 1 + 1
	// command (4), id (4), string (4096), string2 (1)
	maxMessageLength = 4 + 4 + 4096 + 1

	probablySplitIfLargerThan = maxMessageLength - 400
)

const (
	// DefaultPort of the srcds RCON
	DefaultPort = 27015

	// DefaultPassword is empty
	DefaultPassword = ""

	// DefaultTimeout of the connection
	DefaultTimeout = time.Second * 1
)

const (
	tcpNetworkName = "tcp"
	authSuccess    = "success"
)

var (
	errNoConnection     = errors.New("no connection")
	errDialTCPFail      = errors.New("dial TCP fail")
	errConnectionClosed = errors.New("connection closed")
	errBadPassword      = errors.New("bad password")
	errInvalidResponse  = errors.New("invalid response")
	errCrapBytes        = errors.New("response contains crap bytes")
	errWaitingTimeout   = errors.New("timeout while waiting for reply")
)

// A Client of RCON protocol to srcds
// Remember to set Timeout, it will block forever when not set
type Client struct {
	address  string
	password string
	timeout  time.Duration

	reqID   int32
	tcpConn *net.TCPConn

	lock sync.Mutex
}

// New return pointer to a new client, it's safe for concurrency use
func New(address, password string, timeout time.Duration) *Client {
	c := &Client{
		address:  address,
		password: password,
		timeout:  timeout,
	}
	if c.timeout <= 0 {
		c.timeout = DefaultTimeout
	}
	return c
}

// Execute the command.
// Execute once if no "\n" provided. Return result message and nil on success, empty string and an error on failure.
// If cmd includes "\n", it is treated by a script file. Splitted and trimmed into lines. Line starts with "//" will
// be treated as comment and ignored. When all commands seccess, concatted messages and nil will be returned.
// Once failed, concatted previous succeeded messages and an error will be returned.
func (c *Client) Execute(cmd string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cmds := strings.Split(cmd, "\n")
	if len(cmds) == 1 {
		return c.executeWorker(cmd)
	}

	var builder strings.Builder
	for i := range cmds {
		cmd := strings.TrimSpace(cmds[i])
		if len(cmd) == 0 || strings.HasPrefix(cmd, "//") {
			continue
		}

		result, err := c.executeWorker(cmd)
		if err != nil {
			return builder.String(), err
		}

		builder.WriteString(result)
	}
	return builder.String(), nil
}

func (c *Client) executeWorker(cmd string) (string, error) {
	err := c.send(serverdataExecCommand, cmd)
	if err != nil {
		return c.executeRetry(cmd)
	}
	str1, err := c.receive()
	if err != nil {
		return c.executeRetry(cmd)
	}
	return str1, nil
}

func (c *Client) executeRetry(cmd string) (string, error) {
	c.disconnect()
	if err := c.connect(); err != nil {
		return "", err
	}
	c.send(serverdataAuth, c.password)

	auth, err := c.receive()
	if err != nil {
		return "", err
	}
	if len(auth) == 0 {
		auth, err := c.receive()
		if err != nil {
			return "", err
		}
		if auth != authSuccess {
			c.disconnect()
			return "", errBadPassword
		}
	}

	err = c.send(serverdataExecCommand, cmd)
	if err != nil {
		return "", err
	}
	return c.receive()
}

func (c *Client) disconnect() error {
	if c.tcpConn != nil {
		return c.tcpConn.Close()
	}
	return nil
}

func (c *Client) connect() error {
	conn, err := net.DialTimeout(tcpNetworkName, c.address, c.timeout)
	if err != nil {
		return err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return errDialTCPFail
	}

	c.tcpConn = tcpConn
	c.tcpConn.SetDeadline(time.Now().Add(c.timeout))
	return nil
}

func (c *Client) send(cmd int, message string) error {
	if c.tcpConn == nil {
		return errNoConnection
	}

	if len(message) > maxCommandLength {
		return fmt.Errorf("message length exceed: %v/%v", len(message), maxCommandLength)
	}
	c.reqID++

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, int32(c.reqID)); err != nil {
		return err
	}
	if err := binary.Write(&buffer, binary.LittleEndian, int32(cmd)); err != nil {
		return err
	}
	buffer.WriteString(message)
	buffer.Write([]byte{'\x00', '\x00'})
	var buffer2 bytes.Buffer
	if err := binary.Write(&buffer2, binary.LittleEndian, int32(buffer.Len())); err != nil {
		return err
	}
	if _, err := buffer.WriteTo(&buffer2); err != nil {
		return err
	}
	if _, err := buffer2.WriteTo(c.tcpConn); err != nil {
		return err
	}

	return nil
}

func (c *Client) receive() (string, error) {
	if c.tcpConn == nil {
		return "", errNoConnection
	}

	responded := false
	var message bytes.Buffer
	var message2 bytes.Buffer

	// // response may be split into multiple packets, we don't know how many, so we loop until we decide to finish
	for {
		// read & parse packet length
		packetSizeBuffer := make([]byte, 4)
		if _, err := io.ReadFull(c.tcpConn, packetSizeBuffer); err != nil {
			return "", errConnectionClosed
		}
		packetSize := int32(binary.LittleEndian.Uint32(packetSizeBuffer))
		if packetSize < minMessageLength || packetSize > maxMessageLength {
			return "", fmt.Errorf("invalid packet size: %v", packetSize)
		}

		// read packet data
		packetBuffer := make([]byte, packetSize)
		if _, err := io.ReadFull(c.tcpConn, packetBuffer); err != nil {
			return "", errConnectionClosed
		}

		// parse the packet
		requestID := int32(binary.LittleEndian.Uint32(packetBuffer[0:4]))
		if requestID == -1 {
			c.disconnect()
			return "", errBadPassword
		}
		if requestID != c.reqID {
			return "", fmt.Errorf("inconsistent requestID: %v, expected: %v", requestID, c.reqID)
		}

		responded = true
		response := int32(binary.LittleEndian.Uint32(packetBuffer[4:8]))
		if response == serverdataAuthResponse {
			return authSuccess, nil
		}
		if response != serverdataResponseValue {
			return "", errInvalidResponse
		}

		// split message
		pos1 := 8
		str1 := packetBuffer[pos1:packetSize]
		for i, b := range str1 {
			if b == '\x00' {
				pos1 += i
				break
			}
		}
		pos2 := pos1 + 1
		str2 := packetBuffer[pos2:packetSize]
		for i, b := range str2 {
			if b == '\x00' {
				pos2 += i
				break
			}
		}
		if pos2+1 != int(packetSize) {
			return "", errCrapBytes
		}

		// write messages
		message.Write(packetBuffer[8:pos1])
		message2.Write(packetBuffer[pos1+1 : pos2])

		// select here

		break
	}

	if !responded {
		return "", errWaitingTimeout
	}

	if message2.Len() != 0 {
		return "", fmt.Errorf("invalid response message: %v", message2.String())
	}

	return message.String(), nil
}
