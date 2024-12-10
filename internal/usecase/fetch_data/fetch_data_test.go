package fetchdata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchDataUC_processLogLine(t *testing.T) {
	t.Parallel()

	fetchDataUC := NewFetchDataUC(
		conf,
		mongoStorage,
	)

	test := []struct {
		name    string
		logLine string
		wantErr bool
	}{
		{
			name:    "success_01",
			logLine: `{"response":{"size":961,"headers":{"vary":"Origin","access-control-expose-headers":"X-Total-Count","content-type":"application/json; charset=UTF-8","via":"kong/2.8.0-d648489b6","x-kong-proxy-latency":"28","x-kong-upstream-latency":"7","transfer-encoding":"chunked","x-correlation-id":"47d9b07ea6254e5298ad2acea67f619e","connection":"close","date":"Fri, 29 Nov 2024 07:00:03 GMT","content-encoding":"gzip"},"status":200,"body":"{\"data\":[{\"id\":1302748,\"hash\":\"be6f7df3a8f03d7457300fe982938e05\",\"phone\":\"0373601207\",\"order_id\":0,\"account_id\":10963931,\"created_user\":\"admin-ds@chotot.vn\",\"approved_user\":\"\",\"package_id\":561,\"amount\":20000,\"amount_dongtot\":20000,\"bank_id\":\"\",\"transaction_ref\":\"\",\"info\":\"\",\"created_date\":1732863602,\"modified_date\":1732863602,\"is_approval\":false,\"expiration_time\":\"2024-11-30T16:59:59Z\",\"is_expired\":false,\"contract_info\":\"{\\\"account_id\\\":10963931,\\\"bt_order_id\\\":0,\\\"contract_number\\\":\\\"10963931\\\",\\\"contract_date\\\":\\\"\\\",\\\"card_id\\\":\\\"\\\",\\\"permanent_residence\\\":\\\"\\\",\\\"announce_id\\\":\\\"67496672434804201147a165\\\",\\\"payment_method\\\":\\\"\\\",\\\"payment_method_order\\\":\\\"\\\"}\",\"contract_invoice\":\"{\\\"company_name\\\":\\\"\\\",\\\"company_address\\\":\\\"\\\",\\\"tax_id\\\":\\\"\\\",\\\"email\\\":\\\"\\\"}\",\"contract_status\":\"waiting\",\"contract_status_text\":\"Đợi hợp đồng\",\"is_online_contract\":true,\"expiration_type\":\"fixed-date\",\"usage_day\":0,\"is_new_econtract\":false,\"is_read\":true}],\"limit\":20,\"offset\":0}\n"},"upstream_uri":"/private/bank_transfer/contract-history?limit=20&page=0","request":{"body":"","querystring":{"limit":"20","page":"0"},"size":1400,"tls":{"version":"TLSv1.3","cipher":"TLS_AES_256_GCM_SHA384","client_verify":"NONE"},"uri":"/v1/private/bank_transfer/contract-history?limit=20&page=0","headers":{"host":"gateway.chotot.org","x-chotot-id-key":"L6bwKf6aPC6ejBDhcT0jCxRRVIrMTFhP","content-type":"application/json; charset=UTF-8","accept-encoding":"gzip,deflate","authorization":"REDACTED","connection":"Keep-Alive","user-agent":"Apache-HttpClient/4.5.3 (Java/11.0.16)","accept":"*/*"},"method":"GET","url":"https://gateway.chotot.org:443/v1/private/bank_transfer/contract-history?limit=20&page=0"},"started_at":1732863603503,"client_ip":"10.9.24.207"}`,
			wantErr: false,
		},
	}

	for _, tt := range test {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			err := fetchDataUC.ProcessLogLine(ctx, []byte(tt.logLine))
			if tt.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
