package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	ZOOMURLMAIN      = "https://assets.zoom.us/docs/ipranges/Zoom.txt"
	ZOMMURLMEETINGS  = "https://assets.zoom.us/docs/ipranges/ZoomMeetings.txt"
	ZOOMURLCONNECTOR = "https://assets.zoom.us/docs/ipranges/ZoomCRC.txt"
	ZOOMURLPHONE     = "https://assets.zoom.us/docs/ipranges/ZoomPhone.txt"
)

func main() {
	URLs := []string{ZOOMURLMAIN, ZOOMURLCONNECTOR, ZOMMURLMEETINGS, ZOOMURLPHONE}

	BitMaskWilcard := map[string]string{
		"8":  "0.255.255.255",
		"9":  "0.127.255.255",
		"10": "0.63.255.255",
		"11": "0.31.255.255",
		"12": "0.15.255.255",
		"13": "0.7.255.255",
		"14": "0.3.255.255",
		"15": "0.1.255.255",
		"16": "0.0.255.255",
		"17": "0.0.127.255",
		"18": "0.0.63.255",
		"19": "0.0.31.255",
		"20": "0.0.15.255",
		"21": "0.0.7.255",
		"22": "0.0.3.255",
		"23": "0.0.1.255",
		"24": "0.0.0.255",
		"25": "0.0.0.127",
		"26": "0.0.0.63",
		"27": "0.0.0.31",
		"28": "0.0.0.15",
		"29": "0.0.0.7",
		"30": "0.0.0.3",
		"31": "0.0.0.1",
		"32": "0.0.0.0"}

	BitMask := map[string]string{
		"8":  "255.0.0.0",
		"9":  "255.128.0.0",
		"10": "255.192.0.0",
		"11": "255.224.0.0",
		"12": "255.240.0.0",
		"13": "255.248.0.0",
		"14": "255.252.0.0",
		"15": "255.254.0.0",
		"16": "255.255.0.0",
		"17": "255.255.128.0",
		"18": "255.255.192.0",
		"19": "255.255.224.0",
		"20": "255.255.240.0",
		"21": "255.255.248.0",
		"22": "255.255.252.0",
		"23": "255.255.254.0",
		"24": "255.255.255.0",
		"25": "255.255.255.128",
		"26": "255.255.255.192",
		"27": "255.255.255.224",
		"28": "255.255.255.240",
		"29": "255.255.255.248",
		"30": "255.255.255.252",
		"31": "255.255.255.254",
		"32": "255.255.255.255"}

	ACLStartnumber := flag.Int("sn", 10, "Start number inside ACL")
	ACLNumberStep := flag.Int("st", 10, "Step")
	ACLrowForASAPrefix := flag.String("asa", "", "for asa name ACL and line, for example: access-list WCCPREDIRECTACL")
	ACLnoObjectGroupRowOnlyASA := flag.Bool("asaog", false, "Generate Object Group lines")
	ACLRowPrefix := flag.String("rp", "permit ip any", "acl rule prefix, for example: permit ip 10.20.30.0 0.0.0.255")
	ZUUMURLsflags := flag.String("uf", "1.1.1.1", "Flags, 1 - enable URL 0 disable url <zoom>.<soom meetings>.<zoomCRC>.<zoom phones>, example: 1.1.0.1 enable zoom soom meetings and zoom phones")
	flag.Parse()
	ACLtextDestination := make([]string, 0)
	ACLtextDestinationWildcard := make([]string, 0)

	RowNumber := *ACLStartnumber
	URLIndexes := make([]int, 0)
	URLflag := strings.Split(*ZUUMURLsflags, ".")
	for URLindexInR, URLflagCurrent := range URLflag {
		URLindex, URLIndexConversionError := strconv.Atoi(URLflagCurrent)
		if URLIndexConversionError != nil {
			fmt.Println("please provide correct flags!")
			os.Exit(1)
		} else {
			if URLindex == 1 {
				URLIndexes = append(URLIndexes, URLindexInR)
			}
		}
	}

	for _, ZOOMCurrentUrlIndexinR := range URLIndexes {
		ZoomData, ZoomError := http.Get(URLs[ZOOMCurrentUrlIndexinR])
		if !*ACLnoObjectGroupRowOnlyASA {
			RemarkLine := fmt.Sprintf("%s %d remark %s", *ACLrowForASAPrefix, RowNumber, URLs[ZOOMCurrentUrlIndexinR])
			ACLtextDestination = append(ACLtextDestination, RemarkLine)
			ACLtextDestinationWildcard = append(ACLtextDestinationWildcard, RemarkLine)
			RowNumber += *ACLNumberStep
		}

		if ZoomError != nil {
			fmt.Println(ZoomError)
			os.Exit(1)
		} else {
			defer ZoomData.Body.Close()
			ZoomDataBody, ZoomDataBodyReadError := ioutil.ReadAll(ZoomData.Body)
			if ZoomDataBodyReadError != nil {
				fmt.Println(ZoomDataBodyReadError)
				os.Exit(1)
			}
			ACLtextList := strings.Split(string(ZoomDataBody), "\n")
			ACLRowDestination := ""
			ACLRowDestinationWildcard := ""
			for _, ACLrowSource := range ACLtextList {
				TrimmedString := strings.TrimSpace(ACLrowSource)
				IP_prefix := strings.Split(TrimmedString, "/")
				if len(IP_prefix) == 2 {
					if len(IP_prefix[1]) > 0 {
						Mask := BitMask[IP_prefix[1]]
						Wildcard := BitMaskWilcard[IP_prefix[1]]
						if *ACLnoObjectGroupRowOnlyASA {
							ACLRowDestination = fmt.Sprintf("%s %s %s", "network object", IP_prefix[0], Mask)
						} else {
							ACLRowDestination = fmt.Sprintf("%s line %d %s %s %s", *ACLrowForASAPrefix, RowNumber, *ACLRowPrefix, IP_prefix[0], Mask)
							ACLRowDestinationWildcard = fmt.Sprintf("%s line %d %s %s %s", *ACLrowForASAPrefix, RowNumber, *ACLRowPrefix, IP_prefix[0], Wildcard)
						}

					}
				} else {
					if *ACLnoObjectGroupRowOnlyASA {
						ACLRowDestination = fmt.Sprintf("%s %s %s", "network object", "host", IP_prefix[0])
					} else {
						ACLRowDestination = fmt.Sprintf("%d %s host %s", RowNumber, *ACLRowPrefix, TrimmedString)
						ACLRowDestinationWildcard = ACLRowDestination
					}

				}
				if len(ACLRowDestination) > 7 {
					ACLtextDestination = append(ACLtextDestination, ACLRowDestination)
					ACLtextDestinationWildcard = append(ACLtextDestinationWildcard, ACLRowDestinationWildcard)
				}
				RowNumber += *ACLNumberStep
			}
		}
	}

	ACLtext := ""
	for _, ACLdataIn := range ACLtextDestination {
		ACLtext += ACLdataIn + "\n"
	}
	ACLtextWildcard := ""
	for _, ACLdataIn := range ACLtextDestinationWildcard {
		ACLtextWildcard += ACLdataIn + "\n"
	}
	if *ACLnoObjectGroupRowOnlyASA {
		fmt.Println(ACLtext)
	} else {
		if len(*ACLrowForASAPrefix) > 1 {
			fmt.Println(ACLtext)
		} else {
			fmt.Println(ACLtextWildcard)
		}
	}

}
