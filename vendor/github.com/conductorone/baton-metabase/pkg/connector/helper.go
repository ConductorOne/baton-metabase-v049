package connector

import (
	"fmt"
	"strconv"

	"github.com/conductorone/baton-metabase/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func getPageOptions(pToken *pagination.Token, pageSize int) (client.PageOptions, error) {
	var offset int
	if pToken != nil && pToken.Token != "" {
		o, err := strconv.Atoi(pToken.Token)
		if err != nil {
			return client.PageOptions{}, fmt.Errorf("invalid page token: %w", err)
		}
		offset = o
	}

	var limit int
	if pToken != nil && pToken.Size > 0 {
		limit = pToken.Size
	} else {
		limit = pageSize
	}

	return client.PageOptions{
		Limit:  limit,
		Offset: offset,
	}, nil
}
