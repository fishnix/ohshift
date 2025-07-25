// Code generated by BobGen psql v0.38.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/bob/types"
)

var defaultFaker = faker.New()

func random_bool(f *faker.Faker, limits ...string) bool {
	if f == nil {
		f = &defaultFaker
	}

	return f.Bool()
}

func random_int32(f *faker.Faker, limits ...string) int32 {
	if f == nil {
		f = &defaultFaker
	}

	return f.Int32()
}

func random_int64(f *faker.Faker, limits ...string) int64 {
	if f == nil {
		f = &defaultFaker
	}

	return f.Int64()
}

func random_string(f *faker.Faker, limits ...string) string {
	if f == nil {
		f = &defaultFaker
	}

	val := strings.Join(f.Lorem().Words(f.IntBetween(1, 5)), " ")
	if len(limits) == 0 {
		return val
	}
	limitInt, _ := strconv.Atoi(limits[0])
	if limitInt > 0 && limitInt < len(val) {
		val = val[:limitInt]
	}
	return val
}

func random_time_Time(f *faker.Faker, limits ...string) time.Time {
	if f == nil {
		f = &defaultFaker
	}

	year := time.Hour * 24 * 365
	min := time.Now().Add(-year)
	max := time.Now().Add(year)
	return f.Time().TimeBetween(min, max)
}

func random_types_JSON_json_RawMessage_(f *faker.Faker, limits ...string) types.JSON[json.RawMessage] {
	if f == nil {
		f = &defaultFaker
	}

	s := &bytes.Buffer{}
	s.WriteRune('{')
	for i := range f.IntBetween(1, 5) {
		if i > 0 {
			fmt.Fprint(s, ", ")
		}
		fmt.Fprintf(s, "%q:%q", f.Lorem().Word(), f.Lorem().Word())
	}
	s.WriteRune('}')
	return types.NewJSON[json.RawMessage](s.Bytes())
}

func random_uuid_UUID(f *faker.Faker, limits ...string) uuid.UUID {
	if f == nil {
		f = &defaultFaker
	}

	return uuid.Must(uuid.NewV4())
}
