// MDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Functions for IP addresses
// Check for local IP addresses

package main

import "net"

// IsLocalAddr tells if IP address is local (i.e., belongs to
// one of local interfaces)
func AddrIsLocal(addr net.IP) bool {
	// Loopback address is definitely local
	if addr.IsLoopback() {
		return true
	}

	// Well. It can be optimized. But lets do the simple way
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false // Can't tell, lets make a safe assumption
	}

	for _, a := range addrs {
		if local, ok := a.(*net.IPNet); ok {
			if addr.Equal(local.IP) {
				return true
			}
		}
	}

	return false
}

// IsLocalAddr tells if UDP address is local (i.e., belongs to
// one of local interfaces)
func AddrIsLocalUDP(addr *net.UDPAddr) bool {
	return AddrIsLocal(addr.IP)
}
