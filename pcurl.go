package pcurl

import (
	"github.com/guonaihong/clop"
	"github.com/guonaihong/gout"
	"net/http"
	"os"
	"strings"
)

type Curl struct {
	Method string   `clop:"-X; --request" usage:"Specify request command to use"`
	Header []string `clop:"-H; --header" usage:"Pass custom header(s) to server"`
	Data   string   `clop:"-d; --data"   usage:"HTTP POST data"`
	Form   []string `clop:"-F; --form" usage:"Specify multipart MIME data"`
	URL2   string   `clop:"args=url2" usage:"url2"`
	URL    string   `clop:"--url" usage:"URL to work with"`

	Err error
	p   *clop.Clop
}

func ParseAndRequest(curl string) (*http.Request, error) {
	return Parse(curl).Request()
}

func Parse(curl string) *Curl {
	curlSlice := GetArgsToken(curl)
	return ParseSlice(curlSlice)
}

func ParseSlice(curl []string) *Curl {
	c := Curl{}
	if len(curl) > 0 && strings.ToLower(curl[0]) == "curl" {
		curl = curl[1:]
	}

	c.p = clop.New(curl).SetExit(false)
	c.Err = c.p.Bind(&c)
	return &c
}

func (c *Curl) createHeader() []string {
	if len(c.Header) == 0 {
		return nil
	}

	header := make([]string, len(c.Header)*2)
	index := 0
	for _, v := range c.Header {
		pos := strings.IndexByte(v, ':')
		if pos == -1 {
			continue
		}

		header[index] = v[:pos]
		index++
		header[index] = v[pos+1:]
		index++
	}

	return header
}

func (c *Curl) createForm() ([]interface{}, error) {
	if len(c.Form) == 0 {
		return nil, nil
	}

	form := make([]interface{}, len(c.Form)*2)
	index := 0
	for _, v := range c.Form {
		pos := strings.IndexByte(v, '=')
		if pos == -1 {
			continue
		}

		form[index] = v[:pos]
		index++
		fieldValue := v[pos+1:]
		if len(fieldValue) > 0 && fieldValue[0] == '@' {

			form[index] = gout.FormFile(fieldValue[1:])
		} else {

			form[index] = fieldValue
		}

		index++
	}

	return form, nil
}

func (c *Curl) Request() (*http.Request, error) {
	if len(c.Method) == 0 && len(c.Data) > 0 {
		c.Method = "POST"
	}

	var (
		data interface{}
	)

	header := c.createHeader()

	form, err := c.createForm()
	if err != nil {
		return nil, err
	}

	data = c.Data
	if len(c.Data) > 0 && c.Data[0] == '@' {
		fd, err := os.Open(c.Data[1:])
		if err != nil {
			return nil, err
		}

		defer fd.Close()

		data = fd
	}

	g := gout.New().
		SetMethod(c.Method). //设置method POST or GET or DELETE
		Debug(true)          //打开debug模式

	if header != nil {
		g.SetHeader(header) //设置http header
	}

	if len(form) > 0 {
		g.SetForm(form) //设置formdata
	}

	url := c.URL2
	if c.p.GetIndex("url") > c.p.GetIndex("url2") {
		url = c.URL
	}

	return g.SetURL(url). //设置url
				SetBody(data). //设置http body
				Request()      //获取*http.Request
}
