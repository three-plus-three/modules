// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package toolbox

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// DEFAULT_SIZE_PER_PAGE 缺省每页记录数
const DEFAULT_SIZE_PER_PAGE = 12

func toInt(value interface{}) (d int, err error) {
	val := reflect.ValueOf(value)
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = int(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = int(val.Uint())
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	return
}

// Paginator within the state of a http request.
type Paginator struct {
	Request     *http.Request
	PerPageNums int
	MaxPages    int

	nums      int
	pageRange []int
	pageNums  int
	pageIndex int
}

// NeedPaginated wh.
func (p *Paginator) NeedPaginated() bool {
	return p.PageNums() > 1
}

// PageNums Returns the total number of pages.
func (p *Paginator) PageNums() int {
	if p.pageNums != 0 {
		return p.pageNums
	}
	pageNums := math.Ceil(float64(p.nums) / float64(p.PerPageNums))
	if p.MaxPages > 0 {
		pageNums = math.Min(pageNums, float64(p.MaxPages))
	}
	p.pageNums = int(pageNums)
	return p.pageNums
}

// TotalNums Returns the total number of items (e.g. from doing SQL count).
func (p *Paginator) TotalNums() int {
	return p.nums
}

// SetTotalNums Sets the total number of items.
func (p *Paginator) SetTotalNums(nums interface{}) {
	p.nums, _ = toInt(nums)
}

// Page Returns the current page.
func (p *Paginator) Page() int {
	if p.pageIndex > 0 {
		return p.pageIndex
	}
	if p.Request.Form == nil {
		p.Request.ParseForm()
	}
	p.pageIndex, _ = strconv.Atoi(p.Request.Form.Get("pageIndex"))
	if p.pageIndex >= p.PageNums() {
		if 0 == p.PageNums() {
			p.pageIndex = 0
		} else {
			p.pageIndex = p.PageNums() - 1
		}
	}
	return p.pageIndex
}

// Pages Returns a list of all pages.
//
// Usage (in a view template):
//
//  {{range $index, $page := .paginator.Pages}}
//    <li{{if $.paginator.IsActive .}} class="active"{{end}}>
//      <a href="{{$.paginator.PageLink $page}}">{{$page}}</a>
//    </li>
//  {{end}}
func (p *Paginator) Pages() []int {
	if p.pageRange == nil && p.nums > 0 {
		var pages []int
		pageNums := p.PageNums()
		page := p.Page()
		switch {
		case page >= pageNums-4 && pageNums > 9:
			start := pageNums - 9
			pages = make([]int, 9)
			for i := range pages {
				pages[i] = start + i
			}
		case page >= 5 && pageNums > 9:
			start := page - 5
			pages = make([]int, int(math.Min(9, float64(page+4+1))))
			for i := range pages {
				pages[i] = start + i
			}
		default:
			pages = make([]int, int(math.Min(9, float64(pageNums))))
			for i := range pages {
				pages[i] = i
			}
		}
		p.pageRange = pages
	}
	return p.pageRange
}

// PageLink Returns URL for a given page index.
func (p *Paginator) PageLink(page int) string {
	link, _ := url.ParseRequestURI(p.Request.URL.String())
	values := link.Query()
	if page == 0 {
		values.Del("pageIndex")
	} else {
		values.Set("pageIndex", strconv.Itoa(page))
	}
	link.RawQuery = values.Encode()
	return link.String()
}

// PageLinkPrev Returns URL to the previous page.
func (p *Paginator) PageLinkPrev() (link string) {
	if p.HasPrev() {
		link = p.PageLink(p.Page() - 1)
	}
	return
}

// PageLinkNext Returns URL to the next page.
func (p *Paginator) PageLinkNext() (link string) {
	if p.HasNext() {
		link = p.PageLink(p.Page() + 1)
	}
	return
}

// PageLinkFirst Returns URL to the first page.
func (p *Paginator) PageLinkFirst() (link string) {
	return p.PageLink(0)
}

// PageLinkLast Returns URL to the last page.
func (p *Paginator) PageLinkLast() (link string) {
	if p.PageNums() == 0 {
		return p.PageLink(0)
	}
	return p.PageLink(p.PageNums() - 1)
}

// HasPrev Returns true if the current page has a predecessor.
func (p *Paginator) HasPrev() bool {
	return p.Page() > 0
}

// HasNext Returns true if the current page has a successor.
func (p *Paginator) HasNext() bool {
	if p.PageNums() == 0 {
		return false
	}
	return p.Page() < p.PageNums()-1
}

// IsActive Returns true if the given page index points to the current page.
func (p *Paginator) IsActive(page int) bool {
	return p.Page() == page
}

// Offset Returns the current offset.
func (p *Paginator) Offset() int {
	return (p.Page() - 1) * p.PerPageNums
}

// HasPages Returns true if there is more than one page.
func (p *Paginator) HasPages() bool {
	return p.PageNums() > 0
}

// NewPaginator Instantiates a paginator struct for the current http request.
func NewPaginator(req *http.Request, per int, nums interface{}) *Paginator {
	p := Paginator{}
	p.Request = req
	if per <= 0 {
		per = DEFAULT_SIZE_PER_PAGE
	}
	p.PerPageNums = per
	p.SetTotalNums(nums)
	return &p
}
