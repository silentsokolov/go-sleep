package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(indexHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("indexHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `OK`
	if recorder.Body.String() != expected {
		t.Errorf("indexHandler returned unexpected body: got %v want %v", recorder.Body.String(), expected)
	}
}
