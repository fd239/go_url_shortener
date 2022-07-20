protoc:
	protoc -IC:/protoc/include -I. \
		-IC:\Users\fd239\go\pkg\mod\github.com\grpc-ecosystem\grpc-gateway@v1.16.0\third_party\googleapis \
		-IC:\Users\fd239\go\pkg\mod\github.com\envoyproxy\protoc-gen-validate@v0.6.7 \
		--grpc-gateway_out=logtostderr=true:.\ \
		--swagger_out=allow_merge=true,merge_file_name=api:. \
		--go_out=plugins=grpc:.\api \
		--validate_out="lang=go:./api" api/*.proto