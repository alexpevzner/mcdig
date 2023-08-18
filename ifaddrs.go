// MDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Obtaining list of local addresses

package main

import "net"

// IfAddrs returns a slice of local (source) addresses for MDNS
// queries
//
// It honors the following options:
//   * OptIface
//   * Opt4
//   * Opt6
func IfAddrs() []*net.UDPAddr {
	// Obtain list of network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		LogFatal("Can't get list of network interfaces: %s", err)
	}

	// Apply OptIface option, if set
	if OptIface != "" {
		for len(interfaces) != 0 {
			if interfaces[0].Name == OptIface {
				interfaces = interfaces[:1]
				break
			} else {
				interfaces = interfaces[1:]
			}
		}

		if len(interfaces) == 0 {
			LogFatal("Unknown network interface: %q", OptIface)
		}
	}

	// Build list of addresses
	addrs := []*net.UDPAddr{}
	for _, iface := range interfaces {
		ifaddrs, err := iface.Addrs()
		if err != nil {
			LogFatal("%s: can't get interface addresses: %s", iface.Name, err)
		}

		for _, ifaddr := range ifaddrs {
			ip := ifaddr.(*net.IPNet).IP
			ip4 := ip.To4()

			switch {
			case ip.IsLoopback():
				// Loopback addresses cannot be used for MDNS
				ip = nil

			case ip4 != nil:
				// All IPv4 addresses are OK
				ip = ip4

			case !ip.IsLinkLocalUnicast():
				// Only link-local IPv6 addresses are OK
				ip = nil
			}

			// Honor IPv4/IPv6 options
			if ip != nil {
				if ip4 != nil && !Opt4 {
					ip = nil
				} else if ip4 == nil && !Opt6 {
					ip = nil
				}

			}

			// Add address to the list
			if ip != nil {
				LogDebug("Using local IP address: %s", ip)
				addr := &net.UDPAddr{
					IP:   ip,
					Zone: iface.Name,
				}
				addrs = append(addrs, addr)
			}
		}
	}

	// List must be non-empty
	if len(addrs) == 0 {
		LogFatal("No local IP addresses found")
	}

	return addrs
}
