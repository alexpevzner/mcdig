// MCDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// MDNS responses processing

package main

import (
	"bytes"
	"io"
	"sync"

	"github.com/miekg/dns"
)

var (
	rspAnswer     []dns.RR   // Collected answer section
	rspAuthority  []dns.RR   // Collected authority section
	rspAdditional []dns.RR   // Collected additional section
	rspLock       sync.Mutex // Access lock
)

// ResponseInput handles received messages
func ResponseInput(rsp *dns.Msg) {
	// We can be called from different goroutines, so
	// locking is necessary
	rspLock.Lock()
	defer rspLock.Unlock()

	// Save RRs, deduplicate
	rspAnswer = responseAppend(rspAnswer, rsp.Answer)
	rspAuthority = responseAppend(rspAnswer, rsp.Ns)
	rspAdditional = responseAppend(rspAdditional, rsp.Extra)
}

// responseAppend appends newly received response data to the
// section, removes duplicates and returns updated section
func responseAppend(section, data []dns.RR) []dns.RR {
	for _, rr := range data {
		// Skip OPT PSEUDOSECTION records
		//
		// These records are messy on output and doesn't contain
		// any information we need so far
		//
		// TODO: if eventually we will want to include these
		// records, their formatting needs to be revisited
		//
		// (*dns.OPT) String() loogs to be unreliable for
		// these records:
		//
		// --------------------------------------------------
		// ;; OPT PSEUDOSECTION:
		// ; EDNS: version 0; flags:; MBZ: 0x1194, udp: 1440
		// ; ESU: �5�K
		// /�5�K
		// /
		// --------------------------------------------------
		//
		// Please, notice the trailing junk
		if _, ok := rr.(*dns.OPT); ok {
			continue
		}

		// mDNS reuses upper bit of RR class as "unicast response"
		// flag - so we must clear it before data is saved into
		// our records
		rr2 := dns.Copy(rr)
		rr2.Header().Class &^= 1 << 15

		section = append(section, rr2)
	}
	return dns.Dedup(section, nil)
}

// ResponseGet returns responses, collected so far
func ResponseGet() (ans, auth, add []dns.RR) {
	// Acquire the lock
	rspLock.Lock()
	defer rspLock.Unlock()

	// Create copies
	ans = make([]dns.RR, len(rspAnswer))
	copy(ans, rspAnswer)

	auth = make([]dns.RR, len(rspAuthority))
	copy(auth, rspAuthority)

	add = make([]dns.RR, len(rspAdditional))
	copy(add, rspAdditional)

	return
}

// ResponsePrint prints responses into io.Writer
// If question is not nil, it is assumed to be msg.Question
// and used to format QUESTION PSEUDOSECTION (normally
// missed in the MDNS queries
//
// The returned error, if any, comes from w.Write()
func ResponsePrint(w io.Writer, question []dns.Question,
	ans, auth, add []dns.RR) error {
	buf := bytes.Buffer{}

	// QUESTION PSEUDOSECTION
	if question != nil {
		buf.WriteString(";; QUESTION PSEUDOSECTION:\n")
		for _, q := range question {
			buf.WriteString(q.String())
			buf.WriteByte('\n')
		}

		buf.WriteByte('\n')
	}

	// ANSWER SECTION
	if ans != nil {
		buf.WriteString(";; ANSWER SECTION:\n")
		for _, rr := range ans {
			buf.WriteString(rr.String())
			buf.WriteByte('\n')
		}

		buf.WriteByte('\n')
	}

	// AUTHORITY SECTION
	if auth != nil {
		buf.WriteString(";; AUTHORITY SECTION:\n")
		for _, rr := range auth {
			buf.WriteString(rr.String())
			buf.WriteByte('\n')
		}

		buf.WriteByte('\n')
	}

	// ADDITIONAL SECTION
	if add != nil {
		buf.WriteString(";; ADDITIONAL SECTION:\n")
		for _, rr := range add {
			buf.WriteString(rr.String())
			buf.WriteByte('\n')
		}

		buf.WriteByte('\n')
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// ResponseGetAndPrint is the convenience wrapper for
// ResponseGet + ResponsePrint
func ResponseGetAndPrint(w io.Writer, question []dns.Question) error {
	ans, auth, add := ResponseGet()
	return ResponsePrint(w, question, ans, auth, add)
}
