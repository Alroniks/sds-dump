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
    "net/url"
    "github.com/djimenez/iconv-go"
    "github.com/PuerkitoBio/goquery"
    "strconv"
    "sort"
)

const OUTPUT = "resources/output.json"

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
    Video string `json:"video"`
    Link string `json:"link"`

}

var categories = []Category {

    // Category{1459, 2, "Готовые наборы для украшения"},
    Category{1667, 1459, "Интерьерные наборы"},
    Category{1668, 1459, "Уличные наборы"},

	Category{1622, 2, "Промо-гирлянды"},

	// Category{317, 2, "Дюралайт"},
	Category{943, 317, "Дюралайт с постоянным свечением"},
	Category{944, 317, "Дюралайт с динамическим свечением"},
	Category{945, 317, "Дюралайт с эффектом мерцания"},

	// Category{315, 2, "Гибкий неон"},
	Category{946, 315, "Гибкий неон с белой оболочкой"},
	Category{947, 315, "Гибкий неон с цветной оболочкой"},
	Category{948, 315, "Гибкий неон круглый 360 градусов"},
	Category{1416, 315, "Гибкий неон компактный, двусторонний"},
	Category{1417, 315, "Гибкий неон в форме D"},

	// Category{297, 2, "Гирлянда-сетка"},
	Category{658, 297, "Гирлянда-сетка Home"},
	Category{424, 297, "Гирлянда-сетка Original"},
	Category{425, 297, "Гирлянда-сетка Professional"},

	// Category{298, 2, "Гирлянда-дождь"},
	Category{1023, 298, "Гирлянда-дождь Home"},
	Category{1024, 298, "Гирлянда-дождь Original"},
	Category{1025, 298, "Гирлянда-дождь Professional"},
	Category{952, 298, "Гирлянда умный дождь"},

	// Category{299, 2, "Гирлянда-бахрома"},
	Category{1026, 299, "Гирлянда-бахрома Home"},
	Category{1027, 299, "Гирлянда-бахрома Original"},
	Category{1028, 299, "Гирлянда-бахрома Professional"},

	// Category{302, 2, "Гирлянда-нить"},
	Category{1435, 302, "Нить Original"},
	Category{364, 302, "Нить Professional (Дюраплей)"},
	Category{965, 302, "Мишура"},
	Category{968, 302, "Роса"},
	Category{1636, 302, "Гирлянды с насадками"},
	Category{966, 302, "Мультишарики Home"},
	Category{967, 302, "Мультишарики Original"},
	Category{301, 302, "Твинкл-лайт Home"},
	Category{971, 302, "Твинкл-лайт Original"},
	Category{972, 302, "Твинкл-лайт Professional"},

	// Category{312, 2, "Искусственные елки"},
	Category{992, 312, "Еловые шлейфы"},
	Category{1657, 312, "Рождественские венки"},
	Category{993, 312, "Комнатные елки"},
	Category{995, 312, "Уличные елки"},
	Category{996, 312, "Елки фиброоптика"},

	// Category{465, 2, "Елочные игрушки"},
	Category{990, 465, "Фигуры елочные"},
	Category{991, 465, "Шары елочные"},

	// Category{303, 2, "Клип-лайт"},
	Category{1034, 303, "Клип-лайт Original"},
	Category{1035, 303, "Клип-лайт Professional"},

	// Category{300, 2, "Тающие сосульки"},
	Category{1169, 300, "Тающие сосульки Original"},
	Category{960, 300, "Тающие сосульки Professional"},
	Category{961, 300, "Тающие сосульки, готовые комплекты"},

	// Category{304, 2, "Белт-лайт"},
	Category{979, 304, "Двухжильный белт-лайт"},
	Category{980, 304, "Пятижильный белт-лайт"},
	Category{981, 304, "Белт-лайт, готовые комплекты 10 м"},

	// Category{1562, 2, "Интерьерные фигуры"},
	Category{1002, 1562, "Фигуры на присоске"},
	Category{1167, 1562, "Фигуры настольные"},
	Category{1925, 1562, "Декоративные фонарики"},
	Category{1926, 1562, "Деревянные фигурки"},
	Category{1927, 1562, "Керамические фигурки"},
	Category{1432, 1562, "Фигуры напольные"},
	Category{1003, 1562, "Фигуры подвесные"},
	Category{1003, 1562, "Диско-лампы и проекторы"}, // вложенность +1
	Category{1928, 1562, "Силиконовые светильники"},
	Category{1934, 1562, "Светодиодные камины"},

	// Category{307, 2, "Фигуры из дюралайта"},
	Category{296, 307, "Снежинки и звезды 2D"},
	Category{294, 307, "Мотивы малые и средние 2D"},
	Category{292, 307, "Световые панно 2D"},
	Category{293, 307, "Пушистые панно 2D"},
	Category{291, 307, "Мотивы крупные 2D"},
	Category{1001, 307, "Фигуры с заполнением 2D"},

	// Category{295, 2, "Объемные световые фигуры"},
	Category{1013, 295, "Каркасные фигуры 3D"},
	Category{340, 295, "Пушистые фигуры 3D"},
    Category{1015, 295, "Шары каркасные 3D"},
    Category{1016, 295, "Шары пушистые 3D"},
    Category{1017, 295, "Шары с лепестками \"Сакуры\" 3D"},

    Category{308, 2, "Надувные фигуры 3D"},

    // Category{305, 2, "Декоративные лампы"},
    Category{1670, 305, "Лампы"}, // вложенность +1
    Category{306, 305, "Стробы"}, // вложенность +1

    // Category{309, 2, "Акриловые фигуры"},
    Category{1006, 309, "Акриловые звезды 3D"},
    Category{1008, 309, "Акриловые фигуры маленькие 3D"},
    Category{1009, 309, "Акриловые фигуры средние 3D"},
    Category{1007, 309, "Акриловые фигуры крупные 3D"},
    Category{1004, 309, "Акриловые мотивы 2D"},

    // Category{310, 2, "Cветодиодные деревья"},
    Category{373, 310, "Деревья Клён"},
    Category{371, 310, "Деревья Сакура"},
    Category{1010, 310, "Деревья Сакура для помещения"},
    Category{372, 310, "Деревья Яблоня"},
    Category{1012, 310, "Аксессуары для деревьев"},

    // Category{323, 2, "Аксессуары для гирлянд"},
    Category{1660, 323, "Аксессуары для гирлянды-дождь"},
    Category{1661, 323, "Аксессуары для гирлянды-умный дождь"},
    Category{1662, 323, "Аксессуары для гирлянды-сеть"},
    Category{1663, 323, "Аксессуары для гирлянды-нить"},
    Category{1664, 323, "Аксессуары для светодиодных сосулек"},
    Category{1665, 323, "Аксессуары для гирлянды-дюраплей"},
    Category{1666, 323, "Аксессуары для гирлянды-белт-лайт"},
    Category{949, 323, "Аксессуары для гибкого неона"},
    Category{278, 323, "Аксессуары для дюралайта"},

    // Category{1632, 2, "Крупногабаритные консоли и конструкции"},
    Category{1943, 1632, "Арки"},
    Category{1944, 1632, "Декорации"},
    Category{1945, 1632, "Елки"},
    Category{1946, 1632, "Консоли"},
    Category{1947, 1632, "Объёмные фигуры"},
    Category{1948, 1632, "Панно"},
    Category{1949, 1632, "2D фигуры"},
    Category{1950, 1632, "3D фигуры"},
    Category{1951, 1632, "Фонтаны"},
    Category{1952, 1632, "Фотозоны"},
    Category{1958, 1632, "Перетяжки"},

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
        product.Link = fmt.Sprintf("https://neonsvet.by/shop/%d", product.ID)

        out = append(out, product)
    }

    sort.Slice(out, func (i, j int) bool {
        return out[i].ID < out[j].ID
    })

    jsonOutput, _ := json.MarshalIndent(out, "", "  ")

    ioutil.WriteFile(OUTPUT, jsonOutput, 0644)
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
            Image: "https://www.sds-group.ru" + item.Find("img.image").First().AttrOr("data-src", ""),
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