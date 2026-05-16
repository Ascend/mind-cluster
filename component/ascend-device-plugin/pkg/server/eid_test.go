package server

import (
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// makeEidWithNibble returns a 32‑hex EID string with `ch` at index `idx`.
func makeEidWithNibble(idx int, ch byte) string {
	b := []byte(strings.Repeat("0", EidLenHex))
	b[idx] = ch
	return string(b)
}

// makeEidWithIPv4 constructs a valid UBOE IPv4 EID
func makeEidWithIPv4(firstByteHex string, tail string) string {
	b := []byte(strings.Repeat("0", EidLenHex))
	copy(b[EidIdxUboeFirstByteStart:EidIdxUboeFirstByteEnd], []byte(firstByteHex))
	copy(b[EidLenHex-EidUboeIPv4TailLen:], []byte(tail))
	return string(b)
}

func makeUboeEid(dieNibble, feNibble byte) string {
	b := []byte(strings.Repeat("0", EidLenHex))
	b[EidIdxDie] = dieNibble
	b[EidIdxUboeFirstByteStart] = '0'
	b[EidIdxUboeFirstByteStart+1] = feNibble
	copy(b[EidLenHex-EidUboeIPv4TailLen:], []byte("01020304"))
	return string(b)
}

func TestGetPhyPortByEid(t *testing.T) {
	convey.Convey("test GetPhyPortByEid", t, func() {
		// base must be at least 12 chars long; 32 is safe
		base := strings.Repeat("0", 32)

		// case 1: eid[10:12] = "0a" -> 0x0a -> low6 bits == 10
		eid := []byte(base)
		eid[EidIdxPortStart] = '0'
		eid[EidIdxPortStart+1] = 'a'
		convey.So(GetPhyPortByEid(string(eid)), convey.ShouldEqual, 10)

		// case 2: eid[10:12] = "ff" -> 0xff & 0x3f == 0x3f -> 63
		eidFF := []byte(base)
		eidFF[EidIdxPortStart] = 'f'
		eidFF[EidIdxPortStart+1] = 'f'
		convey.So(GetPhyPortByEid(string(eidFF)), convey.ShouldEqual, 63)

		// case 3: invalid hex char in the two-char field should return -1
		eidInvalid := []byte(base)
		eidInvalid[EidIdxPortStart] = 'Z'
		eidInvalid[EidIdxPortStart+1] = '0'
		convey.So(GetPhyPortByEid(string(eidInvalid)), convey.ShouldEqual, -1)
	})
}

func TestGetPgEid(t *testing.T) {
	convey.Convey("test GetPgEid", t, func() {
		// base must be at least EidIdxPortEnd long; 32 is safe
		base := strings.Repeat("0", 32)

		// normal eid: eid[10:12] = "01" (not PG)
		normal := []byte(base)
		normal[EidIdxPortStart] = '0'
		normal[EidIdxPortStart+1] = '1'

		// pg eid: eid[10:12] = "3f" -> 0x3f -> matches pgPhyValue (63)
		pg := []byte(base)
		pg[EidIdxPortStart] = '3'
		pg[EidIdxPortStart+1] = 'f'

		found, err := GetPgEid([]string{string(normal), string(pg)})
		convey.So(err, convey.ShouldBeNil)
		convey.So(found, convey.ShouldEqual, string(pg))

		// not found case
		_, err = GetPgEid([]string{string(normal)})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestIsPgEid(t *testing.T) {
	convey.Convey("test IsPgEid", t, func() {
		base := strings.Repeat("0", 32)

		// pg case: "3f" -> should be true
		pg := []byte(base)
		pg[EidIdxPortStart] = '3'
		pg[EidIdxPortStart+1] = 'f'
		convey.So(IsPgEid(string(pg)), convey.ShouldBeTrue)

		// non-pg case: "01" -> should be false
		non := []byte(base)
		non[EidIdxPortStart] = '0'
		non[EidIdxPortStart+1] = '1'
		convey.So(IsPgEid(string(non)), convey.ShouldBeFalse)

		// invalid hex should be false (GetPhyPortByEid returns -1)
		invalid := []byte(base)
		invalid[EidIdxPortStart] = 'Z'
		invalid[EidIdxPortStart+1] = '0'
		convey.So(IsPgEid(string(invalid)), convey.ShouldBeFalse)
	})
}


func TestGetDieIdByEid(t *testing.T) {
	convey.Convey("test GetDieIdByEid", t, func() {
		eid1 := makeEidWithNibble(EidIdxDie, '4') // 0100b → bit2=1
		convey.So(GetDieIdByEid(eid1), convey.ShouldEqual, 1)

		eid0 := makeEidWithNibble(EidIdxDie, '0')
		convey.So(GetDieIdByEid(eid0), convey.ShouldEqual, 0)

		eidInvalid := makeEidWithNibble(EidIdxDie, 'Z')
		convey.So(GetDieIdByEid(eidInvalid), convey.ShouldEqual, -1)
	})
}

func setEidChar(eid string, idx int, ch byte) string {
	b := []byte(eid)
	if idx < 0 || idx >= len(b) {
		return eid
	}
	b[idx] = ch
	return string(b)
}

func TestGetFeIdByEid(t *testing.T) {
	convey.Convey("test GetFeIdByEid", t, func() {
		base := makeEidWithNibble(EidIdxFeStart, '0')
		eid := setEidChar(base, EidIdxFeStart, '0')
		eid = setEidChar(eid, EidIdxFeStart+1, 'a')

		convey.So(GetFeIdByEid(eid), convey.ShouldEqual, 10)

		base2 := makeEidWithNibble(EidIdxFeStart, '0')
		eidInvalid := setEidChar(base2, EidIdxFeStart, 'Z')
		eidInvalid = setEidChar(eidInvalid, EidIdxFeStart+1, '0')
		convey.So(GetFeIdByEid(eidInvalid), convey.ShouldEqual, -1)
	})
}

func TestIsUboeEid(t *testing.T) {
	convey.Convey("test IsUboeEid", t, func() {
		base := strings.Repeat("0", 40)

		// case: set first byte [12:14] = "08" (first byte 0x08), and set eid[14] = 'c' (0xC -> bits3..2 == 11)
		e1 := []byte(base)
		e1[EidIdxUboeFirstByteStart] = '0'
		e1[EidIdxUboeFirstByteStart+1] = '8'
		// IsUboeEid reads eid[14], set it to 'c' to make (0xC & 0xC) == 0xC
		e1[EidIdxUboeFirstByteEnd] = 'c'
		convey.So(IsUboeEid(string(e1)), convey.ShouldBeTrue)

		// case: first byte same but eid[14] = '0' -> not UBOE
		e0 := []byte(base)
		e0[EidIdxUboeFirstByteStart] = '0'
		e0[EidIdxUboeFirstByteStart+1] = '8'
		e0[EidIdxUboeFirstByteEnd] = '0'
		convey.So(IsUboeEid(string(e0)), convey.ShouldBeFalse)

		// invalid hex at eid[14] -> parse error -> false
		einv := []byte(base)
		einv[EidIdxUboeFirstByteStart] = '0'
		einv[EidIdxUboeFirstByteStart+1] = '8'
		einv[EidIdxUboeFirstByteEnd] = 'Z'
		convey.So(IsUboeEid(string(einv)), convey.ShouldBeFalse)
	})
}

func TestGetUboeIPv4ByEid(t *testing.T) {
	convey.Convey("test GetUboeIPv4ByEid", t, func() {
		eid := makeEidWithIPv4("0a", "01020304")
		ip, err := GetUboeIPv4ByEid(eid)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, "10.1.2.3")

		eidLL := makeEidWithIPv4("fe", "01020304")
		_, err = GetUboeIPv4ByEid(eidLL)
		convey.So(err, convey.ShouldNotBeNil)

		eidInvalid := makeEidWithIPv4("0g", "01020304")
		_, err = GetUboeIPv4ByEid(eidInvalid)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestUrmaDeviceMethods(t *testing.T) {
	convey.Convey("test UrmaDevice methods", t, func() {
		base := strings.Repeat("0", 40)
		eid := []byte(base)

		// die nibble at EidIdxDie = '4' -> die bits -> 1
		eid[EidIdxDie] = '4'

		// set first byte [12:14] = "08" so GetUboeIPv4 will see b0 = 0x08 (8)
		eid[EidIdxUboeFirstByteStart] = '0'
		eid[EidIdxUboeFirstByteStart+1] = '8'

		// set eid[14] = 'c' so IsUboeEid returns true
		eid[EidIdxUboeFirstByteEnd] = 'c'

		// set UBOE tail ip (last 8 hex chars) to "01020304" -> ip bytes 1,2,3
		copy(eid[len(eid)-EidUboeIPv4TailLen:], "01020304")

		u := &UrmaDevice{EidList: []string{string(eid)}}

		convey.So(u.GetDieId(), convey.ShouldEqual, 1)
		// FE parsed from [12:14] == "08" -> 0x08 == 8
		convey.So(u.GetFeId(), convey.ShouldEqual, 8)
		convey.So(u.IsUboe(), convey.ShouldBeTrue)

		ip, err := u.GetUboeIPv4()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, "8.1.2.3")
	})

	convey.Convey("empty EidList", t, func() {
		u := &UrmaDevice{}
		convey.So(u.GetDieId(), convey.ShouldEqual, -1)
		convey.So(u.GetFeId(), convey.ShouldEqual, -1)
		convey.So(u.IsUboe(), convey.ShouldBeFalse)
		_, err := u.GetUboeIPv4()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestRawBytesToEidString(t *testing.T) {
	convey.Convey("test RawBytesToEidString", t, func() {
		raw := []byte{0xAA, 0xBB, 0xCC}
		convey.So(RawBytesToEidString(raw), convey.ShouldEqual, "aabbcc")
	})
}

func TestParseUrmaDevices(t *testing.T) {
	convey.Convey("test ParseUrmaDevices", t, func() {
		base := strings.Repeat("0", 40)

		// construct uboe EID: first byte [12:14] = "08", eid[14] = 'c', tail ip "08010203"
		uboe := []byte(base)
		uboe[EidIdxDie] = '4' // die nibble -> die=1
		uboe[EidIdxUboeFirstByteStart] = '0'
		uboe[EidIdxUboeFirstByteStart+1] = '8'
		uboe[EidIdxUboeFirstByteEnd] = 'c' // IsUboeEid reads this position
		copy(uboe[len(uboe)-EidUboeIPv4TailLen:], "08010203") // first byte 0x08 -> ip 8.1.2.3

		// construct pg EID inline: set eid[10:12] = "3f"
		pg := []byte(base)
		pg[EidIdxPortStart] = '3'
		pg[EidIdxPortStart+1] = 'f'

		u := &UrmaDevice{EidList: []string{string(uboe), string(pg)}}
		parsed := ParseUrmaDevices([]*UrmaDevice{u})

		convey.So(len(parsed), convey.ShouldEqual, 1)
		p := parsed[0]

		convey.So(p.Die, convey.ShouldEqual, 1)
		// FE parsed from [12:14] == "08" -> 8
		convey.So(p.Fe, convey.ShouldEqual, 8)
		convey.So(p.IsUboe, convey.ShouldBeTrue)
		convey.So(p.PgEid, convey.ShouldEqual, string(pg))
		convey.So(len(p.Eids), convey.ShouldEqual, 2)
	})

	convey.Convey("nil device skipped", t, func() {
		parsed := ParseUrmaDevices([]*UrmaDevice{nil})
		convey.So(len(parsed), convey.ShouldEqual, 0)
	})
}
