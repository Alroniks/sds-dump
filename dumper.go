package main

import (
    "fmt"
    "sync"
    "strings"
    "net/http"
    "encoding/base64"
    "encoding/json"
    "net/url"
    "github.com/djimenez/iconv-go"
    "github.com/PuerkitoBio/goquery"
    "strconv"
)

var wg sync.WaitGroup

var tube chan Product

type Product struct {

    ID string `json:"id"`
    Article string `json:"article"`
    Title string `json:"title"`
    Link string `json:"link"` // ? не нужен же вроде
    Image string `json:"image"` // нужно скачивать файл в папку с id товара
    Brand string `json:"brand"`
    Price string `json:"price"`
    Units string `json:"units"`
    InPack string `json:"inpack"`
    Description string `json:"description"` // нужно писать в файл видимо
    Availability int `json:"availability"`

}

const CATEGORY_LINK_TEMPLATE = "https://www.sds-group.ru/catalog_table_%s.htm"

func main() {

    tube = make(chan Product)

    counter := 1204;

    for counter <= 1204 {
        wg.Add(1)
        go parse(fmt.Sprintf(CATEGORY_LINK_TEMPLATE, strconv.Itoa(counter)))
        counter++;
    }

    go func() {
        wg.Wait()
        close(tube)
    }()    

    var out []Product = []Product{}

    for product := range tube {
        out = append(out, product)
    }

    json, _ := json.MarshalIndent(out, "", "  ")

    fmt.Println(string(json))
}

func parse(page string) {

    defer wg.Done()

    response, _ := http.PostForm(page, url.Values{"pager": {"all"}})
    defer response.Body.Close()

    bodyInUnicode, _ := iconv.NewReader(response.Body, "windows-1251", "utf-8")

    document, _ := goquery.NewDocumentFromReader(bodyInUnicode)

    document.Find("div.new-style-row > div.js-shopitem").Each(func(i int, item *goquery.Selection) {

        id := item.AttrOr("data-id", "")

        divs := item.Find("div > span")

        desc, _ := item.Find(".description-item").First().Html()

        tube <- Product{
            ID: id,
            Article: divs.Eq(0).Text(),
            Brand: divs.Eq(1).Text(),
            Image: "https://www.sds-group.ru" + item.Find("img.image").First().AttrOr("src", ""),
            Link: fmt.Sprintf("https://www.sds-group.ru/items_%s.htm", id),
            Title: item.Find("div.product-name > a").First().Text(),
            Price: divs.Eq(2).Text(),
            Units: divs.Eq(3).Text(),
            InPack: divs.Eq(4).Text(),
            Description: base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(desc))),
            Availability: item.Find("table .gray-kol--active-color").Length(),
        }
    })

}