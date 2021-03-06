package influx

import (
	"errors"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/tests/services"
	influxLib "github.com/orourkedd/influxdb1-client"
	"github.com/orourkedd/influxdb1-client/models"
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	influxClientMock := services.NewClientMock()

	influxClient := Influx{
		config: &configuration.ConfigStruct{
			Debug: true,
		},
		client: &influxClientMock,
	}

	t.Run("generateQuery", func(t *testing.T) {
		t.Run("empty set", func(t *testing.T) {
			q := generateQuery(uniqueMeasurementsColumns{})
			expect := "SELECT  FROM "
			if q != expect {
				t.Error("expect", expect, "actual", q)
			}
		})
		t.Run("empty strings", func(t *testing.T) {
			measurements := make(map[string]struct{})
			measurements[""] = struct{}{}
			columns := make(map[string]map[string]struct{})
			columns[""] = make(map[string]struct{})
			columns[""][""] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Measurements: measurements,
				Columns:      columns,
			})
			expect := "SELECT  FROM "
			if q != expect {
				t.Error("expect", expect, "actual", q)
			}
		})
		t.Run("empty measurements", func(t *testing.T) {
			columns := make(map[string]map[string]struct{})
			columns["c1"] = make(map[string]struct{})
			columns["c1"][""] = struct{}{}
			columns["c2"] = make(map[string]struct{})
			columns["c2"][""] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Columns: columns,
			})
			expect := "SELECT \"c1\", \"c2\" FROM "
			expectAlt := "SELECT \"c2\", \"c1\" FROM "
			if q != expect && q != expectAlt {
				t.Error("\nexpect\n", expect, "\nor\n", expectAlt, "\nactual\n", q)
			}
		})
		t.Run("empty columns", func(t *testing.T) {
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Measurements: measurements,
			})
			expect := "SELECT  FROM \"m1\", \"m2\""
			expectAlt := "SELECT  FROM \"m2\", \"m1\""
			if q != expect && q != expectAlt {
				t.Error("expect\n", expect, "\nor\n", expectAlt, "\nactual\n", q)
			}
		})
		t.Run("normal set", func(t *testing.T) {
			columns := make(map[string]map[string]struct{})
			columns["c1"] = make(map[string]struct{})
			columns["c1"][""] = struct{}{}
			columns["c2"] = make(map[string]struct{})
			columns["c2"][""] = struct{}{}
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Columns:      columns,
				Measurements: measurements,
			})
			validResults := []string{}
			validResults = append(validResults,
				"SELECT \"c1\", \"c2\" FROM \"m1\", \"m2\"",
				"SELECT \"c2\", \"c1\" FROM \"m1\", \"m2\"",
				"SELECT \"c1\", \"c2\" FROM \"m2\", \"m1\"",
				"SELECT \"c2\", \"c1\" FROM \"m2\", \"m1\"")

			foundValid := false

			for _, validResult := range validResults {
				if q == validResult {
					foundValid = true
				}
			}
			if !foundValid {
				t.Error("expect any of\n", strings.Join(validResults, "\n"), "\nactual\n", q)
			}
		})
		t.Run("normal set with math", func(t *testing.T) {
			columns := make(map[string]map[string]struct{})
			columns["c1"] = make(map[string]struct{})
			columns["c1"]["+3"] = struct{}{}
			columns["c2"] = make(map[string]struct{})
			columns["c2"]["-5"] = struct{}{}
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Columns:      columns,
				Measurements: measurements,
			})
			validResults := []string{}
			validResults = append(validResults,
				"SELECT \"c1\"+3 AS \"c1+3\", \"c2\"-5 AS \"c2-5\" FROM \"m1\", \"m2\"",
				"SELECT \"c2\"-5 AS \"c2-5\", \"c1\"+3 AS \"c1+3\" FROM \"m1\", \"m2\"",
				"SELECT \"c1\"+3 AS \"c1+3\", \"c2\"-5 AS \"c2-5\" FROM \"m2\", \"m1\"",
				"SELECT \"c2\"-5 AS \"c2-5\", \"c1\"+3 AS \"c1+3\" FROM \"m2\", \"m1\"")

			foundValid := false

			for _, validResult := range validResults {
				if q == validResult {
					foundValid = true
				}
			}
			if !foundValid {
				t.Error("expect any of\n", strings.Join(validResults, "\n"), "\nactual\n", q)
			}
		})
	})

	t.Run("executeQuery", func(t *testing.T) {
		t.Run("net error", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, netError{
				error: errors.New("net error"),
			})
			_, err := influxClient.ExecuteQuery("test", "test")
			if err != ErrInfluxConnection {
				t.Fail()
			}
		})
		t.Run("other err", func(t *testing.T) {
			testErr := errors.New("other err")
			influxClientMock.SetQueryResponse(nil, testErr)
			_, err := influxClient.ExecuteQuery("test", "test")
			if err != testErr {
				t.Fail()
			}
		})
		t.Run("response nil", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, nil)
			_, err := influxClient.ExecuteQuery("test", "test")
			if err != ErrNULL {
				t.Fail()
			}
		})
		t.Run("response not found", func(t *testing.T) {
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Err: errors.New("DB test not found"),
			}, nil)
			_, err := influxClient.ExecuteQuery("test", "test")
			if err != ErrNotFound {
				t.Fail()
			}
		})
		t.Run("response other found", func(t *testing.T) {
			testErr := errors.New("unknown error message")
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Err: testErr,
			}, nil)
			_, err := influxClient.ExecuteQuery("test", "test")
			if err != testErr {
				t.Fail()
			}
		})
		t.Run("response normal", func(t *testing.T) {
			expect := &influxLib.Response{
				Results: []influxLib.Result{
					{
						Series: []models.Row{
							{
								Name: "test",
							},
						},
					},
				},
			}
			influxClientMock.SetQueryResponse(expect, nil)
			actual, err := influxClient.ExecuteQuery("test", "test")
			if err != nil {
				t.Fail()
			}
			if actual != expect {
				t.Fail()
			}
		})
	})
}
