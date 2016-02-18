/*
Website.go : Website structure and all method used to make http calls


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
	"log"
	"strings"
)

// Website checked by checkHttp.go
type Website struct {
	Label string // used to identify the website into influxdb
	URL   string // url to check
}

// DecomposeURL split the urls defined into config file as label / URL.
// A URL can have two format : label::URL or only URL. In the first case,
// DecomposeURL returns label, URL, in the second case it returns
// URL, URL (URL will be used as a label)
func DecomposeURL(URL string) (string, string) {

	// init label and url with the whole URL (URL without label case)
	var label = URL
	var url = URL

	if strings.Index(URL, "::") >= 0 {
		urlElem := strings.Split(URL, "::")
		if len(urlElem) == 2 {
			label = urlElem[0]
			url = urlElem[1]
		} else {
			// too much part in this url => not valid for goCheckHttp
			log.Printf("Bad Url %s", URL)
		}
	}
	return label, url
}
