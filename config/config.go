package config

import (
	"hash/fnv"
	"math"
)

func Apply(rules []Rule, ctx Context) map[string]interface{} {
	config := make(map[string]interface{})

	// apply all matching rules
	for _, rule := range rules {
		if rule.Matches(ctx) {
			config[rule.Key] = rule.Value
		}
	}

	return config
}

type Context struct {
	Version    int
	DeviceHash string
}

type Rule struct {
	Key           string
	Value         interface{}
	MinVersion    int
	MaxVersion    int
	MinPercentile float64
	MaxPercentile float64
}

func (r *Rule) Matches(ctx Context) bool {
	if r.MinVersion != 0 && ctx.Version < r.MinVersion {
		return false
	}

	if r.MaxVersion != 0 && ctx.Version > r.MaxVersion {
		return false
	}

	// if percentiles are not configured, this one matches!
	if r.MaxPercentile <= r.MinPercentile {
		return true
	}

	// check if we are in the configured percentile.
	value := uniqueRandomValue(ctx.DeviceHash, r.Key)
	return r.MinPercentile <= value && value <= r.MaxPercentile
}

func uniqueRandomValue(key, salt string) float64 {
	hash := fnv.New64a()
	hash.Write([]byte(key))
	hash.Write([]byte(salt))

	return float64(hash.Sum64()) / float64(math.MaxUint64)
}
