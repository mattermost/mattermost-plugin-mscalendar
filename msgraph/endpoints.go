package msgraph

import (
	"golang.org/x/oauth2"
)

var tenantLoginEndpoint string
var tenantMSGraphEndpoint string

// Returns the Entra ID endpoint for the given tenant and tenant type.
func EntraIDEndpoint(tenant string, tenantType string) oauth2.Endpoint {
	if tenant == "" {
		tenant = "common"
	}

	if tenantType == "commercial" {
		tenantLoginEndpoint = "https://login.microsoftonline.com/"
	} else if tenantType == "gcch" {
		tenantLoginEndpoint = "https://login.microsoftonline.us/"
	} else if tenantType == "usgov" {
		tenantLoginEndpoint = "https://login.microsoftonline.us/"
	} else if tenantType == "china" {
		tenantLoginEndpoint = "https://login.chinacloudapi.cn/"
	} else {
		tenantLoginEndpoint = "https://login.microsoftonline.com/"
	}

	return oauth2.Endpoint{
		AuthURL:       tenantLoginEndpoint + tenant + "/oauth2/v2.0/authorize",
		TokenURL:      tenantLoginEndpoint + tenant + "/oauth2/v2.0/token",
		DeviceAuthURL: tenantLoginEndpoint + tenant + "/oauth2/v2.0/devicecode",
	}
}

// Returns the Microsoft Graph endpoint for the given tenant type.
func MSGraphEndpoint(tenantType string) string {
	if tenantType == "commercial" {
		tenantMSGraphEndpoint = "https://graph.microsoft.com"
	} else if tenantType == "gcch" {
		tenantMSGraphEndpoint = "https://graph.microsoft.us"
	} else if tenantType == "usgov" {
		tenantMSGraphEndpoint = "https://dod-graph.microsoft.us"
	} else if tenantType == "china" {
		tenantMSGraphEndpoint = "https://microsoftgraph.chinacloudapi.cn"
	} else {
		tenantMSGraphEndpoint = "https://graph.microsoft.com"
	}

	return tenantMSGraphEndpoint + "/v1.0"
}
