package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"gopkg.in/yaml.v2"
)

type Config struct {
	APIToken string `yaml:"api_token"`
	Email    string `yaml:"email"`
	Key      string `yaml:"key"`
	UseToken bool   `yaml:"use_token"`
}

// OutputData represents the data to output along with an optional error message.
type OutputData struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var outputFormatGlob string = ""

func main() {
	configPath := flag.String("config", "./CloudflareRDNS.yaml", "path to the configuration file")
	outputFormat := flag.String("output", "json", "output format: json or markdown")
	outputFormatGlob = *outputFormat
	ip := flag.String("ip", "", "IP address for rDNS operation")
	check := flag.String("check-api", "", "Runs a call to get your user details from the Cloudflare API")
	newPtr := flag.String("set-rdns", "", "Set or update rDNS to this value")
	flag.Parse()

	config := Config{}
	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		os.Exit(1)
	}

	var api *cloudflare.API
	if config.UseToken {
		api, err = cloudflare.NewWithAPIToken(config.APIToken)
	} else {
		api, err = cloudflare.New(config.Key, config.Email)
	}
	if err != nil {
		fmt.Println("Error creating Cloudflare API instance:", err)
		os.Exit(1)
	}
	if *check != "" {
		// Most API calls require a Context
		ctx := context.Background()

		// Fetch user details on the account
		u, err := api.UserDetails(ctx)
		if err != nil {
			log.Fatal(err)
		}
		// Print user details
		fmt.Println(u)
		os.Exit(0)
	}
	if *ip == "" {
		fmt.Println("Please provide the IP address to update")

		os.Exit(1)
	}
	if *newPtr != "" {
		// Set or update rDNS
		//fmt.Println("Updating rDNS for IP:", *ip)
		// Dummy function call, replace with actual Cloudflare API call
		if err := updateRDNS(api, *ip, *newPtr); err != nil {
			sendOutput(false, "Error updating rDNS", err.Error())
			os.Exit(1)
		}
	} else {
		// Get rDNS
		//fmt.Println("Fetching rDNS for IP:", *ip)
		_, err := getRDNS(api, *ip)
		if err != nil {
			sendOutput(false, "Error fetching rDNS", err.Error())
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func getArpaZone(ip string) (string, error) {
	// Parse the IP address to ensure it is valid
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		sendOutput(false, "Error to get rDNS Zone", "Invalid IP address")
		return "", nil
	}

	// Extract the first three octets for a /24 network
	octets := strings.Split(ip, ".")
	if len(octets) != 4 {
		sendOutput(false, "Error to get rDNS Zone", "Invalid IP address")
		return "", fmt.Errorf("invalid IP address format")
	}

	// Reverse the order of the first three octets
	reversedOctets := fmt.Sprintf("%s.%s.%s.in-addr.arpa", octets[2], octets[1], octets[0])
	return reversedOctets, nil
}

func getCloudflareZone(api *cloudflare.API, ip string) string {
	arpa, _ := getArpaZone(ip)
	zoneID, err := api.ZoneIDByName(arpa)
	if err != nil {
		sendOutput(false, "Error to get rDNS Zone", err.Error())
		return ""
	}
	return zoneID
}

func getRDNS(api *cloudflare.API, ip string) (string, error) {
	// Dummy implementation, replace with actual Cloudflare API call
	zoneID := getCloudflareZone(api, ip)
	octets := strings.Split(ip, ".")
	arpa, _ := getArpaZone(ip)
	records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Name: octets[3] + "." + arpa})
	// Fetch DNS records according to the criteria
	if err != nil {
		fmt.Println("Error fetching DNS records:", err)
		return "", err
	}

	// Check if any records are returned and print them
	if len(records) == 0 {
		sendOutput(true, "", "No PTR record found for IP")
	} else {
		for _, record := range records {
			sendOutput(true, record.Content, record.Name)
			return record.Content, nil
		}
	}

	return "", nil
}

func updateRDNS(api *cloudflare.API, ip, ptr string) error {
	// Dummy implementation, replace with actual Cloudflare API call
	zoneID := getCloudflareZone(api, ip)
	octets := strings.Split(ip, ".")
	arpa, _ := getArpaZone(ip)
	records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Name: octets[3] + "." + arpa})
	// Fetch DNS records according to the criteria
	if err != nil {
		fmt.Println("Error fetching DNS records:", err)
		return err
	}

	// Check if any records are returned and print them
	if len(records) == 0 {
		rr := cloudflare.CreateDNSRecordParams{
			Name:    octets[3] + "." + arpa,
			Type:    "ptr",
			Content: ptr,
		}

		// TODO: Print the response.
		_, err := api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), rr)
		if err != nil {
			sendOutput(false, "Error creating DNS record", err.Error())
			return err
		}
		sendOutput(true, ptr, "RDNS created for "+ip)

	} else {
		for _, r := range records {

			rr := cloudflare.UpdateDNSRecordParams{}
			rr.ID = r.ID
			rr.Type = r.Type
			rr.Content = ptr

			_, err = api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), rr)
			if err != nil {
				fmt.Println("Error updating DNS record:", err)
				return err
			}
			sendOutput(true, ptr, "RDNS Updated for "+ip)
		}
	}
	return nil
}

func (o OutputData) ToJSON() (string, error) {
	jsonData, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ToMarkdown outputs the OutputData as a Markdown string.
func (o OutputData) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("### Output\n\n")
	sb.WriteString(fmt.Sprintf("**Success:** `%t`\n\n", o.Success))
	sb.WriteString(fmt.Sprintf("**Message:** %s\n\n", o.Message))
	if o.Data != nil {
		sb.WriteString("#### Data\n")
		sb.WriteString(fmt.Sprintf("```json\n%v\n```\n", o.Data))
	}
	return sb.String()
}
func (o OutputData) Output(format string) (string, error) {
	switch format {
	case "json":
		return o.ToJSON()
	case "markdown":
		return o.ToMarkdown(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
func sendOutput(success bool, message string, error string) {
	data := OutputData{
		Success: success,
		Message: message,
		Data: map[string]interface{}{
			"details": error,
		},
	}

	output, err := data.Output(outputFormatGlob)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(output)
}
