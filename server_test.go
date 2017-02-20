package main

import (
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestResponseJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := pageContext{Message: "Test message"}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseJSON(w, context, http.StatusBadRequest)
	})

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("responseJSON returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `{"message":"Test message"}`
	if recorder.Body.String() != expected {
		t.Errorf("responseJSON returned unexpected body: got %v want %v", recorder.Body.String(), expected)
	}
}

func TestGetDefaultSleepAfter(t *testing.T) {
	var timeTable = []struct {
		in  int64
		out time.Duration
	}{
		{100, time.Duration(100) * time.Second},
		{0, defaultSleepAfter},
		{-1, time.Duration(math.MaxInt64)},
	}

	for _, test := range timeTable {
		if s := getDefaultSleepAfter(test.in); s != test.out {
			t.Errorf("getDefaultSleepAfter returned %v, want %v", s, test.out)
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
