package main

import (
	"bytes"
	"html/template"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Block struct {
    Id int
}

type Blocks struct {
    Start int
    Next int
    More bool
    Blocks []Block
}

type Count struct {
    Count int
}

type Contact struct {
    Name string
    Email string
}

func newContact(name string, email string) Contact {
    return Contact{
        Name: name,
        Email: email,
    }
}

type Data struct {
    Count int
    Contacts []Contact
}


type BlockName struct {
    Name string
}

type Cloud struct {
    HMovementSpeed int
    VMovementSpeed int
}

type Main struct {
    Body template.HTML
    CSS template.HTML
    BrightnessFactor float64
}

type MainBody struct {
    Clouds []template.HTML
}

type CloudVars struct {
    CloudSpeed int
    CloudHeight int
    CloudDelay float64
}

type WeatherControl struct {
    WindSpeed int
    CloudCover int
    BrightnessFactor float64
}

func newData() []Contact {
    return []Contact { 
            newContact("John", "john@example.com"),
            newContact("Jane", "jane@example.com"),
            newContact("Joe", "joe@example.com"),
        }
}

func renderTemplate(e *echo.Echo, name string, data interface{}) template.HTML {
    var tmplBytes bytes.Buffer
    err := e.Renderer.(*Templates).templates.ExecuteTemplate(&tmplBytes, name, data)
    if err != nil {
        return template.HTML("")
    }
    return template.HTML(tmplBytes.String())
}

func renderCloud(windSpeed int, cloudCover int, e *echo.Echo) []template.HTML {
    cloudOptions := []string{"cloudOne", "cloudTwo", "cloudThree", "cloudFour", "cloudFive"}
        var clouds []template.HTML
        for i := 0; i < int(cloudCover*3/10); i++ {
            //  max 250 = 2, min 0 ~= 100, decay at higher speeds, random speed up of 0-30% so all clouds dont move together
            cloudSpeed := int(math.Max(250*math.Exp(-float64((windSpeed+80)/50)), 2) * max(rand.Float64(), 0.7))
            clouds = append(clouds, renderTemplate(e, cloudOptions[rand.Intn(len(cloudOptions))], CloudVars{
                CloudSpeed: cloudSpeed,
                CloudHeight: rand.Intn(20),
                // random start 
                CloudDelay: -rand.Float64() * float64(cloudSpeed), 
            }))
        }
    return clouds
}


func main() {
	e := echo.New()
    e.Renderer = NewTemplates()
    e.Use(middleware.Logger())

    e.Static("/css", "css")

    data := Data {
        Count: 0,
        Contacts: newData(),
    }

    e.GET("/", func(c echo.Context) error {
        weatherControl := WeatherControl{
            WindSpeed: 40, // km/h
            CloudCover: 20, // %
            BrightnessFactor: 1, // 1 == full sun and day, 0 == night/dark
        }


        BodyClouds := MainBody{
            Clouds: renderCloud(weatherControl.WindSpeed, weatherControl.CloudCover, e),
        }
        return c.Render(http.StatusOK, "index", Main{
            Body: renderTemplate(e, "rootBody", BodyClouds),
            CSS: renderTemplate(e, "rootCSS", nil),
            BrightnessFactor: weatherControl.BrightnessFactor,
        })
    })

    e.GET("/cc", func(c echo.Context) error {
        return c.Render(http.StatusOK, "index", c.Render(http.StatusOK, "body2", data))
    })

    e.POST("/count", func(c echo.Context) error {
        data.Count++
        return c.Render(http.StatusOK, "count", data)
    })

    e.POST("/contacts", func(c echo.Context) error {
        name := c.FormValue("name")
        email := c.FormValue("email")
        data.Contacts = append(data.Contacts, newContact(name, email))
        return c.Render(http.StatusOK, "display", data)
    })

    e.GET("/blocks", func(c echo.Context) error {
        startStr := c.QueryParam("start")
        start, err := strconv.Atoi(startStr)
        if err != nil {
            start = 0
        }

        blocks := []Block{}
        for i := start; i < start + 10; i++ {
            blocks = append(blocks, Block{Id: i})
        }

        template := "blocks"
        if start == 0 {
            template = "blocks-index"
        }
        return c.Render(http.StatusOK, template, Blocks{
            Start: start,
            Next: start + 10,
            More: start + 10 < 100,
            Blocks: blocks,
        });
    });

    // e.GET("/*", func(c echo.Context) error {
    //     return c.Render(http.StatusOK, "main", c.Render(http.StatusOK, "test2", nil))
    // })

    e.Logger.Fatal(e.Start(":42069"))
}
