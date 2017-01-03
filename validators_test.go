package revel

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	errorsMessage   = "validation for %s should not be satisfied with %s\n"
	noErrorsMessage = "validation for %s should be satisfied with %s\n"
)

type Expect struct {
	input          interface{}
	expectedResult bool
	errorMessage   string
}

func performTests(validator Validator, tests []Expect, t *testing.T) {
	for _, test := range tests {
		if validator.IsSatisfied(test.input) != test.expectedResult {
			if test.expectedResult == false {
				t.Errorf(errorsMessage, reflect.TypeOf(validator), test.errorMessage)
			} else {
				t.Errorf(noErrorsMessage, reflect.TypeOf(validator), test.errorMessage)
			}
		}
	}
}

func TestRequired(t *testing.T) {

	tests := []Expect{
		Expect{nil, false, "nil data"},
		Expect{"Testing", true, "non-empty string"},
		Expect{"", false, "empty string"},
		Expect{true, true, "true boolean"},
		Expect{false, false, "false boolean"},
		Expect{1, true, "positive integer"},
		Expect{-1, true, "negative integer"},
		Expect{0, false, "0 integer"},
		Expect{time.Now(), true, "current time"},
		Expect{time.Time{}, false, "a zero time"},
		Expect{func() {}, true, "other non-nil data types"},
	}

	// testing both the struct and the helper method
	for _, required := range []Required{Required{}, ValidRequired()} {
		performTests(required, tests, t)
	}
}

func TestMin(t *testing.T) {
	tests := []Expect{
		Expect{11, true, "val > min"},
		Expect{10, true, "val == min"},
		Expect{9, false, "val < min"},
		Expect{true, false, "TypeOf(val) != int"},
	}
	for _, min := range []Min{Min{10}, ValidMin(10)} {
		performTests(min, tests, t)
	}
}

func TestMax(t *testing.T) {
	tests := []Expect{
		Expect{9, true, "val < max"},
		Expect{10, true, "val == max"},
		Expect{11, false, "val > max"},
		Expect{true, false, "TypeOf(val) != int"},
	}
	for _, max := range []Max{Max{10}, ValidMax(10)} {
		performTests(max, tests, t)
	}
}

func TestRange(t *testing.T) {
	tests := []Expect{
		Expect{50, true, "min <= val <= max"},
		Expect{10, true, "val == min"},
		Expect{100, true, "val == max"},
		Expect{9, false, "val < min"},
		Expect{101, false, "val > max"},
	}

	goodValidators := []Range{
		Range{Min{10}, Max{100}},
		ValidRange(10, 100),
	}
	for _, rangeValidator := range goodValidators {
		performTests(rangeValidator, tests, t)
	}

	tests = []Expect{
		Expect{10, true, "min == val == max"},
		Expect{9, false, "val < min && val < max && min == max"},
		Expect{11, false, "val > min && val > max && min == max"},
	}

	goodValidators = []Range{
		Range{Min{10}, Max{10}},
		ValidRange(10, 10),
	}
	for _, rangeValidator := range goodValidators {
		performTests(rangeValidator, tests, t)
	}

	tests = make([]Expect, 7)
	for i, num := range []int{50, 100, 10, 9, 101, 0, -1} {
		tests[i] = Expect{
			num,
			false,
			"min > val < max",
		}
	}
	// these are min/max with values swapped, so the min is the high
	// and max is the low. rangeValidator.IsSatisfied() should ALWAYS
	// result in false since val can never be greater than min and less
	// than max when min > max
	badValidators := []Range{
		Range{Min{100}, Max{10}},
		ValidRange(100, 10),
	}
	for _, rangeValidator := range badValidators {
		performTests(rangeValidator, tests, t)
	}
}

func TestMinSize(t *testing.T) {
	greaterThanMessage := "len(val) >= min"
	tests := []Expect{
		Expect{"12", true, greaterThanMessage},
		Expect{"123", true, greaterThanMessage},
		Expect{[]int{1, 2}, true, greaterThanMessage},
		Expect{[]int{1, 2, 3}, true, greaterThanMessage},
		Expect{"", false, "len(val) <= min"},
		Expect{"手", false, "len(val) <= min"},
		Expect{[]int{}, false, "len(val) <= min"},
		Expect{nil, false, "TypeOf(val) != string && TypeOf(val) != slice"},
	}

	for _, minSize := range []MinSize{MinSize{2}, ValidMinSize(2)} {
		performTests(minSize, tests, t)
	}
}

