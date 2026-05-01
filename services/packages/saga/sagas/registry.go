// Package sagas provides saga handler implementations for all modules
package sagas

// ServiceRegistry maps service names to their HTTP endpoints
// Used by RpcConnector to discover services for saga step execution
var ServiceRegistry = map[string]string{
	// Sales Module (Ports 8119-8128)
	"sales-order":      "http://localhost:8119",
	"sales-invoice":    "http://localhost:8120",
	"crm":              "http://localhost:8121",
	"territory":        "http://localhost:8122",
	"commission":       "http://localhost:8123",
	"pricing":          "http://localhost:8124",
	"dealer":           "http://localhost:8125",
	"sales-analytics":  "http://localhost:8126",
	"route-planning":   "http://localhost:8127",
	"field-sales":      "http://localhost:8128",

	// Purchase Module (Ports 8141-8143)
	"procurement":      "http://localhost:8141",
	"purchase-order":   "http://localhost:8142",
	"purchase-invoice": "http://localhost:8143",

	// Inventory Module (Ports 8179-8186)
	"inventory-core": "http://localhost:8179",
	"wms":             "http://localhost:8180",
	"stock-transfer":  "http://localhost:8181",
	"qc":              "http://localhost:8182",
	"lot-serial":      "http://localhost:8183",
	"cycle-count":     "http://localhost:8184",
	"barcode":         "http://localhost:8185",
	"planning":        "http://localhost:8186",
	"fulfillment":     "http://localhost:8187",
	"shipping":        "http://localhost:8188",

	// Finance Module (Ports 8100-8104)
	"general-ledger":        "http://localhost:8100",
	"accounts-receivable":   "http://localhost:8103",
	"accounts-payable":      "http://localhost:8104",
	"payroll":               "http://localhost:8116",

	// Banking & Tax Module (Ports 8155-8159)
	"tds":          "http://localhost:8158",
	"gst":          "http://localhost:8155",
	"e-invoice":    "http://localhost:8156",
	"e-way-bill":   "http://localhost:8157",
	"banking":      "http://localhost:8159",

	// Manufacturing Module (Ports 8190-8198)
	"bom":                  "http://localhost:8190",
	"production-order":     "http://localhost:8191",
	"production-planning":  "http://localhost:8192",
	"shop-floor":           "http://localhost:8193",
	"quality-production":   "http://localhost:8194",
	"subcontracting":       "http://localhost:8195",
	"work-center":          "http://localhost:8196",
	"routing":              "http://localhost:8197",
	"job-card":             "http://localhost:8198",

	// Finance Module - Additional (Ports 8083-8084, 8105, 8107-8112)
	"journal":              "http://localhost:8083",
	"transaction":          "http://localhost:8084",
	"billing":              "http://localhost:8105",
	"cash-management":      "http://localhost:8088",
	"reconciliation":       "http://localhost:8107",
	"cost-center":          "http://localhost:8109",
	"tax-engine":           "http://localhost:8091",
	"financial-reports":    "http://localhost:8111",
	"financial-close":      "http://localhost:8093",
	"compliance-postings":  "http://localhost:8094",
	"fixed-assets":         "http://localhost:8108",
	"currency":             "http://localhost:7026",
	"depreciation":         "http://localhost:8169",

	// HR & Payroll Module (Ports 8113-8118, 8173-8175)
	"employee":             "http://localhost:8113",
	"leave":                "http://localhost:8114",
	"attendance":           "http://localhost:8115",
	"salary-structure":     "http://localhost:8117",
	"recruitment":          "http://localhost:8118",
	"appraisal":            "http://localhost:8173",
	"expense":              "http://localhost:8174",
	"exit":                 "http://localhost:8175",

	// Projects Module (Ports 8160-8166)
	"project":              "http://localhost:8160",
	"task":                 "http://localhost:8161",
	"timesheet":            "http://localhost:8162",
	"project-costing":      "http://localhost:8163",
	"boq":                  "http://localhost:8164",
	"sub-contractor":       "http://localhost:8165",
	"progress-billing":     "http://localhost:8166",

	// Budget Module (Ports 8193-8196)
	"budget": "http://localhost:8193",

	// Audit & Common Services
	"audit":        "http://localhost:7007",
	"notification": "http://localhost:7005",
	"approval":     "http://localhost:6008",
	"user":         "http://localhost:6003",
	"access":       "http://localhost:6002",
	"asset":        "http://localhost:8167",

	// Additional Services
	"returns": "http://localhost:8189",
}

// GetServiceEndpoint returns the endpoint URL for a service
// Returns empty string if service is not registered
func GetServiceEndpoint(serviceName string) string {
	if endpoint, ok := ServiceRegistry[serviceName]; ok {
		return endpoint
	}
	return ""
}

// RegisterService adds or updates a service endpoint in the registry
func RegisterService(serviceName string, endpoint string) {
	ServiceRegistry[serviceName] = endpoint
}
