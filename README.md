# Cloudflare rDNS Management Tool

This Go application manages reverse DNS (rDNS) configurations using the Cloudflare API. It allows you to set or update rDNS records, check API connectivity, and handle configurations through a YAML file.

## Features

- Set or update rDNS records for specified IP addresses.
- Fetch existing rDNS settings for an IP.
- Validate Cloudflare API connectivity.
- Support for output in JSON or Markdown formats.

## Prerequisites

- Go (Golang) installed on your machine.
- Cloudflare account with API token or Key and Email.

## Installation

Clone the repository to your local machine:

```
git clone https://github.com/AndrewAubury/CloudflareRDNSClient.git
cd CloudflareRDNSClient
```

To build the application and install it to your OS path:

```
go build -o cloudflare-rdns
sudo mv cloudflare-rdns /usr/local/bin/
```

This will make `cloudflare-rdns` executable available globally on your system.

## Configuration

Create a YAML configuration file named `CloudflareRDNS.yaml` and update it with your Cloudflare credentials:

```
api_token: "your_cloudflare_api_token"
email: "your_email@example.com"
key: "your_cloudflare_global_api_key"
use_token: true  # Use true to use token, false to use key and email
```

## Usage

To run the tool, use the following commands:

- **Check API Connectivity:**

  ```
  cloudflare-rdns --check-api "test"
  ```

- **Set or Update rDNS:**

  ```
  cloudflare-rdns --set-rdns "ptr.example.com" --ip "192.0.2.1"
  ```

- **Fetch rDNS:**

  ```
  cloudflare-rdns --ip "192.0.2.1"
  ```

## Output Formats

Specify the output format with the `--output` flag:

- JSON (default)
- Markdown

For example:

```
cloudflare-rdns --ip "192.0.2.1" --output "markdown"
```

## Contributing

Contributions are welcome. Please open an issue first to discuss what you would like to change.

## Support

For support, open an issue in the GitHub repository.
