package dynu

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/stretchr/testify/assert"
)

var zone = os.Getenv("TEST_ZONE")
var apiToken = os.Getenv("TEST_API_TOKEN")
var testRealApi = zone != "" && apiToken != ""

var domain = "dynu.com"
var ownDomain = "my.dynu.com"

var dummyDynuProviderData = DynuProviderData{
	ID:       123,
	DomainID: 456,
}

func checkSkipApiTest(t *testing.T) {
	if !testRealApi {
		t.Skip("Env variables not set. Skipping api test.")
	}
}

func TestGetRecords(t *testing.T) {
	checkSkipApiTest(t)

	ctx := context.TODO()

	provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}

	recs, err := provider.GetRecords(ctx, zone)

	if !assert.NoError(t, err) {
		return
	}

	for _, rec := range recs {
		t.Log(rec)
	}
}

func TestAddAndDeleteTxtRecord(t *testing.T) {
	checkSkipApiTest(t)

	useRecordIdCases := []bool{true, false}
	for _, useRecordId := range useRecordIdCases {
		t.Run(fmt.Sprintf("use record id:%t", useRecordId), func(t *testing.T) {
			ctx := context.TODO()

			provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}
			testRecord := libdns.TXT{
				Name: "@",
				Text: "TEST TXT RECORD",
				TTL:  time.Duration(120) * time.Second,
			}

			t.Log(testRecord)

			addedRecords, err := provider.AppendRecords(ctx, zone, []libdns.Record{testRecord})
			if !assert.NoError(t, err) {
				return
			}

			for _, rec := range addedRecords {
				t.Log(rec)
			}

			for _, rec := range addedRecords {
				assert.Equal(t, testRecord.RR().Type, rec.RR().Type)
				assert.Equal(t, testRecord.RR().Name, rec.RR().Name)
				assert.Equal(t, testRecord.RR().Data, rec.RR().Data)
				assert.Equal(t, testRecord.RR().TTL, rec.RR().TTL)
			}

			if !useRecordId {
				stripRecordIds(addedRecords)
			}

			for _, rec := range addedRecords {
				t.Log(rec)
			}

			deletedRecords, err := provider.DeleteRecords(ctx, zone, addedRecords)
			if !assert.NoError(t, err) {
				return
			}

			for _, rec := range deletedRecords {
				t.Log(rec)
			}
		})
	}
}

func TestAddUpdateAndDeleteTxtRecord(t *testing.T) {
	checkSkipApiTest(t)

	useRecordIdCases := []bool{true, false}
	for _, useRecordId := range useRecordIdCases {
		t.Run(fmt.Sprintf("use record id:%t", useRecordId), func(t *testing.T) {
			ctx := context.TODO()

			provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}
			var testRecord libdns.Record = libdns.TXT{
				Name: "test",
				Text: "TEST TXT RECORD",
				TTL:  time.Duration(120) * time.Second,
			}

			t.Log(testRecord)

			addedRecords, err := provider.AppendRecords(ctx, zone, []libdns.Record{testRecord})
			if !assert.NoError(t, err) {
				return
			}

			for _, rec := range addedRecords {
				t.Log(rec)
			}

			for _, rec := range addedRecords {
				assert.Equal(t, testRecord.RR().Type, rec.RR().Type)
				assert.Equal(t, testRecord.RR().Name, rec.RR().Name)
				assert.Equal(t, testRecord.RR().Data, rec.RR().Data)
				assert.Equal(t, testRecord.RR().TTL, rec.RR().TTL)
			}

			testRecord = addedRecords[0]
			if txtRecord, ok := testRecord.(libdns.TXT); ok {
				txtRecord.Text = "TEST UPDATED TXT RECORD"
				testRecord = txtRecord
			}

			if !useRecordId {
				testRecord = stripRecordId(testRecord)
			}

			t.Log(testRecord)

			addedRecords, err = provider.SetRecords(ctx, zone, []libdns.Record{testRecord})
			if !assert.NoError(t, err) {
				return
			}

			for _, rec := range addedRecords {
				t.Log(rec)
			}

			for _, rec := range addedRecords {
				assert.Equal(t, testRecord.RR().Type, rec.RR().Type)
				assert.Equal(t, testRecord.RR().Name, rec.RR().Name)
				assert.Equal(t, testRecord.RR().Data, rec.RR().Data)
				assert.Equal(t, testRecord.RR().TTL, rec.RR().TTL)
			}

			if !useRecordId {
				stripRecordIds(addedRecords)
			}

			for _, rec := range addedRecords {
				t.Log(rec)
			}

			deletedRecords, err := provider.DeleteRecords(ctx, zone, addedRecords)
			if !assert.NoError(t, err) {
				return
			}

			for _, rec := range deletedRecords {
				t.Log(rec)
			}
		})
	}
}

