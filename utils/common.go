package utils

import (
	"best-ticker/config"
	"math"
	"strings"
	"time"
)

func ConvertToOkxSpotInstID(okxFuturesInstID string) string {
	// BTC-USDT-SWAP => BTC-USDT
	return strings.Replace(okxFuturesInstID, "-SWAP", "", -1)
}

func ConvertToStdInstType(exchange config.Exchange, instType string) config.InstrumentType {
	if exchange == config.OkxExchange {
		switch instType {
		case "SWAP":
			return config.FuturesInstrument
		case "SPOT":
			return config.SpotInstrument
		}
	}
	return config.UnknownInstrument
}

func MaxFloat64(list []float64) (max float64) {
	max = list[0]
	for _, v := range list {
		if v > max {
			max = v
		}
	}
	return
}

func MinFloat64(list []float64) (min float64) {
	min = list[0]
	for _, v := range list {
		if v < min {
			min = v
		}
	}
	return
}

func Round(value float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	rounded := math.Round(value*shift) / shift
	return rounded
}

func InArray(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}

func GetTimestampInMS() int64 {
	return time.Now().UnixNano() / 1e6
}
