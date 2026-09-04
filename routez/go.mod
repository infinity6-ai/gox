module github.com/infinity6-ai/gox/routez

go 1.26.1

require (
	github.com/infinity6-ai/gox/commonz v0.0.0-20260904123318-490ac4a7cad8
	github.com/infinity6-ai/gox/httpz v0.0.0-20260904121639-0426e5496634
	github.com/infinity6-ai/gox/schemaz v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.12.1
)

require (
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
)

replace github.com/infinity6-ai/gox/schemaz => ../schemaz
