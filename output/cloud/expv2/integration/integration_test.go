package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/klauspost/compress/snappy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/lib/testutils"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output/cloud/expv2"
	"go.k6.io/k6/output/cloud/expv2/pbcloud"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/guregu/null.v3"
)

// This test runs an integration tests for the Output cloud.
// It only calls public API of the Output and
// it implements a concrete http endpoint where to get
// the protobuf flush requests.

func TestOutputFlush(t *testing.T) {
	// TODO: it has 3s for aggregation time
	// then it means it will execute for +3s that it is a waste of time
	// because it isn't really required.
	// Reduce the aggregation time (to 1s?) then run always.
	t.Skip("Skip the integration test, if required enable it manually")
	t.Parallel()

	results := make(chan *pbcloud.MetricSet)
	ts := httptest.NewServer(metricsHandler(results))
	defer ts.Close()

	// init conifg
	c := cloudapi.NewConfig()
	c.Host = null.StringFrom(ts.URL)
	c.Token = null.StringFrom("my-secret-token")
	c.AggregationPeriod = types.NullDurationFrom(3 * time.Second)
	c.AggregationWaitPeriod = types.NullDurationFrom(1 * time.Second)

	// init and start the output
	o, err := expv2.New(testutils.NewLogger(t), c)
	require.NoError(t, err)
	o.SetReferenceID("my-test-run-id-123")
	require.NoError(t, o.Start())

	// collect and flush samples
	o.AddMetricSamples([]metrics.SampleContainer{
		testSamples(),
	})

	// wait for results
	mset := <-results
	close(results)
	assert.NoError(t, o.StopWithTestError(nil))

	// read and convert the json version
	// of the expected protobuf sent request
	var exp pbcloud.MetricSet
	expjs, err := os.ReadFile("./testdata/metricset.json") //nolint:forbidigo // ReadFile here is used in a test
	require.NoError(t, err)
	err = protojson.Unmarshal(expjs, &exp)
	require.NoError(t, err)

	msetjs, err := protojson.Marshal(mset)
	require.NoError(t, err)
	assert.JSONEq(t, string(expjs), string(msetjs))
}

func metricsHandler(results chan<- *pbcloud.MetricSet) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Token my-secret-token" {
			http.Error(rw, fmt.Sprintf("token is required; got %q", token), http.StatusUnauthorized)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		mset, err := metricSetFromRequest(b)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		results <- mset
	}
}

func metricSetFromRequest(b []byte) (*pbcloud.MetricSet, error) {
	b, err := snappy.Decode(nil, b)
	if err != nil {
		return nil, err
	}
	var mset pbcloud.MetricSet
	err = proto.Unmarshal(b, &mset)
	if err != nil {
		return nil, err
	}
	return &mset, nil
}

func testSamples() metrics.Samples {
	r := metrics.NewRegistry()
	m1 := r.MustNewMetric("metric_counter_1", metrics.Counter)
	m2 := r.MustNewMetric("metric_gauge_2", metrics.Gauge)
	m3 := r.MustNewMetric("metric_rate_3", metrics.Rate)
	m4 := r.MustNewMetric("metric_trend_4", metrics.Trend)

	samples := []metrics.Sample{
		{
			TimeSeries: metrics.TimeSeries{
				Metric: m1,
				Tags:   r.RootTagSet().With("my_label_1", "my_label_value_1"),
			},
			Time:  time.Date(2023, time.May, 1, 1, 0, 0, 0, time.UTC),
			Value: 42.2,
		},
		{
			TimeSeries: metrics.TimeSeries{
				Metric: m2,
				Tags:   r.RootTagSet().With("my_label_2", "my_label_value_2"),
			},
			Time:  time.Date(2023, time.May, 1, 2, 0, 0, 0, time.UTC),
			Value: 3.14,
		},
		{
			TimeSeries: metrics.TimeSeries{
				Metric: m3,
				Tags:   r.RootTagSet().With("my_label_3", "my_label_value_3"),
			},
			Time:  time.Date(2023, time.May, 1, 3, 0, 0, 0, time.UTC),
			Value: 2.718,
		},
		{
			TimeSeries: metrics.TimeSeries{
				Metric: m4,
				Tags:   r.RootTagSet().With("my_label_4", "my_label_value_4"),
			},
			Time:  time.Date(2023, time.May, 1, 4, 0, 0, 0, time.UTC),
			Value: 6,
		},
	}
	return samples
}
