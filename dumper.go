package main

import (
    "os"
    "io"
    "io/ioutil"
    "fmt"
    "sync"
    "strings"
    "net/http"
    "encoding/base64"
    "encoding/json"
    "encoding/csv"
    "net/url"
    "github.com/djimenez/iconv-go"
    "github.com/PuerkitoBio/goquery"
    "strconv"
)

var wg sync.WaitGroup

var tube chan Product

type Category struct {
    ID int
    Parent int
    Name string
}

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

var categories = []Category {
    Category{295, 2, "Объемные световые фигуры"},
    Category{340, 295, "Пушистые фигуры 3D"},
    Category{1013, 295, "Каркасные фигуры 3D"},
    Category{1015, 295, "Шары каркасные 3D"},
    Category{1016, 295, "Шары пушистые 3D"},
    Category{1017, 295, "Шары с лепестками \"Сакуры\" 3D"},
}

const CATEGORY_LINK_TEMPLATE = "https://www.sds-group.ru/catalog_table_%s.htm"

func main() {

    tube = make(chan Product)

    for _, category := range categories {
        wg.Add(1)
        go parse(category.ID)
    }

    go func() {
        wg.Wait()
        close(tube)
    }()    

    file, _ := os.Create("import.csv")
    defer file.Close()
    writer := csv.NewWriter(file)


    var out []Product = []Product{}

    for product := range tube {

        dest := fmt.Sprintf("data/img/%s.jpg", strconv.Itoa(product.ID))

        err := download(dest, product.Image)
        
        if err != nil {
            panic(err)
        }

        description, _ := base64.StdEncoding.DecodeString(product.Description)
        
        ioutil.WriteFile(fmt.Sprintf("data/dsc/%s.txt", strconv.Itoa(product.ID)), description, 0644)

        product.Description = ""

        csverr := writer.Write([]string {
            strconv.Itoa(product.ID),
            product.Article,
            product.Brand,
            product.Price,
            product.Title,
            product.Units,
            product.InPack,
            strconv.Itoa(product.Availability),
        });
        
        if csverr != nil {
            fmt.Println("Error of writing record to csv: ", csverr)
        }

        out = append(out, product)
    }

    json, _ := json.MarshalIndent(out, "", "  ")

    ioutil.WriteFile("output.json", json, 0644)

    writer.Flush()
}

func parse(category int) {

    defer wg.Done()

    page := fmt.Sprintf(CATEGORY_LINK_TEMPLATE, strconv.Itoa(category))

    response, _ := http.PostForm(page, url.Values{"pager": {"all"}})
    defer response.Body.Close()

    bodyInUnicode, _ := iconv.NewReader(response.Body, "windows-1251", "utf-8")

    document, _ := goquery.NewDocumentFromReader(bodyInUnicode)

    document.Find("div.new-style-row > div.js-shopitem").Each(func(i int, item *goquery.Selection) {

        id, _ := strconv.Atoi(item.AttrOr("data-id", ""))

        divs := item.Find("div > span")

        desc, _ := item.Find(".description-item").First().Html()

        tube <- Product{
            ID: id,
            Category: category,
            Article: divs.Eq(0).Text(),
            Brand: divs.Eq(1).Text(),
            Image: "https://www.sds-group.ru" + item.Find("img.image").First().AttrOr("src", ""),
            Title: item.Find("div.product-name > a").First().Text(),
            Price: divs.Eq(2).Text(),
            Units: divs.Eq(3).Text(),
            InPack: divs.Eq(4).Text(),
            Description: base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(desc))),
            Availability: item.Find("table .gray-kol--active-color").Length(),
        }
    })

}

func download(filepath string, url string) error {

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return err
    }

    return nil
}