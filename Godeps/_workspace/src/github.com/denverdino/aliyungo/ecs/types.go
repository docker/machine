package ecs

import (
	"fmt"
	"github.com/denverdino/aliyungo/util"
	"log"
	"time"
)

// Constants for ECS requests
const (
	ECSAPIVersion      = "2014-05-26"
	SignatureVersion   = "1.0"
	SignatureMethod    = "HMAC-SHA1"
	JSONResponseFormat = "JSON"
	XMLResponseFormat  = "XML"
	ECSRequestMethod   = "GET"
)

type Request struct {
	Format               string
	Version              string
	AccessKeyId          string
	Signature            string
	SignatureMethod      string
	Timestamp            util.ISO6801Time
	SignatureVersion     string
	SignatureNonce       string
	ResourceOwnerAccount string
	Action               string
}

func (request *Request) init(action string, AccessKeyId string) {
	request.Format = JSONResponseFormat
	request.Timestamp = util.NewISO6801Time(time.Now().UTC())
	request.Version = ECSAPIVersion
	request.SignatureVersion = SignatureVersion
	request.SignatureMethod = SignatureMethod
	request.SignatureNonce = util.CreateRandomString()
	request.Action = action
	request.AccessKeyId = AccessKeyId
}

type CommonResponse struct {
	RequestId string
}

type ErrorResponse struct {
	CommonResponse
	HostId  string
	Code    string
	Message string
}

// An Error represents a custom error for ECS failure response
type Error struct {
	ErrorResponse
	StatusCode int //Status Code of HTTP Response
}

func (e *Error) Error() string {
	return fmt.Sprintf("ECS Error: Status Code %d: Code %s: Message %s", e.StatusCode, e.Code, e.Message)
}

type Pagination struct {
	PageNumber int
	PageSize   int
}

func (p *Pagination) SetPageSize(size int) {
	p.PageSize = size
}

// A PaginationResult represents a result with pagination information
type PaginationResult struct {
	TotalCount int
	PageNumber int
	PageSize   int
}

// NextPage gets the next page of the result set
func (r *PaginationResult) NextPage() *Pagination {
	if r.PageNumber*r.PageSize >= r.TotalCount {
		return nil
	}
	return &Pagination{PageNumber: r.PageNumber + 1, PageSize: r.PageSize}
}

func (p *Pagination) validate() {
	if p.PageNumber < 0 {
		log.Printf("Invalid PageNumber: %d", p.PageNumber)
		p.PageNumber = 1
	}
	if p.PageSize < 0 {
		log.Printf("Invalid PageSize: %d", p.PageSize)
		p.PageSize = 10
	} else if p.PageSize > 50 {
		log.Printf("Invalid PageSize: %d", p.PageSize)
		p.PageSize = 50
	}
}
