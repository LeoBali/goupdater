package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const serviceUrl = "https://appsitory.com/updater.json?method=update"

func Update(ver string, apps string, uid string) (count int, link string) {
	//fmt.Println(apps)
	encodedString := base64.StdEncoding.EncodeToString([]byte(apps))
	//fmt.Println(encodedString)
	log.Println("requesting Update service...")
    response, err := http.PostForm(serviceUrl, url.Values{
		"winver": {ver},
		"data": {encodedString},
		"uid": {uid},
	})
	if err != nil {
		log.Println(err)
		return 0, ""
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	sbody := string(body)
	log.Println(sbody)
	fmt.Sscanf(sbody, "{\"count\":%d,\"link\":\"%s\"}", &count, &link)
	if count == 0 {
		log.Println("no updates found")
		return 0, ""
	} else {
		link = strings.TrimSuffix(link, "}")
		link = strings.TrimSuffix(link, "\"")
		link = strings.ReplaceAll(link, "\\/", "/")
		log.Printf("updates count: %d, link: %s", count, link)
		return count, link
	}
}