func TestMaxSize(t *testing.T) {
	lessThanMessage := "len(val) <= max"
	tests := []Expect{
		Expect{"", true, lessThanMessage},
		Expect{"12", true, lessThanMessage},
		Expect{"ルビー", true, lessThanMessage},
		Expect{[]int{}, true, lessThanMessage},
		Expect{[]int{1, 2}, true, lessThanMessage},
		Expect{[]int{1, 2, 3}, true, lessThanMessage},
		Expect{"1234", false, "len(val) >= max"},
		Expect{[]int{1, 2, 3, 4}, false, "len(val) >= max"},
	}
	for _, maxSize := range []MaxSize{MaxSize{3}, ValidMaxSize(3)} {
		performTests(maxSize, tests, t)
	}
}

func TestLength(t *testing.T) {
	tests := []Expect{
		Expect{"12", true, "len(val) == length"},
		Expect{"火箭", true, "len(val) == length"},
		Expect{[]int{1, 2}, true, "len(val) == length"},
		Expect{"123", false, "len(val) > length"},
		Expect{[]int{1, 2, 3}, false, "len(val) > length"},
		Expect{"1", false, "len(val) < length"},
		Expect{[]int{1}, false, "len(val) < length"},
		Expect{nil, false, "TypeOf(val) != string && TypeOf(val) != slice"},
	}
	for _, length := range []Length{Length{2}, ValidLength(2)} {
		performTests(length, tests, t)
	}
}

func TestMatch(t *testing.T) {
	tests := []Expect{
		Expect{"bca123", true, `"[abc]{3}\d*" matches "bca123"`},
		Expect{"bc123", false, `"[abc]{3}\d*" does not match "bc123"`},
		Expect{"", false, `"[abc]{3}\d*" does not match ""`},
	}
	regex := regexp.MustCompile(`[abc]{3}\d*`)
	for _, match := range []Match{Match{regex}, ValidMatch(regex)} {
		performTests(match, tests, t)
	}
}

func TestEmail(t *testing.T) {
	// unicode char included
	validStartingCharacters := strings.Split("!#$%^&*_+1234567890abcdefghijklmnopqrstuvwxyzñ", "")
	invalidCharacters := strings.Split(" ()", "")

	definiteInvalidDomains := []string{
		"",                  // any empty string (x@)
		".com",              // only the TLD (x@.com)
		".",                 // only the . (x@.)
		".*",                // TLD containing symbol (x@.*)
		"asdf",              // no TLD
		"a!@#$%^&*()+_.com", // characters which are not ASCII/0-9/dash(-) in a domain
		"-a.com",            // host starting with any symbol
		"a-.com",            // host ending with any symbol
		"aå.com",            // domain containing unicode (however, unicode domains do exist in the state of xn--<POINT>.com e.g. å.com = xn--5ca.com)
	}

	for _, email := range []Email{Email{Match{emailPattern}}, ValidEmail()} {
		var currentEmail string

		// test invalid starting chars
		for _, startingChar := range validStartingCharacters {
			currentEmail = fmt.Sprintf("%sñbc+123@do-main.com", startingChar)
			if email.IsSatisfied(currentEmail) {
				t.Errorf(noErrorsMessage, "starting characters", fmt.Sprintf("email = %s", currentEmail))
			}

			// validation should fail because of multiple @ symbols
			currentEmail = fmt.Sprintf("%s@ñbc+123@do-main.com", startingChar)
			if email.IsSatisfied(currentEmail) {
				t.Errorf(errorsMessage, "starting characters with multiple @ symbols", fmt.Sprintf("email = %s", currentEmail))
			}

			// should fail simply because of the invalid char
			for _, invalidChar := range invalidCharacters {
				currentEmail = fmt.Sprintf("%sñbc%s+123@do-main.com", startingChar, invalidChar)
				if email.IsSatisfied(currentEmail) {
					t.Errorf(errorsMessage, "invalid starting characters", fmt.Sprintf("email = %s", currentEmail))
				}
			}
		}

		// test invalid domains
		for _, invalidDomain := range definiteInvalidDomains {
			currentEmail = fmt.Sprintf("a@%s", invalidDomain)
			if email.IsSatisfied(currentEmail) {
				t.Errorf(errorsMessage, "invalid domain", fmt.Sprintf("email = %s", currentEmail))
			}
		}

		// should always be satisfied
		if !email.IsSatisfied("t0.est+email123@1abc0-def.com") {
			t.Errorf(noErrorsMessage, "guaranteed valid email", fmt.Sprintf("email = %s", "t0.est+email123@1abc0-def.com"))
		}

		// should never be satisfied (this is redundant given the loops above)
		if email.IsSatisfied("a@xcom") {
			t.Errorf(noErrorsMessage, "guaranteed invalid email", fmt.Sprintf("email = %s", "a@xcom"))
		}
		if email.IsSatisfied("a@@x.com") {
			t.Errorf(noErrorsMessage, "guaranteed invalid email", fmt.Sprintf("email = %s", "a@@x.com"))
		}
	}
}

