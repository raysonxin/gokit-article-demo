package main

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

func decodeArithmeticRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars()
	requestType, ok := vars["type"]
}
