// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package net_util

/*
   net_util.go - convenience methods for networking
*/

import (
	"net"
	"net/http"
	"time"
	)


/*
Returns an http.Client whose Transport uses the net.DialTimeout function rather than
straight net.Dial to initiate connections. The dialing function will time out the
connection establishment attempt after the specified number of seconds.

Usage:

client := HttpTimeoutClient(20)
resp, err := client.Get("http://example.com")  // Times out if can't establish connection in 20s

*/
func HttpTimeoutClient(requestTimeoutSeconds int) *http.Client {
   timeoutDuration := time.Duration(requestTimeoutSeconds) * time.Second	
   timeoutDial := func(network, addr string) (net.Conn, error) {
   	  return net.DialTimeout(network, addr, timeoutDuration)
   }
   transport := &http.Transport{Proxy: http.ProxyFromEnvironment, 
   	                            Dial: timeoutDial,
                               }
   return &http.Client{Transport: transport}                            
}