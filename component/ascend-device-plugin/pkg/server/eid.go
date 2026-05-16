/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	// EidLenHex Total length of an EID in hex characters.
	EidLenHex = 32
	// EidIdxDie Die ID nibble: eid[10]
	EidIdxDie = 10
	// EidIdxFeStart FE ID nibble: eid[12:14]
	EidIdxFeStart = 12
	// EidIdxFeEnd FE ID nibble: eid[12:14]
	EidIdxFeEnd = 14
	// EidIdxPortStart PhyPortID nibble: eid[10:12]
	EidIdxPortStart = 10
	// EidIdxPortEnd PhyPortID nibble: eid[10:12]
	EidIdxPortEnd   = 12
	// EidIdxUboeFirstByteStart UBOE IPv4 first byte: eid[12:14]
	EidIdxUboeFirstByteStart = 12
	// EidIdxUboeFirstByteEnd UBOE IPv4 first byte: eid[12:14]
	EidIdxUboeFirstByteEnd   = 14
	// EidUboeIPv4TailLen UBOE IPv4 tail: last 8 hex chars (eid[24:32]) → 4 bytes
	EidUboeIPv4TailLen = 8
)

const (
	// DieIdBitMask Bit mask for die ID: extract bit3 and bit2 from eid[EidIdxDie]
	DieIdBitMask  = 0xC
	// DieIdBitShift Bit mask for die ID: extract bit3 and bit2 from eid[EidIdxDie]
	DieIdBitShift = 2
	// UboeBitMask Bit mask for UBOE flag: check bit3 + bit2 of eid[EidIdxFe]
	UboeBitMask = 0xC
	// FeMask mask for low 7 bits (0b01111111)
	FeMask = 0x7F
	// phyPortMask mask for low 6 bits 0b00111111
    phyPortMask = 0x3F
	// HexBase Hex parsing parameters
	HexBase   = 16
	// IntBitSize Hex parsing parameters
	IntBitSize = 64
)

// Offsets inside the 8‑hex UBOE IPv4 tail (ipHex).
const (
	// UboeIPv4Byte0Start ipHex[0:2]
	UboeIPv4Byte0Start = 0
	// UboeIPv4Byte1Start ipHex[2:4]
	UboeIPv4Byte1Start = 2
	// UboeIPv4Byte2Start ipHex[4:6]
	UboeIPv4Byte2Start = 4
	// UboeIPv4ByteStep each byte is 2 hex chars
	UboeIPv4ByteStep   = 2
)


type UrmaDevice struct {
	EidList []string
}

// GetPhyPortByEid eid[10:12], returns the lower 6 bits (bit0..bit5)
func GetPhyPortByEid(eid string) int {
	s := eid[EidIdxPortStart:EidIdxPortEnd]
	v, err := strconv.ParseInt(s, HexBase, IntBitSize)
	if err != nil {
		return -1
	}

	return int(v & int64(phyPortMask))
}

// GetPgEid PG EID：eid[10:12] lower 6 bits (bit0..bit5) == '3f'
func GetPgEid(eids []string) (string, error) {
	for _, eid := range eids {
		if GetPhyPortByEid(eid) == phyPortMask {
			return eid, nil
		}
	}
	return "", fmt.Errorf("pg eid not found")
}

// IsPgEid PG EID：eid[11] == 'f'
func IsPgEid(eid string) bool {
	return GetPhyPortByEid(eid) == phyPortMask
}

// GetDieIdByEid die_id：eid[10] Extract the 2nd bit after converting the value to binary.
func GetDieIdByEid(eid string) int {
	ch := eid[EidIdxDie:EidIdxDie+1]
	v, err := strconv.ParseInt(ch, HexBase, IntBitSize)
	if err != nil {
		return -1
	}
	return int((v & int64(DieIdBitMask)) >> DieIdBitShift)
}

// GetFeIdByEid returns the lower 7 bits of the two-hex at eid[12:14].
func GetFeIdByEid(eid string) int {
	s := eid[EidIdxFeStart:EidIdxFeEnd]
	v, err := strconv.ParseInt(s, HexBase, IntBitSize)
	if err != nil {
		return -1
	}
	return int(v & int64(FeMask))
}

// IsUboeEid determines whether the EID represents a UBOE device by checking bit3 of eid[14] bit3 and bit2
func IsUboeEid(eid string) bool {
	ch := eid[EidIdxUboeFirstByteEnd:EidIdxUboeFirstByteEnd+1]
	v, err := strconv.ParseInt(ch, HexBase, IntBitSize)
	if err != nil {
		return false
	}
	// bit3 and bit2 -> 0b1100
	return (v & UboeBitMask) == UboeBitMask
}

