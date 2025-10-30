package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var (
	databaseResourceType = &v2.ResourceType{
		Id:          "database",
		DisplayName: "Database",
	}
)
