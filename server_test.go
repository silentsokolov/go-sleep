package main

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestSleepDuration(t *testing.T) {
	var timeTable = []struct {
		in  int64
		out time.Duration
	}{
		{100, time.Duration(100) * time.Second},
		{0, defaultSleepAfter},
		{-1, time.Duration(math.MaxInt64)},
	}

	for _, test := range timeTable {
		if s := sleepDuration(test.in); s != test.out {
			t.Errorf("sleepDuration returned %v, want %v", s, test.out)
		}
	}
}

func TestParserBasicUsers(t *testing.T) {
	users := []string{"test:$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51", "user:test:test"}

	usersMap, err := parserBasicUsers(users)
	if err != nil {
		t.Fatalf("parserBasicUsers returned unexpected error: %v", err)
	}

	want := map[string]string{
		"test": "$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51",
		"user": "test:test",
	}
	if !reflect.DeepEqual(usersMap, want) {
		t.Errorf("parserBasicUsers returned %+v, want %+v", usersMap, want)
	}
}

func TestServerRoute_SecretBasic(t *testing.T) {
	route := &serverRoute{}
	route.basicUsers = map[string]string{
		"test": "$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51",
		"user": "test:test",
	}

	var userTable = []struct {
		in  string
		out string
	}{
		{"test", "$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51"},
		{"unknown", ""},
	}

	for _, test := range userTable {
		if s := route.secretBasic(test.in, ""); s != test.out {
			t.Errorf("serverRoute.secretBasic returned %v, want %v", s, test.out)
		}
	}
}
