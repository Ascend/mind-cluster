/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package utils offer the some utils for certificate handling
package utils

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	dnsReqTimeoutForCheck = time.Second
	resolveNetwork        = "ip"
	domainReg             = "^[a-zA-Z0-9][a-zA-Z0-9.-]{1,256}[a-zA-Z0-9]$"
)

// ClientIP try to get the clientIP
func ClientIP(r *http.Request) string {
	// get forward ip fistly
	var ip string
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	forwardSlice := strings.Split(xForwardedFor, ",")
	if len(forwardSlice) >= 1 {
		if ip = strings.TrimSpace(forwardSlice[0]); ip != "" {
			return ip
		}
	}
	// try get ip from "X-Real-Ip"
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	var err error
	if ip, _, err = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// CheckDomain check domain which by regex and blacklist
// Note1: If parameter 'forLocalUsage' is true, which indicate url in this check act is used by the local, the checker
// return s error when any hostname in url is equivalent to localhost.
// Warning: this func may call DNS (configured in file /etc/resolv.conf) by UDP (port: 53).
// !! Make sure this net chain is added in Communication Matrix !!
// Note 2: When a new domain name is configured, the IP address corresponding to the domain name cannot be resolved, so
// the parsing error can be ignored. If the domain is used for configuration, the 'ignoreLookupIPErr' value can be true.
// If the domain is used for usage, the value can be false.
func CheckDomain(domain string, forLocalUsage bool, ignoreLookupIPErr bool) error {
	matched, err := regexp.MatchString(domainReg, domain)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("domain does not match allowed regex")
	}
	if !forLocalUsage {
		return nil
	}
	if IsDigitString(domain) {
		return errors.New("domain can not be all digits")
	}
	if strings.Contains(domain, "localhost") {
		return errors.New("domain can not contain localhost")
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), dnsReqTimeoutForCheck)
	defer cancelFunc()
	ips, err := net.DefaultResolver.LookupNetIP(ctx, resolveNetwork, domain)
	if err != nil {
		// When a new domain name is configured, the IP address corresponding to the domain name cannot be resolved, so
		// the parsing error can be ignored.
		if ignoreLookupIPErr {
			return nil
		}
		return errors.New("domain resolve failed")
	}
	for _, ip := range ips {
		parsedIP := net.ParseIP(ip.String())
		if parsedIP != nil && parsedIP.IsLoopback() {
			return errors.New("domain is not allowed to be a loop back address")
		}
	}
	return nil
}

// IsHostValid check if the host is valid
func IsHostValid(host string) error {
	parsedIp := net.ParseIP(host)
	if parsedIp != nil {
		return IsIPValid(parsedIp)
	}
	return CheckDomain(host, false, true)
}

// IsIPValid check ip valid
func IsIPValid(parsedIp net.IP) error {
	if parsedIp == nil {
		return errors.New("parse ip is nil")
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		return errors.New("not a valid ipv4 or ipv6 ip")
	}
	if parsedIp.IsUnspecified() {
		return errors.New("is all zeros ip")
	}
	if parsedIp.IsMulticast() {
		return errors.New("is multicast ip")
	}
	return nil
}
