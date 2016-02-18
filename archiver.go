/*
archiver.go : A archiver listen a data channel to get Metric and send them
to an influx database

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
	"fmt"
	"log"

	"github.com/influxdata/influxdb/client/v2"
)

// ArchiverConfig contains parameters to link archiver with influxdb
type ArchiverConfig struct {
	InfluxURL  string
	InfluxDB   string
	InfluxUser string
	InfluxPwd  string
}

// Archiver get data from checkHttp and send them to selected output
type Archiver struct {
	Config ArchiverConfig
	Clnt   client.Client
}

// NewArchiver returns a archiver for the specified output
func NewArchiver(conf ArchiverConfig) *Archiver {
	arch := &Archiver{
		Config: conf,
	}

	return arch
}

// writeInfluxPoints write Metric into influxdb (Metrics are sended through
// archiver connection with influxdb : arch.Clnt)
func (arch *Archiver) writeInfluxPoints(value Metric) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "http_request",
		Precision: "s",
	})

	tags := map[string]string{
		"website": value.Web.Label,
	}

	fields := map[string]interface{}{
		"status":        value.StatusCode,
		"response_time": value.ResponseTime.Seconds(),
	}

	point, _ := client.NewPoint("request_result", tags, fields, value.Ts)
	bp.AddPoint(point)

	err := arch.Clnt.Write(bp)
	if err != nil {
		log.Printf("Can not write data into influxdb : %s", err)
	}
}

// Run the archiver
func (arch *Archiver) Run(data chan Metric) {
	var err error
	arch.Clnt, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     arch.Config.InfluxURL,
		Username: arch.Config.InfluxUser,
		Password: arch.Config.InfluxPwd,
	})
	if err != nil {
		log.Println("Archiver : can not contact influx")
	}

	fmt.Println("Archiver start to listen data chan")
	for {
		arch.writeInfluxPoints(<-data)
	}
}
