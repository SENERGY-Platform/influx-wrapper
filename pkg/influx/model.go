package influx

type TimeValuePair struct {
	Time  string      `json:"time"`
	Value interface{} `json:"value"`
}
