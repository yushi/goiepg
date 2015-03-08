package goiepg

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// IEPG represents IEPG data
type IEPG struct {
	Header map[string][]string
	Body   string
}

// GetUniqueName returns summarized name
func (i *IEPG) GetUniqueName() string {
	return fmt.Sprintf("%s-%s-%s",
		i.Start().Format("2006-01-02-1504"),
		i.End().Format("1504"),
		i.getHeaderByString("Program-Title"))
}

// Start returns start time
func (i *IEPG) Start() time.Time {
	return i.getTime("Start")
}

// Start returns end time
func (i *IEPG) End() time.Time {
	return i.getTime("End")
}

func (i *IEPG) getTime(key string) time.Time {
	datestr := fmt.Sprintf("%d/%d/%d %s",
		i.getHeaderByInt("Year"),
		i.getHeaderByInt("Month"),
		i.getHeaderByInt("Date"),
		i.getHeaderByString(key))
	form := "2006/1/2 15:04"
	t, _ := time.Parse(form, datestr)
	return t
}

func (i *IEPG) getHeaderByInt(key string) int {
	v, ok := i.Header[key]
	if !ok {
		return 0
	}
	if len(v) == 0 {
		return 0
	}
	n, err := strconv.Atoi(v[0])
	if err != nil {
		return 0
	}
	return n
}

func (i *IEPG) getHeaderByString(key string) string {
	v, ok := i.Header[key]
	if !ok {
		return ""
	}
	if len(v) == 0 {
		return ""
	}
	return i.Header[key][0]
}

// ParseIEPG parse IEPG from io.Reader
func ParseIEPG(r io.Reader) (*IEPG, error) {
	br := bufio.NewReader(r)
	tr := textproto.NewReader(br)
	headers, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	for k, values := range headers {
		decoded := make([]string, len(values))
		for i := range values {
			d, err := decodeSJIS(values[i])
			if err != nil {
				d = values[i]
			}
			decoded[i] = d
		}
		headers[k] = decoded
	}
	b, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, err
	}
	body, err := decodeSJIS(string(b))
	if err != nil {
		return nil, err
	}
	return &IEPG{Header: headers, Body: body}, nil
}

func decodeSJIS(in string) (string, error) {
	r := transform.NewReader(strings.NewReader(in), japanese.ShiftJIS.NewDecoder())
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
