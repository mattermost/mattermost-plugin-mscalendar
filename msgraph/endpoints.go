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

	switch tenantType {
	case "gcch", "usgov":
		tenantLoginEndpoint = "https://login.microsoftonline.us/"
	case "china":
		tenantLoginEndpoint = "https://login.chinacloudapi.cn/"
	default:
		tenantLoginEndpoint = "https://login.microsoftonline.com/"
	}

	return oauth2.Endpoint{
		AuthURL:       tenantLoginEndpoint + tenant + "/oauth2/v2.0/authorize",
		TokenURL:      tenantLoginEndpoint + tenant + "/oauth2/v2.0/token",
		DeviceAuthURL: tenantLoginEndpoint + tenant + "/oauth2/v2.0/devicecode",
	}
}

// nolint:revive
// Returns the Microsoft Graph endpoint for the given tenant type.
func MSGraphEndpoint(tenantType string) string {
	switch tenantType {
	case "gcch":
		tenantMSGraphEndpoint = "https://graph.microsoft.us"
	case "usgov":
		tenantMSGraphEndpoint = "https://dod-graph.microsoft.us"
	case "china":
		tenantMSGraphEndpoint = "https://microsoftgraph.chinacloudapi.cn"
	default:
		tenantMSGraphEndpoint = "https://graph.microsoft.com"
	}

	return tenantMSGraphEndpoint + "/v1.0"
}