func runIPAddrTestfunc(t *testing.T, test_type int, ipaddr_list map[string]bool, msg_fmt string) {

	// generate dataset for test
	test_ipaddr_list := []Expect{}
	for ipaddr, expected := range ipaddr_list {
		test_ipaddr_list = append(test_ipaddr_list, Expect{input: ipaddr, expectedResult: expected, errorMessage: fmt.Sprintf(msg_fmt, ipaddr)})
	}

	for _, ip_test_list := range []IPAddr{IPAddr{[]int{test_type}}, ValidIPAddr(test_type)} {
		performTests(ip_test_list, test_ipaddr_list, t)
	}
}

func TestIPAddr(t *testing.T) {

	//IPv4
	test_ipv4_ipaddrs := map[string]bool{
		"192.168.1.1":     true,
		"127.0.0.1":       true,
		"10.10.90.12":     true,
		"8.8.8.8":         true,
		"4.4.4.4":         true,
		"912.456.123.123": false,
		"999.999.999.999": false,
		"192.192.19.999":  false,
	}

	//IPv4 with CIDR
	test_ipv4_with_cidr_ipaddrs := map[string]bool{
		"192.168.1.1/24": true,
		"127.0.0.1/32":   true,
		"10.10.90.12/8":  true,
		"8.8.8.8/1":      true,
		"4.4.4.4/7":      true,
		"192.168.1.1/99": false,
		"127.0.0.1/9999": false,
		"10.10.90.12/33": false,
		"8.8.8.8/128":    false,
		"4.4.4.4/256":    false,
	}

	//IPv6
	test_ipv6_ipaddrs := map[string]bool{
		"2607:f0d0:1002:51::4":                    true,
		"2607:f0d0:1002:0051:0000:0000:0000:0004": true,
		"ff05::1:3":                               true,
		"FE80:0000:0000:0000:0202:B3FF:FE1E:8329": true,
		"FE80::0202:B3FF:FE1E:8329":               true,
		"fe80::202:b3ff:fe1e:8329":                true,
		"fe80:0000:0000:0000:0202:b3ff:fe1e:8329": true,
		"2001:470:1f09:495::3":                    true,
		"2001:470:1f1d:275::1":                    true,
		"2600:9000:5304:200::1":                   true,
		"2600:9000:5306:d500::1":                  true,
		"2600:9000:5301:b600::1":                  true,
		"2600:9000:5303:900::1":                   true,
		"127:12:12:12:12:12:!2:!2":                false,
		"127.0.0.1":                               false,
		"234:23:23:23:23:23:23":                   false,
	}

	//IPv6 with CIDR
	test_ipv6_with_cidr_ipaddrs := map[string]bool{
		"2000::/5":      true,
		"2000::/15":     true,
		"2001:db8::/33": true,
		"2001:db8::/48": true,
		"fc00::/7":      true,
	}

	//IPv4-Mapped Embedded IPv6 Address
	test_ipv4_mapped_ipv6_ipaddrs := map[string]bool{
		"2001:470:1f09:495::3:217.126.185.215":         true,
		"2001:470:1f1d:275::1:213.0.69.132":            true,
		"2600:9000:5304:200::1:205.251.196.2":          true,
		"2600:9000:5306:d500::1:205.251.198.213":       true,
		"2600:9000:5301:b600::1:205.251.193.182":       true,
		"2600:9000:5303:900::1:205.251.195.9":          true,
		"0:0:0:0:0:FFFF:222.1.41.90":                   true,
		"::FFFF:222.1.41.90":                           true,
		"0000:0000:0000:0000:0000:FFFF:12.155.166.101": true,
		"12.155.166.101":                               false,
		"12.12/12":                                     false,
	}

	runIPAddrTestfunc(t, IPv4, test_ipv4_ipaddrs, "invalid (%s) IPv4 address")
	runIPAddrTestfunc(t, IPv4CIDR, test_ipv4_with_cidr_ipaddrs, "invalid (%s) IPv4 with CIDR address")

	runIPAddrTestfunc(t, IPv6, test_ipv6_ipaddrs, "invalid (%s) IPv6 address")
	runIPAddrTestfunc(t, IPv6CIDR, test_ipv6_with_cidr_ipaddrs, "invalid (%s) IPv6 with CIDR address")
	runIPAddrTestfunc(t, IPv4MappedIPv6, test_ipv4_mapped_ipv6_ipaddrs, "invalid (%s) IPv4-Mapped Embedded IPv6 address")
}

