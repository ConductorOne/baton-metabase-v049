package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	MetabaseBaseUrl = field.StringField(
		"metabase-base-url",
		field.WithRequired(true),
		field.WithDescription("Metabase Base URL e.g. https://metabase.example.com"),
		field.WithDisplayName("Base URL"),
	)

	MetabaseApiKey = field.StringField(
		"metabase-api-key",
		field.WithRequired(true),
		field.WithIsSecret(true),
		field.WithDescription("Metabase API Key"),
		field.WithDisplayName("API Key"),
	)

	MetabaseWithPaidPlan = field.BoolField(
		"metabase-with-paid-plan",
		field.WithDescription("Set to true if using Metabase paid plan, false for Open Source / self-hosted (free)"),
		field.WithDisplayName("Metabase with paid plan"),
		field.WithDefaultValue(false),
	)

	// ConfigurationFields defines the external configuration required for the connector to run.
	ConfigurationFields = []field.SchemaField{
		MetabaseBaseUrl,
		MetabaseApiKey,
		MetabaseWithPaidPlan,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(ConfigurationFields,
	field.WithConnectorDisplayName("Metabase-v049"),
	field.WithHelpUrl("/docs/baton/metabase"),
	field.WithIconUrl("/static/app-icons/metabase.svg"),
)
