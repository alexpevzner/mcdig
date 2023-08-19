// MDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// MDNS queries

package main

import (
	"context"
	"net"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// QueryRun runs MDNS query
func QueryRun() {
	// Obtain local addresses and relevant interfaces
	addrs, if4, if6 := IfAddrs()

	for _, addr := range addrs {
		LogDebug("Using local IP address: %s@%s", addr.IP, addr.Zone)
	}

	for _, iface := range if4 {
		LogDebug("Using IPv4 interface: %s", iface.Name)
	}

	for _, iface := range if6 {
		LogDebug("Using IPv6 interface: %s", iface.Name)
	}

	// Create unicast sockets, one socket per local address
	conns := []*net.UDPConn{}

	conf := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd),
					syscall.SOL_SOCKET,
					syscall.SO_REUSEADDR, 1)
			})
			return err
		},
	}

	for _, addr := range addrs {
		conn, err := conf.ListenPacket(context.Background(),
			"udp", addr.String())

		if err != nil {
			LogFatal("%s", err)
		}

		conns = append(conns, conn.(*net.UDPConn))
	}

	// Create multicast sockets, one socket per interface
	mcast4 := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	mcast6 := &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}

	mconns := []*net.UDPConn{}
	for _, iface := range if4 {
		conn, err := net.ListenMulticastUDP("udp4", &iface, mcast4)
		if err != nil {
			LogFatal("%s", err)
		}

		mconns = append(mconns, conn)
	}

	for _, iface := range if6 {
		conn, err := net.ListenMulticastUDP("udp6", &iface, mcast6)
		if err != nil {
			LogFatal("%s", err)
		}

		mconns = append(mconns, conn)
	}

	// Start receivers
	for _, conn := range conns {
		go queryRecv(conn)
	}

	for _, conn := range mconns {
		go queryRecv(conn)
	}

	// Create DNS query message
	rq := queryNewRequest()
	rqBytes, err := rq.Pack()
	if err != nil {
		LogFatal("%s: %s", OptDomain, err)
	}

	// Begin sending queries until time is expired
	tmRxmt := time.NewTicker(OptRetransmitInterval)
	tmDone := time.NewTimer(OptQueryTime)
	done := false

	for !done {
		for _, conn := range conns {
			is4 := conn.LocalAddr().(*net.UDPAddr).IP.To4() != nil
			if is4 {
				conn.WriteToUDP(rqBytes, mcast4)
			} else {
				conn.WriteToUDP(rqBytes, mcast6)
			}
		}

		select {
		case <-tmRxmt.C:
		case <-tmDone.C:
			done = true
		}
	}
}

// queryNewQuestion creates q new request message
func queryNewRequest() *dns.Msg {
	rq := &dns.Msg{}

	// Make sure domain is FQDN
	labels, ok := dns.IsDomainName(OptDomain)
	if !ok {
		LogFatal("%s: invalid domain name")
	}

	fqdn := OptDomain
	if labels < 2 {
		fqdn += ".local."
	}

	fqdn = dns.Fqdn(fqdn)

	// Set question
	rq.Id = dns.Id()
	rq.RecursionDesired = false
	rq.Question = make([]dns.Question, 1)
	rq.Question[0] = dns.Question{fqdn, OptQType, OptQClass}

	return rq
}

// queryRecv runs on its own goroutine and receives and handles
// all UDP datagrams, received from connection
func queryRecv(conn *net.UDPConn) {
	buf := make([]byte, 65536)

	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		LogDebug("%d bytes received from %s", n, from)
	}
}