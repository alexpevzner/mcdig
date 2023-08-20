// MCDIG - DIG for MDNS (Multicast DNS lookup utility)
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
//   - OptIface
//   - Opt4
//   - Opt6
//
// It returns (filtered, according to options) list of addresses
// and two lists of network interfaces: one for IPv4 and one for
// IPv6. Note, interfaces are only included into the list if they
// are really in use, after address filtering
func IfAddrs() (addrs []*net.UDPAddr, if4, if6 []net.Interface) {
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

	// Build list of addresses and interfaces
	addrs = []*net.UDPAddr{}
	if4seen := make(map[int]bool)
	if6seen := make(map[int]bool)

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

			// Add address and interface to the list
			if ip != nil {
				addr := &net.UDPAddr{
					IP:   ip,
					Port: 5353,
					Zone: iface.Name,
				}
				addrs = append(addrs, addr)

				switch {
				case ip4 != nil && !if4seen[iface.Index]:
					if4 = append(if4, iface)
					if4seen[iface.Index] = true

				case ip4 == nil && !if6seen[iface.Index]:
					if6 = append(if6, iface)
					if6seen[iface.Index] = true
				}
			}
		}
	}

	// List must be non-empty
	if len(addrs) == 0 {
		LogFatal("No local IP addresses found")
	}

	return addrs, if4, if6
}