func Test_dnsRecordToLibdnsRecord_Basic(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	aRecord := libdnsRecord.(libdns.Address)
	assertProviderDataIdAndDomainId(t, aRecord.ProviderData, 123, 456)

	assert.Equal(t, "A", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, time.Duration(120)*time.Second, libdnsRecord.RR().TTL)
}

func Test_dnsRecordToLibdnsRecord_EmptyNodeNameDomain(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.NodeName = ""
	dnsRecord.DomainName = "dynu.com"
	dnsRecord.Hostname = "dynu.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "@", libdnsRecord.RR().Name)
}

func Test_dnsRecordToLibdnsRecord_EmptyNodeNameSubdomain(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.NodeName = ""
	dnsRecord.Hostname = "my.dynu.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "my", libdnsRecord.RR().Name)
}

func Test_dnsRecordToLibdnsRecord_A(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "A"
	dnsRecord.Ipv4Address = "1.2.3.4"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "A", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, "1.2.3.4", libdnsRecord.RR().Data)
}

func Test_dnsRecordToLibdnsRecord_AAAA(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "AAAA"
	dnsRecord.Ipv6Address = "::1"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "AAAA", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, "::1", libdnsRecord.RR().Data)
}

func Test_dnsRecordToLibdnsRecord_CAA(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "CAA"
	dnsRecord.Flags = 128
	dnsRecord.Tag = "issue"
	dnsRecord.Value = "\"comodoca.com\""
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	cnameRecord := libdnsRecord.(libdns.CAA)
	assertProviderDataIdAndDomainId(t, cnameRecord.ProviderData, 123, 456)

	assert.Equal(t, "CAA", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, uint8(128), cnameRecord.Flags)
	assert.Equal(t, "issue", cnameRecord.Tag)
	assert.Equal(t, "\"comodoca.com\"", cnameRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_CNAME(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "CNAME"
	dnsRecord.Host = "www.example.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	cnameRecord := libdnsRecord.(libdns.CNAME)
	assertProviderDataIdAndDomainId(t, cnameRecord.ProviderData, 123, 456)

	assert.Equal(t, "CNAME", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, "www.example.com", cnameRecord.Target)
}

func Test_dnsRecordToLibdnsRecord_MX(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "MX"
	dnsRecord.Host = "www.example.com"
	dnsRecord.Priority = 1
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	mxRecord := libdnsRecord.(libdns.MX)
	assertProviderDataIdAndDomainId(t, mxRecord.ProviderData, 123, 456)

	assert.Equal(t, "MX", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", mxRecord.Name)
	assert.Equal(t, "www.example.com", mxRecord.Target)
	assert.Equal(t, uint16(1), mxRecord.Preference)
}

func Test_dnsRecordToLibdnsRecord_NS(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "NS"
	dnsRecord.Host = "www.example.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	nsRecord := libdnsRecord.(libdns.NS)
	assertProviderDataIdAndDomainId(t, nsRecord.ProviderData, 123, 456)

	assert.Equal(t, "NS", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", nsRecord.Name)
	assert.Equal(t, "www.example.com", nsRecord.Target)
}

func Test_dnsRecordToLibdnsRecord_PTR(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "PTR"
	dnsRecord.Host = "10.207.160.216.in-addr.arpa"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "PTR", libdnsRecord.RR().Type)
	assert.Equal(t, "10.207.160.216.in-addr.arpa", libdnsRecord.RR().Name)
	assert.Equal(t, "abc.my.dynu.com", libdnsRecord.RR().Data)
}

func Test_dnsRecordToLibdnsRecord_SRV(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "SRV"
	dnsRecord.NodeName = "_sip._udp.abc"
	dnsRecord.Priority = 10
	dnsRecord.Weight = 5
	dnsRecord.Port = 5060
	dnsRecord.Host = "example.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	srvRecord := libdnsRecord.(libdns.SRV)
	assertProviderDataIdAndDomainId(t, srvRecord.ProviderData, 123, 456)

	assert.Equal(t, "SRV", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", srvRecord.Name)
	assert.Equal(t, "sip", srvRecord.Service)
	assert.Equal(t, "udp", srvRecord.Transport)
	assert.Equal(t, uint16(10), srvRecord.Priority)
	assert.Equal(t, uint16(5), srvRecord.Weight)
	assert.Equal(t, uint16(5060), srvRecord.Port)
	assert.Equal(t, "example.com", srvRecord.Target)
}

func Test_dnsRecordToLibdnsRecord_TXT(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "TXT"
	dnsRecord.TextData = "ABCD"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	txtRecord := libdnsRecord.(libdns.TXT)
	assertProviderDataIdAndDomainId(t, txtRecord.ProviderData, 123, 456)

	assert.Equal(t, "TXT", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", txtRecord.Name)
	assert.Equal(t, "ABCD", txtRecord.Text)
}

func Test_dnsRecordToLibdnsRecord_UNKNOWN(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "UNKNOWN"
	dnsRecord.Content = "CONTENT"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "UNKNOWN", libdnsRecord.RR().Type)
	assert.Equal(t, "abc.my", libdnsRecord.RR().Name)
	assert.Equal(t, "CONTENT", libdnsRecord.RR().Data)
}

func getBasicDnsRecord() DNSRecord {
	return DNSRecord{
		ID:          123,
		DomainID:    456,
		Type:        "A",
		NodeName:    "abc",
		DomainName:  "my.dynu.com",
		Hostname:    "abc.my.dynu.com",
		TTL:         120,
		Ipv4Address: "0.0.0.0",
	}
}

func Test_libdnsRecordToDnsRecord_Basic(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "A", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, 120, dnsRecord.TTL)
	assert.Equal(t, true, dnsRecord.State)
}

func Test_libdnsRecordToDnsRecord_NoProviderData(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	libdnsRecord.ProviderData = nil
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 0, 0)

	assert.Equal(t, "A", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, 120, dnsRecord.TTL)
	assert.Equal(t, true, dnsRecord.State)
}

func Test_libdnsRecordToDnsRecord_EmptyNodeNameDomain(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	libdnsRecord.Name = "@"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, domain) // test the case where domain is same as owned domain

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "", dnsRecord.NodeName)
}