// GetUboeIPv4ByEid UBOE IP: eid[12:14].eid[24:26].eid[26:28].eid[28:30]
func GetUboeIPv4ByEid(eid string) (string, error) {
	ipHex := eid[len(eid)-EidUboeIPv4TailLen:]
	firstByteHex := eid[EidIdxUboeFirstByteStart:EidIdxUboeFirstByteEnd]

	b0, err0 := strconv.ParseInt(firstByteHex, HexBase, IntBitSize)
	b1, err1 := strconv.ParseInt(ipHex[UboeIPv4Byte0Start:UboeIPv4Byte0Start+UboeIPv4ByteStep], HexBase, IntBitSize)
	b2, err2 := strconv.ParseInt(ipHex[UboeIPv4Byte1Start:UboeIPv4Byte1Start+UboeIPv4ByteStep], HexBase, IntBitSize)
	b3, err3 := strconv.ParseInt(ipHex[UboeIPv4Byte2Start:UboeIPv4Byte2Start+UboeIPv4ByteStep], HexBase, IntBitSize)
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("parse ip failed")
	}
	// if b0 is 0xfe (link-local in new requirement), skip
	if b0 == 0xfe {
		return "", fmt.Errorf("link-local ip, skip")
	}

	return fmt.Sprintf("%d.%d.%d.%d", b0, b1, b2, b3), nil
}

func (u *UrmaDevice) GetFeId() int {
	if len(u.EidList) == 0 {
		return -1
	}
	return GetFeIdByEid(u.EidList[0])
}

func (u *UrmaDevice) GetDieId() int {
	if len(u.EidList) == 0 {
		return -1
	}
	return GetDieIdByEid(u.EidList[0])
}

func (u *UrmaDevice) GetPgEid() (string, error) {
	return GetPgEid(u.EidList)
}

func (u *UrmaDevice) IsUboe() bool {
	if len(u.EidList) == 0 {
		return false
	}
	return IsUboeEid(u.EidList[0])
}

func (u *UrmaDevice) GetUboeIPv4() (string, error) {
	for _, eid := range u.EidList {
		ip, err := GetUboeIPv4ByEid(eid)
		if err == nil {
			return ip, nil
		}
	}
	return "", fmt.Errorf("no valid uboe ip found")
}

func (u *UrmaDevice) GetDieIdByEid(eid string) int {
	return GetDieIdByEid(eid)
}

func (u *UrmaDevice) GetPhyPortByEid(eid string) int {
	return GetPhyPortByEid(eid)
}

// RawBytesToEidString converts raw bytes into a lowercase 32‑character EID hex string.
func RawBytesToEidString(raw []byte) string {
	return strings.ToLower(hex.EncodeToString(raw))
}

// ParsedEid represents parsed information of a single EID
type ParsedEid struct {
	Eid     string
	Die     int
	Fe      int
	Port    int
	IsPg    bool
}

// ParsedUrma represents parsed information of a URMA device
type ParsedUrma struct {
	Raw       *UrmaDevice
	Eids      []ParsedEid
	PgEid     string
	Die       int
	Fe        int
	IsUboe    bool
	IPv4      string
}

// ParseUrmaDevices parses a list of URMA devices into structured ParsedUrma objects.
func ParseUrmaDevices(list []*UrmaDevice) []*ParsedUrma {
	result := make([]*ParsedUrma, 0)

	for _, u := range list {
		if u == nil {
			continue
		}
		p := &ParsedUrma{
			Raw:  u,
			Eids: make([]ParsedEid, 0),
		}
		// Parse URMA-level attributes: die / fe / uboe
		p.Die = u.GetDieId()
		p.Fe = u.GetFeId()
		p.IsUboe = u.IsUboe()
		// Parse all EIDs of this URMA device
		for _, eid := range u.EidList {
			pe := ParsedEid{
				Eid:  eid,
				Die:  GetDieIdByEid(eid),
				Fe:   GetFeIdByEid(eid),
				Port: GetPhyPortByEid(eid),
				IsPg: IsPgEid(eid),
			}
			if pe.IsPg {
				p.PgEid = eid
			}
			p.Eids = append(p.Eids, pe)
		}

		// If this is a UBOE device, parse its encoded IPv4 address
		if p.IsUboe {
			ip, _ := u.GetUboeIPv4()
			p.IPv4 = ip
		}
		result = append(result, p)
	}
	return result
}
