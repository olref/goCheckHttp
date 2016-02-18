/*
checkHttp is a simple piece of code to check websites availability.

Usage: checkHttp [options] urls...
With options :
-interval : interval between two checks of a website
-nbArchiver : number of archivers charged to send data into influxdb

Args:
urls : a list of url you want to check. I you have already defined URLs into
the config.toml file urls args are merged with config urls

https://github.com/olref/goCheckHttp

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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type tomlConfig struct {
	General generalConfig
	Influx  ArchiverConfig
}

type generalConfig struct {
	urls []string
}

// RemoveDuplicates remove duplicate string from a slice of string
// see : https://groups.google.com/d/msg/golang-nuts/-pqkICuokio/ZfSRfU_CdmkJ
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func main() {
	// define command paramters
	var interval = flag.Int("interval", 60, "Check periodicity in seconds")
	var nbArchiver = flag.Int("nbArchiver", 1, "Number of archivers")
	var nbCollector = flag.Int("nbCollector", 5, "Number of collectors")
	//var configFile = flag.String("conf", "config.toml", "File of URLs (one URL by line)")

	// define command usage
	flag.Usage = func() {
		fmt.Printf("Usage: checkHttp [options] [url1 url2 url3 ...]\n")
		fmt.Printf("  If you don't pass a URL as argument, you must specify a urls list in your config file\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	urlList := flag.Args()

	// load config file
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	archConf := ArchiverConfig{
		InfluxURL:  viper.GetString("influx.influxurl"),
		InfluxDB:   viper.GetString("influx.influxdb"),
		InfluxUser: viper.GetString("influx.influxuser"),
		InfluxPwd:  viper.GetString("influx.influxpwd"),
	}

	confURLList := viper.GetStringSlice("general.urls")

	// if we need at least one url
	if len(urlList) == 0 && len(confURLList) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// concatenate urls from args and from config file
	urlList = append(urlList, confURLList...)

	// remove duplicated line
	RemoveDuplicates(&urlList)

	var websiteList []Website
	for _, url := range urlList {
		lab, u := DecomposeURL(url)
		websiteList = append(websiteList, Website{Label: lab, URL: u})
	}

	// manage os shutdown signals
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// define the main ticker
	ticker := time.NewTicker(time.Duration(*interval) * time.Second)

	// create a channel to distribute the work among the collectors
	work := make(chan Website, 100)

	// create a channel of communication between collectors and archivers
	data := make(chan Metric, 100)

	//create archivers to send data to output
	for i := 0; i < *nbArchiver; i++ {
		arch := NewArchiver(archConf)

		// launch a archiver to send data to influxdb
		go arch.Run(data)
	}

	//create collector to check websites
	for i := 0; i < *nbCollector; i++ {
		col := NewCollector()

		// launch a collector to check website
		go col.Run(work, data)
	}

	for {
		for _, web := range websiteList {
			work <- web
		}

		select {
		case <-ticker.C:
			continue

		case <-shutdown:
			fmt.Println("See you soon !")
			os.Exit(0)
		}
	}
}