func TestMacAddr(t *testing.T) {

	macaddr_list := map[string]bool{
		"02:f3:71:eb:9e:4b": true,
		"02-f3-71-eb-9e-4b": true,
		"02f3.71eb.9e4b":    true,
		"87:78:6e:3e:90:40": true,
		"87-78-6e-3e-90-40": true,
		"8778.6e3e.9040":    true,
		"e7:28:b9:57:ab:36": true,
		"e7-28-b9-57-ab-36": true,
		"e728.b957.ab36":    true,
		"eb:f8:2b:d7:e9:62": true,
		"eb-f8-2b-d7-e9-62": true,
		"ebf8.2bd7.e962":    true,
	}

	test_macaddr_list := []Expect{}
	for macaddr, expected := range macaddr_list {
		test_macaddr_list = append(test_macaddr_list, Expect{input: macaddr, expectedResult: expected, errorMessage: fmt.Sprintf("invalid (%s) MAC address", macaddr)})
	}

	for _, mac_test_list := range []MacAddr{MacAddr{}, ValidMacAddr()} {
		performTests(mac_test_list, test_macaddr_list, t)
	}
}

func TestDomain(t *testing.T) {

	test_domains := map[string]bool{
		"대한민국.xn-korea.co.kr":           true,
		"google.com":                    true,
		"masełkowski.pl":                true,
		"maselkowski.pl":                true,
		"m.maselkowski.pl":              true,
		"www.masełkowski.pl.com":        true,
		"xn--masekowski-d0b.pl":         true,
		"中国互联网络信息中心.中国":                 true,
		"masełkowski.pl.":               false,
		"中国互联网络信息中心.xn--masekowski-d0b": false,
		"a.jp":                     true,
		"a.co":                     true,
		"a.co.jp":                  true,
		"a.co.or":                  true,
		"a.or.kr":                  true,
		"qwd-qwdqwd.com":           true,
		"qwd-qwdqwd.co_m":          false,
		"qwd-qwdqwd.c":             false,
		"qwd-qwdqwd.-12":           false,
		"qwd-qwdqwd.1212":          false,
		"qwd-qwdqwd.org":           true,
		"qwd-qwdqwd.ac.kr":         true,
		"qwd-qwdqwd.gov":           true,
		"chicken.beer":             true,
		"aa.xyz":                   true,
		"google.asn.au":            true,
		"google.com.au":            true,
		"google.net.au":            true,
		"google.priv.at":           true,
		"google.ac.at":             true,
		"google.gv.at":             true,
		"google.avocat.fr":         true,
		"google.geek.nz":           true,
		"google.gen.nz":            true,
		"google.kiwi.nz":           true,
		"google.org.il":            true,
		"google.net.il":            true,
		"www.google.edu.au":        true,
		"www.google.gov.au":        true,
		"www.google.csiro.au":      true,
		"www.google.act.au":        true,
		"www.google.avocat.fr":     true,
		"www.google.aeroport.fr":   true,
		"www.google.co.nz":         true,
		"www.google.geek.nz":       true,
		"www.google.gen.nz":        true,
		"www.google.kiwi.nz":       true,
		"www.google.parliament.nz": true,
		"www.google.muni.il":       true,
		"www.google.idf.il":        true,
	}

	tests := []Expect{}

	for domain, expected := range test_domains {
		tests = append(tests, Expect{input: domain, expectedResult: expected, errorMessage: fmt.Sprintf("invalid (%s) domain", domain)})
	}

	for _, domain := range []Domain{Domain{}, ValidDomain()} {
		performTests(domain, tests, t)
	}
}

