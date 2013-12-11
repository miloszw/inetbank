package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"bytes"
)

const (
	port = "1337"
	timeout = 30 // in seconds

	// Control bits
	Failure = 0
	Success = 1
	Request = 2
)

type ClientConn struct {
	conn net.Conn
}

func main() {
	// start the server
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bank started listening on port " + port)

	// shut down server before exiting
	defer l.Close()

	// loop forever
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Connection from %s\n", conn.LocalAddr())

		clc := new(ClientConn)
		clc.conn = conn

		// handle conn in a new goroutine
		go func(cl *ClientConn) {
			c := cl.conn
			defer c.Close()

			// create request buffer, max 10 bytes
			request := make([]byte, 10)
			for {
				// set idle timeout
				c.SetDeadline(time.Now().Add(timeout*time.Second))

				// read
				_, err := c.Read(request)

				if err != nil {
					// check for Timeout (err.(net.Error) => type assertion)
					if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
						log.Printf("Connection from %s timed out.\n", conn.LocalAddr())
					} else {
						log.Printf("Connection from %s closed remotely.\n", conn.LocalAddr())
					}
					return // exit goroutine and close conn
				}

				// process request
				log.Printf("Request from %s: %s\n", conn.LocalAddr(), string(request))
				perr := cl.processRequest(request)
				if perr != nil {
					fmt.Println(perr)
				}
			}
		}(clc)
	}
}

func (c *ClientConn) processRequest(req []byte) error {
	resp := make([]byte, 10)
	// trim trailing 0's
	req = bytes.TrimRight(req, string([]byte{0}))
	// conv to string
	sreq := string(req)
	
	switch sreq {

	// Get welcome message
	case "get_wmsg":
		// Get client language
		resp = []byte("get_lang")
		resp = bytes.Join([][]byte{{Request},resp}, []byte{})
		c.conn.Write(resp)
		// clear resp between wr/rd
		// and set limit to 80 bytes
		resp = make([]byte, 80)
		c.conn.Read(resp)
		lang := string(resp)

		welcome_message := "Temp welcome message in " + lang
		resp = []byte(welcome_message)
		resp = bytes.Join([][]byte{{Success},resp}, []byte{})
		c.conn.Write(resp)
	default:
		return fmt.Errorf("Request fallthrough (bad request: %s)", sreq)
	}
	return nil
}