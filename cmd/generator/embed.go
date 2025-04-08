package main

import _ "embed"

//go:embed templates/tags.gotmpl
var tagsTemplate string

//go:embed templates/headers.tmpl
var headerTemplate string

//go:embed templates/func_read_api.gotmpl
var funcReadApiStart string

//go:embed templates/variable.gotmpl
var variableTemplate string

//go:embed templates/custom_fields.gotmpl
var customFieldsTemplate string

//go:embed templates/nullable_string.gotmpl
var nullableStringTemplate string