func TestURL(t *testing.T) {

	test_urls := map[string]bool{
		"https://www.google.co.kr/url?sa=t&rct=j&q=&esrc=s&source=web":                                      true,
		"http://stackoverflow.com/questions/27812164/can-i-import-3rd-party-package-into-golang-playground": true,
		"https://tour.golang.org/welcome/4":                                                                 true,
		"https://revel.github.io/":                                                                          true,
		"https://github.com/revel/revel/commit/bd1d083ee4345e919b3bca1e4c42ca682525e395#diff-972a2b2141d27e9d7a8a4149a7e28eef": true,
		"https://github.com/ndevilla/iniparser/pull/82#issuecomment-261817064":                                                 true,
		"http://www.baidu.com/s?ie=utf-8&f=8&rsv_bp=0&rsv_idx=1&tn=baidu&wd=golang":                                            true,
		"http://www.baidu.com/link?url=DrWkM_beo2M5kB5sLYnItKSQ0Ib3oDhKcPprdtLzAWNfFt_VN5oyD3KwnAKT6Xsk":                       true,
	}

	tests := []Expect{}

	for url, expected := range test_urls {
		tests = append(tests, Expect{input: url, expectedResult: expected, errorMessage: fmt.Sprintf("invalid (%s) url", url)})
	}

	for _, url := range []URL{URL{}, ValidURL()} {
		performTests(url, tests, t)
	}

}

func TestPureText(t *testing.T) {

	test_txts := map[string]bool{
		"<script ?>qwdpijqwd</script>qd08j123lneqw\t\nqwedojiqwd\rqwdoihjqwd1d[08jaedl;jkqwd\r\nqdolijqdwqwd":       false,
		"a\r\nb<script ?>qwdpijqwd</script>qd08j123lneqw\t\nqwedojiqwd\rqwdoihjqwd1d[08jaedl;jkqwd\r\nqdolijqdwqwd": false,
		`Foo<script type="text/javascript">alert(1337)</script>Bar`:                                                 false,
		`Foo<12>Bar`:              true,
		`Foo<>Bar`:                true,
		`Foo</br>Bar`:             false,
		`Foo <!-- Bar --> Baz`:    true,
		`I <3 Ponies!`:            true,
		`I &#32; like Golang\t\n`: true,
		`I &amp; like Golang\t\n`: false,
		`<?xml version="1.0" encoding="UTF-8" ?> <!DOCTYPE log4j:configuration SYSTEM "log4j.dtd"> <log4j:configuration debug="true" xmlns:log4j='http://jakarta.apache.org/log4j/'> <appender name="console" class="org.apache.log4j.ConsoleAppender"> <layout class="org.apache.log4j.PatternLayout"> <param name="ConversionPattern" value="%d{yyyy-MM-dd HH:mm:ss} %-5p %c{1}:%L - %m%n" /> </layout> </appender> <root> <level value="DEBUG" /> <appender-ref ref="console" /> </root> </log4j:configuration>`: false,
		"I like Golang\r\n":  true,
		"I like Golang\r\na": false,
	}

	tests := []Expect{}

	for txt, expected := range test_txts {
		tests = append(tests, Expect{input: txt, expectedResult: expected, errorMessage: fmt.Sprintf("invalid (%#v) text", txt)})
	}

	for _, txt := range []PureText{PureText{}, ValidPureText()} {
		performTests(txt, tests, t)
	}
}
