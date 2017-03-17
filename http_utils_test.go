package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := pageContext{Message: "Test message"}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseJSON(w, http.StatusBadRequest, context)
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

func TestResponseHTML(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	context := pageContext{Message: "Test message"}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseHTML(w, http.StatusBadRequest, "wait.html", context)
	})

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("responseHTML returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := `"Test message"`
	if strings.Contains(recorder.Body.String(), expected) {
		t.Errorf("responseHTML returned unexpected body: got %v want %v", recorder.Body.String(), expected)
	}
}
