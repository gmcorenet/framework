module github.com/gmcorenet/framework

go 1.23

require (
	github.com/gmcorenet/sdk-gmcore-events v0.1.0
	golang.org/x/crypto v0.28.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gmcorenet/sdk-gmcore-config v0.1.0 // indirect
	github.com/gmcorenet/sdk-gmcore-error v0.1.0 // indirect
)

replace (
	github.com/gmcorenet/sdk-gmcore-config => ../sdks/gmcore-config
	github.com/gmcorenet/sdk-gmcore-error => ../sdks/gmcore-error
	github.com/gmcorenet/sdk-gmcore-events => ../sdks/gmcore-events
)
