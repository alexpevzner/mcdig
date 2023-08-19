// MDIG - DIG for MDNS (Multicast DNS lookup utility)
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

// AnswerInput handles received messages
func ResponseInput(rsp *dns.Msg) {
	// We can be called from different goroutines, so
	// locking is necessary
	rspLock.Lock()
	defer rspLock.Unlock()

	// Save RRs, deduplicate
	rspAnswer = dns.Dedup(append(rspAnswer, rsp.Answer...), nil)
	rspAuthority = dns.Dedup(append(rspAnswer, rsp.Ns...), nil)
	rspAdditional = dns.Dedup(append(rspAnswer, rsp.Extra...), nil)
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
// The returned error, if any, comes from w.Write()
func ResponsePrint(w io.Writer, ans, auth, add []dns.RR) error {
	buf := bytes.Buffer{}

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
func ResponseGetAndPrint(w io.Writer) error {
	ans, auth, add := ResponseGet()
	return ResponsePrint(w, ans, auth, add)
}
