package main 

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const serviceUrl = "https://appsitory.com/updater.json?method=update"

func Update(ver string, apps string, uid string) (count int, link string) {
	encodedString := base64.StdEncoding.EncodeToString([]byte(apps))
	fmt.Println("requesting Update service...")
    response, err := http.PostForm(serviceUrl, url.Values{
		"winver": {ver},
		"data": {encodedString},
		"uid": {uid},
	})
	if err != nil {
		fmt.Println(err)
		return 0, ""
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	sbody := string(body)
	//fmt.Println(sbody)
	//{"count":0,"link":null}
	//{"count":2,"link":"https:\/\/appsitory.com\/update\/?id=ea662981f96b3cf05290934bccf689f9"}
	// var count int
	// var link string
	fmt.Sscanf(sbody, "{\"count\":%d,\"link\":\"%s\"}", &count, &link)
	if count == 0 {
		fmt.Println("no updates found")
		return 0, ""
	} else {
		link = strings.TrimSuffix(link, "}")
		link = strings.TrimSuffix(link, "\"")
		link = strings.ReplaceAll(link, "\\/", "/")
		fmt.Printf("updates count: %d, link: %s", count, link)
		return count, link
	}
}