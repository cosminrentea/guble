package metrics

import (
	"expvar"
	"strconv"
	"time"
)

func SetRate(m Map, key string, value expvar.Var, timeframe, scale time.Duration) {
	if value != nil {
		v, _ := strconv.ParseInt(value.String(), 10, 64)
		m.Set(key, newRate(v, timeframe, scale))
	} else {
		m.Set(key, zeroValue)
	}
}

func SetAverage(m Map, key string, totalVar, casesVar expvar.Var, scale int64, defaultValue string) {
	if totalVar != nil && casesVar != nil {
		total, _ := strconv.ParseInt(totalVar.String(), 10, 64)
		cases, _ := strconv.ParseInt(casesVar.String(), 10, 64)
		m.Set(key, newAverage(total, cases, scale, defaultValue))
	} else {
		m.Set(key, zeroValue)
	}
}

func AddToMaps(key string, value int64, maps ...Map) {
	for _, m := range maps {
		m.Add(key, value)
	}
}
