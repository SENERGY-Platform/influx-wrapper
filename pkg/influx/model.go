package influx

type TimeValuePair struct {
	Time  string      `json:"time"`
	Value interface{} `json:"value"`
}

type MeasurementColumnPair struct {
	Measurement string `json:"measurement"`
	ColumnName  string `json:"columnName"`
}
