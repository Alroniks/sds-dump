package main

import (
    "os"
    "fmt"
    "crypto/tls"
    "time"
    "strings"
    "io/ioutil"
    "encoding/json"
    "sync"
    "strconv"
    "github.com/PuerkitoBio/goquery"
    "github.com/djimenez/iconv-go"
    "net/http"
)

var wg sync.WaitGroup

type Product struct {

    ID int `json:"id"`
    Article string `json:"article"`
    Title string `json:"title"`
    Image string `json:"image"`
    Brand string `json:"brand"`
    Price string `json:"price"`
    Units string `json:"units"`
    InPack string `json:"inpack"`
    Description string `json:"description"`
    Availability int `json:"availability"`
    Category int `json:"category"`

}

type Video struct {
    Product int `json:"id"`
    Video string `json:"video"`
}

var tube chan Video

const PRODUCT_LINK_TEMPLATE = "https://www.sds-group.ru/items_%s.htm"

func main() {

    file, err := os.Open("resources/output.json");

    if err != nil {
        fmt.Println(err)
    }
    defer file.Close();

    bytes, _ := ioutil.ReadAll(file)

    var result = []Product{}

    json.Unmarshal([]byte(bytes), &result)

    tube = make(chan Video)

    for _, item := range result {
        wg.Add(1)
        go fetch(item.ID)
        // fmt.Println(item)
        // go fetch(18751)
    }
    
    go func() {
        wg.Wait()
        close(tube)
    }()    

    var out []Video = []Video{}

    for video := range tube {
        out = append(out, video)
    }

    json, _ := json.MarshalIndent(out, "", "  ")

    ioutil.WriteFile("resources/videos.json", json, 0644)

}

func fetch(id int) {
    defer wg.Done()

    if _, err := os.Stat(fmt.Sprintf("data/spc/%s.html", strconv.Itoa(id))); !os.IsNotExist(err) {
        // skipping existing
        return
    }


    link := fmt.Sprintf(PRODUCT_LINK_TEMPLATE, strconv.Itoa(id))

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    }

    client := &http.Client{
        Transport: transport,
        Timeout: 50 * time.Second,
    }

    response, err := client.Get(link)

    if err != nil {
        fmt.Println(err);
        return
    }

    defer response.Body.Close()

    bodyInUnicode, _ := iconv.NewReader(response.Body, "windows-1251", "utf-8")

    document, _ := goquery.NewDocumentFromReader(bodyInUnicode)

    spec, _ := document.Find("div#tab-techs").First().Html();

    ioutil.WriteFile(fmt.Sprintf("data/spc/%s.html", strconv.Itoa(id)), []byte(spec), 0644)

    video := document.Find("div#tab-video").First().Find("iframe").First().AttrOr("src", "")

    video = strings.Replace(video, "embed/", "watch?v=", 1)
    video = strings.Replace(video, "?rel=0", "", 1)
    
    tube <- Video{
        Product: id,
        Video: video,
    }
}