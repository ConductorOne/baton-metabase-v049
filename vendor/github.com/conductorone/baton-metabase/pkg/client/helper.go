package client

import (
	"net/url"
	"strconv"
)

const (
	ItemsPerPage = 100
)

type PageOptions struct {
	Limit  int
	Offset int
}

type ReqOpt func(reqURL *url.URL)

func withLimitParam(limit int) ReqOpt {
	if limit <= 0 {
		limit = ItemsPerPage
	}
	return withQueryParam("limit", strconv.Itoa(limit))
}

func withOffsetParam(offset int) ReqOpt {
	if offset <= 0 {
		return func(reqURL *url.URL) {}
	}
	return withQueryParam("offset", strconv.Itoa(offset))
}

func withQueryParam(key string, value string) ReqOpt {
	return func(reqURL *url.URL) {
		q := reqURL.Query()
		q.Set(key, value)
		reqURL.RawQuery = q.Encode()
	}
}

func getNextPageToken(offset, limit, total int) string {
	if offset+limit < total {
		return strconv.Itoa(offset + limit)
	}
	return ""
}
