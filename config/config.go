package config

import (
	"hash/fnv"
	"math"
	"time"
)

var EuropeBerlin = mustLoadLocation("Europe/Berlin")

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
	Beta       bool
}

type Range struct {
	Min, Max float64
}

type Rule struct {
	Key   string
	Value interface{}

	// limit the version or user range this property applies to.
	Versions    []Range
	Percentiles []Range
	Times       []Range

	// if set, this rule only applies to beta users.
	Beta bool
}

func (r *Rule) Matches(ctx Context) bool {
	// check that all version restrictions match
	if !containsValueOrEmpty(r.Versions, float64(ctx.Version)) {
		return false
	}

	// check that the user is in the correct percentile
	value := uniqueRandomValue(ctx.DeviceHash, r.Key)
	if !containsValueOrEmpty(r.Percentiles, value) {
		return false
	}

	// check that the time matches
	now := time.Now().In(EuropeBerlin)
	hourOfDay := float64(now.Hour()) + float64(now.Minute())/60.0
	if !containsValueOrEmpty(r.Times, hourOfDay) {
		return false
	}

	// if beta is set, we want to apply the rule only if the request
	// comes from a user who has beta activates.
	if r.Beta {
		return ctx.Beta
	}

	// rule matches!
	return true
}

func (r Range) Contains(value float64) bool {
	return r.Min <= value && value < r.Max
}

func uniqueRandomValue(key, salt string) float64 {
	hash := fnv.New64a()
	hash.Write([]byte(key))
	hash.Write([]byte(salt))

	return float64(hash.Sum64()) / float64(math.MaxUint64)
}

func containsValueOrEmpty(ranges []Range, value float64) bool {
	for _, r := range ranges {
		if r.Contains(value) {
			return true
		}
	}

	return len(ranges) == 0
}

func mustLoadLocation(timezone string) *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}

	return loc
}