func Test_libdnsRecordToDnsRecord_EmptyNodeNameSubdomain(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	libdnsRecord.Name = "my"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "", dnsRecord.NodeName)
}

func Test_libdnsRecordToDnsRecord_A(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	libdnsRecord.IP = netip.MustParseAddr("1.2.3.4")
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "A", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "1.2.3.4", dnsRecord.Ipv4Address)
}

func Test_libdnsRecordToDnsRecord_AAAA(t *testing.T) {
	libdnsRecord := getBasicLibDnsAddrRecord()
	libdnsRecord.IP = netip.MustParseAddr("::1")
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "AAAA", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "::1", dnsRecord.Ipv6Address)
}

func Test_libdnsRecordToDnsRecord_CAA(t *testing.T) {
	libdnsRecord := libdns.CAA{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Flags:        128,
		Tag:          "issue",
		Value:        "\"comodoca.com\"",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "CAA", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, 128, dnsRecord.Flags)
	assert.Equal(t, "issue", dnsRecord.Tag)
	assert.Equal(t, "\"comodoca.com\"", dnsRecord.Value)
}

func Test_libdnsRecordToDnsRecord_CNAME(t *testing.T) {
	libdnsRecord := libdns.CNAME{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Target:       "www.example.com",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "CNAME", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_MX(t *testing.T) {
	libdnsRecord := libdns.MX{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Preference:   uint16(1),
		Target:       "www.example.com",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "MX", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
	assert.Equal(t, 1, dnsRecord.Priority)
}

func Test_libdnsRecordToDnsRecord_NS(t *testing.T) {
	libdnsRecord := libdns.NS{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Target:       "www.example.com",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "NS", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_PTR(t *testing.T) {
	libdnsRecord := libdns.RR{
		Name: "10.207.160.216.in-addr.arpa",
		TTL:  time.Duration(120) * time.Second,
		Type: "PTR",
		Data: "abc.my.dynu.com",
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "PTR", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "10.207.160.216.in-addr.arpa", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_SRV(t *testing.T) {
	libdnsRecord := libdns.SRV{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Service:      "sip",
		Transport:    "udp",
		Priority:     10,
		Weight:       5,
		Port:         5060,
		Target:       "example.com",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "SRV", dnsRecord.Type)
	assert.Equal(t, "_sip._udp.abc", dnsRecord.NodeName)
	assert.Equal(t, 10, dnsRecord.Priority)
	assert.Equal(t, 5, dnsRecord.Weight)
	assert.Equal(t, 5060, dnsRecord.Port)
	assert.Equal(t, "example.com", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_TXT(t *testing.T) {
	libdnsRecord := libdns.TXT{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		Text:         "ABCD",
		ProviderData: dummyDynuProviderData,
	}
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assertDnsRecordIdAndDomainId(t, dnsRecord, 123, 456)

	assert.Equal(t, "TXT", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "ABCD", dnsRecord.TextData)
}

func Test_libdnsRecordToDnsRecord_UNKNOWN(t *testing.T) {
	libdnsRecord := libdns.RR{
		Name: "abc.my",
		TTL:  time.Duration(120) * time.Second,
		Type: "UNKNOWN",
		Data: "CONTENT",
	}
	_, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	assert.Error(t, err)
}

func getBasicLibDnsAddrRecord() libdns.Address {
	return libdns.Address{
		Name:         "abc.my",
		TTL:          time.Duration(120) * time.Second,
		IP:           netip.MustParseAddr("0.0.0.0"),
		ProviderData: dummyDynuProviderData,
	}
}

func assertProviderDataIdAndDomainId(t *testing.T, providerData any, id int64, domainId int64) {
	dynuProviderData := providerData.(DynuProviderData)
	assert.Equal(t, id, dynuProviderData.ID)
	assert.Equal(t, domainId, dynuProviderData.DomainID)
}

func assertDnsRecordIdAndDomainId(t *testing.T, dnsRecord DNSRecord, id int64, domainId int64) {
	assert.Equal(t, id, dnsRecord.ID)
	assert.Equal(t, domainId, dnsRecord.DomainID)
}

func stripRecordId(libdnsRecord libdns.Record) libdns.Record {
	// remove ProviderData and hence ID
	return libdnsRecord.RR()
}

func stripRecordIds(libdnsRecords []libdns.Record) {
	for i, rec := range libdnsRecords {
		libdnsRecords[i] = stripRecordId(rec)
	}
}
