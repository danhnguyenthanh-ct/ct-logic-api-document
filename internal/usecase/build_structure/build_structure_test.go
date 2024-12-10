package buildstructure

import (
	"context"
	"fmt"
	"testing"

	fetchdata "github.com/ct-logic-api-document/internal/usecase/fetch_data"
	"github.com/stretchr/testify/require"
)

func TestBuildStructure(t *testing.T) {
	// fake fetch data
	fetchdataUC := fetchdata.NewFetchDataUC(
		conf,
		mongoStorage,
	)
	samplesForFetchData := []string{
		`{"request":{"method":"POST","body":"{\"services\":[{\"target\":1138330,\"type\":\"bump\",\"params\":{\"duration\":1}},{\"target\":1138330,\"type\":\"sticky_ad\",\"params\":{\"duration\":3}},{\"target\":1138330,\"type\":\"special_display\",\"params\":{\"duration\":14}},{\"target\":1138330,\"type\":\"bundle\",\"params\":{\"ad_id\":1138330,\"bundle_id\":\"151\",\"category_id\":1030,\"region_id\":12000,\"segment_id\":\"all\"}}]}","querystring":{"cart_id":"oneclick"},"uri":"/v2/private/cart/services?cart_id=oneclick","headers":{"x-chotot-id-key":"L6bwKf6aPC6ejBDhcT0jCxRRVIrMTFhP","cdn-loop":"cloudflare; loops=1","cf-ipcountry":"SG","authorization":"REDACTED","x-forwarded-for":"35.187.226.205","host":"gateway.chotot.org","content-type":"application/json","cf-visitor":"{\"scheme\":\"https\"}","cache-control":"no-cache","accept-encoding":"gzip, br","postman-token":"c1c1f04f-415d-435a-9b56-b459ac10d68d","cf-connecting-ip":"35.187.226.205","accept":"*/*","content-length":"338","x-forwarded-proto":"https","cf-ray":"8ea0b8314f38604d-SIN","user-agent":"PostmanRuntime/7.39.1"},"size":1373,"url":"https://gateway.chotot.org:443/v2/private/cart/services?cart_id=oneclick","tls":{"cipher":"TLS_AES_128_GCM_SHA256","client_verify":"NONE","version":"TLSv1.3"}},"response":{"status":200,"size":479,"body":"{\"success\":[{\"target\":1138330,\"type\":\"bump\",\"params\":{\"duration\":1,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138330,\"type\":\"sticky_ad\",\"params\":{\"duration\":3,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138330,\"type\":\"special_display\",\"params\":{\"duration\":14,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138330,\"type\":\"bundle\",\"params\":{\"ad_id\":1138330,\"ad_type\":\"let\",\"bundle_id\":\"151\",\"category_id\":1030,\"region_id\":12000,\"segment_id\":\"all\",\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"}],\"fail\":[]}\n","headers":{"vary":"Origin","connection":"close","x-correlation-id":"3e17e6c035544d929535e4e6e6041766","content-encoding":"gzip","via":"kong/2.8.0-d648489b6","x-kong-proxy-latency":"2","content-type":"application/json; charset=UTF-8","x-kong-upstream-latency":"103","date":"Fri, 29 Nov 2024 07:00:13 GMT","access-control-expose-headers":"X-Total-Count"}},"client_ip":"35.187.226.205","started_at":1732863613661,"upstream_uri":"/private/cart/services?cart_id=oneclick"}`,
		`{"started_at":1732863617097,"upstream_uri":"/private/cart/services?cart_id=oneclick","client_ip":"35.187.226.205","response":{"body":"{\"success\":[{\"target\":1138331,\"type\":\"bump\",\"params\":{\"duration\":1,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138331,\"type\":\"sticky_ad\",\"params\":{\"duration\":1,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138331,\"type\":\"special_display\",\"params\":{\"duration\":14,\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"},{\"target\":1138331,\"type\":\"bundle\",\"params\":{\"ad_id\":1138331,\"ad_type\":\"let\",\"bundle_id\":\"153\",\"category_id\":1030,\"region_id\":12000,\"segment_id\":\"all\",\"user_type\":\"private\"},\"price\":0,\"priceUnit\":null,\"error\":\"\"}],\"fail\":[]}\n","status":200,"headers":{"connection":"close","x-correlation-id":"f716556a198a4edabea8d7acf3dcd141","content-type":"application/json; charset=UTF-8","access-control-expose-headers":"X-Total-Count","vary":"Origin","content-encoding":"gzip","x-kong-upstream-latency":"276","via":"kong/2.8.0-d648489b6","x-kong-proxy-latency":"2","date":"Fri, 29 Nov 2024 07:00:17 GMT"},"size":477},"request":{"body":"{\"services\":[{\"target\":1138331,\"type\":\"bump\",\"params\":{\"duration\":1}},{\"target\":1138331,\"type\":\"sticky_ad\",\"params\":{\"duration\":1}},{\"target\":1138331,\"type\":\"special_display\",\"params\":{\"duration\":14}},{\"target\":1138331,\"type\":\"bundle\",\"params\":{\"ad_id\":1138331,\"bundle_id\":\"153\",\"category_id\":1030,\"region_id\":12000,\"segment_id\":\"all\"}}]}","url":"https://gateway.chotot.org:443/v2/private/cart/services?cart_id=oneclick","querystring":{"cart_id":"oneclick"},"tls":{"version":"TLSv1.3","client_verify":"NONE","cipher":"TLS_AES_128_GCM_SHA256"},"headers":{"x-forwarded-proto":"https","host":"gateway.chotot.org","cf-ipcountry":"SG","x-forwarded-for":"35.187.226.205","cf-ray":"8ea0b846ccf391bd-SIN","content-length":"338","cf-visitor":"{\"scheme\":\"https\"}","postman-token":"d328a2ee-b322-4e64-8af6-1e5880f68020","cf-connecting-ip":"35.187.226.205","accept-encoding":"gzip, br","x-chotot-id-key":"L6bwKf6aPC6ejBDhcT0jCxRRVIrMTFhP","cache-control":"no-cache","accept":"*/*","cdn-loop":"cloudflare; loops=1","authorization":"REDACTED","content-type":"application/json","user-agent":"PostmanRuntime/7.39.1"},"uri":"/v2/private/cart/services?cart_id=oneclick","method":"POST","size":1366}}`,
		`{"upstream_uri":"/private/cart/info?cart_id=oneclick","request":{"tls":{"client_verify":"NONE","cipher":"TLS_AES_128_GCM_SHA256","version":"TLSv1.3"},"url":"https://gateway.chotot.org:443/v2/private/cart/info?cart_id=oneclick","headers":{"content-type":"application/json","cache-control":"no-cache","cf-ew-via":"15","host":"gateway.chotot.org","x-chotot-id-key":"L6bwKf6aPC6ejBDhcT0jCxRRVIrMTFhP","cdn-loop":"cloudflare; loops=1; subreqs=1","cf-ipcountry":"SG","x-forwarded-for":"35.187.226.205,35.187.226.205","cf-ray":"8ea0b9f8935c604d-SIN","authorization":"REDACTED","cf-visitor":"{\"scheme\":\"https\"}","accept":"*/*","x-forwarded-proto":"https","postman-token":"4ea7109f-210a-45c9-b32d-9de52e27b027","accept-encoding":"gzip","cf-connecting-ip":"35.187.226.205","user-agent":"PostmanRuntime/7.39.1"},"querystring":{"cart_id":"oneclick"},"body":"","size":1217,"method":"GET","uri":"/v2/private/cart/info?cart_id=oneclick"},"response":{"body":"{\"services\":[{\"target\":1138335,\"type\":\"sticky_ad\",\"params\":{\"duration\":1,\"user_type\":\"private\"},\"price\":31000,\"priceUnit\":{\"vnd\":31000,\"credit\":31000,\"promotion\":31000},\"error\":\"\"},{\"target\":1138335,\"type\":\"bundle\",\"params\":{\"ad_id\":1138335,\"ad_type\":\"let\",\"bundle_id\":\"155\",\"category_id\":1050,\"region_id\":12000,\"segment_id\":\"all\",\"user_type\":\"private\"},\"price\":135000,\"priceUnit\":{\"vnd\":135000,\"credit\":135000,\"promotion\":161000},\"error\":\"\"},{\"target\":1138335,\"type\":\"special_display\",\"params\":{\"duration\":30,\"user_type\":\"private\"},\"price\":144000,\"priceUnit\":{\"vnd\":144000,\"credit\":144000,\"promotion\":240000},\"error\":\"\"},{\"target\":1138335,\"type\":\"3days_bump\",\"params\":{\"duration\":3,\"user_type\":\"private\"},\"price\":43000,\"priceUnit\":{\"vnd\":43000,\"credit\":43000,\"promotion\":43000},\"error\":\"\"}],\"show_token\":{\"atm\":{\"123pay\":1,\"napas\":1,\"zalopay\":1},\"visa\":{\"123pay\":1,\"napas\":1,\"zalopay\":1}},\"campaigns\":null}\n","size":595,"status":200,"headers":{"access-control-expose-headers":"X-Total-Count","connection":"close","content-type":"application/json; charset=UTF-8","vary":"Origin","x-correlation-id":"aa23fca6928f42ff95db1c5e50fffd9c","x-kong-proxy-latency":"7","date":"Fri, 29 Nov 2024 07:01:26 GMT","x-kong-upstream-latency":"298","via":"kong/2.8.0-d648489b6","content-encoding":"gzip"}},"started_at":1732863686524,"client_ip":"35.187.226.205"}`,
	}
	for _, sample := range samplesForFetchData {
		ctx := context.Background()
		err := fetchdataUC.ProcessLogLine(ctx, []byte(sample))
		require.Nil(t, err)
	}
	fmt.Println("fetch data done")
	buildStructureUC := NewBuildStructureUC(
		conf,
		mongoStorage,
	)
	ctx := context.Background()
	buildStructureUC.BuildStructure(ctx)
}
