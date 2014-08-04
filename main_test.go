package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestContentLength(t *testing.T) {

	resp, err := http.Get("http://google.com/")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()

	fmt.Println(resp.ContentLength)
	if resp.ContentLength < 1 {
		t.Error("expected > 1 , got:", resp.ContentLength)
	}

}

func TestBodyContentLength(t *testing.T) {

	resp, _ := http.Get("http://google.com/")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(len(body))
	if len(body) < 1 {
		t.Error("expected > 1 , got:", len(body))
	}
}
