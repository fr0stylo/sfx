module github.com/fr0stylo/sfx/plugins/providers/file

go 1.25.1

require gopkg.in/yaml.v3 v3.0.1

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

require (
	github.com/fr0stylo/sfx v0.0.0-20251018111239-fd3d6a6525d1
	github.com/stretchr/testify v1.11.1
)

replace github.com/fr0stylo/sfx v0.0.0-20251018111239-fd3d6a6525d1 => ../../..
