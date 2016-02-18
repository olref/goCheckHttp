/*
collector.go : A collector listen a work channel to get a Website to check.
It makes a get on the website Url and create a Metric instance with the
result of the request. Then he push the metric on the data channel for the
collectors

The MIT License (MIT)

Copyright (c) 2016 olref

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Collector request website and push result on a data channel
type Collector struct {
	Counter int // counter of treated request
}

// NewCollector returns a collecctor
func NewCollector() *Collector {
	col := &Collector{
		Counter: 0,
	}

	return col
}

// Run the collector
func (col *Collector) Run(work chan Website, data chan Metric) {
	fmt.Println("Collector start to listen data chan")
	for {
		met, err := col.callURL(<-work)
		if err == nil {
			data <- met
		}
	}
}

func (col *Collector) callURL(webSite Website) (Metric, error) {
	// add jitter (max 5 seconds)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	jitter := time.Duration(r.Intn(5)) * time.Second
	time.Sleep(jitter)

	fmt.Println("call : " + webSite.Label)
	tr := &http.Transport{
		// disable ssl certificate check
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	timeout := time.Duration(60 * time.Second)
	client := &http.Client{Transport: tr, Timeout: timeout}

	t0 := time.Now()
	resp, err := client.Get(webSite.URL)
	t1 := time.Now()
	duration := t1.Sub(t0)

	if err != nil {
		log.Printf("Failed to get data from %s, error : '%s' \n", webSite.URL, err)
		return Metric{}, errors.New("no data")
	}

	return Metric{
		Web:          webSite,
		StatusCode:   resp.StatusCode,
		Ts:           time.Now(),
		ResponseTime: duration,
	}, nil
}
