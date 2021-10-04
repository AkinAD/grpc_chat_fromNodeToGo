.IFYOUDONTHAVEPROTOSTUFF:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u "google.golang.org/protobuf"
# https://github.com/golang/protobuf/issues/811#issuecomment-531870562

protoGen-server: 
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/chat.proto
protoGen-client:
#	mkdir -p ./grpc-chat-app/src/proto
# fucking windows
#	mkdir -p .\grpc-chat-app\src\proto
# to resolve the pesky error with proto-gen-web https://huayuzhang.medium.com/i-encountered-the-error-572fba76bd1d
# do it in WSL as well
	protoc -I=. ./proto/*.proto \
		--js_out=import_style=commonjs:./grpc-chat-app/src \
		--grpc-web_out=import_style=typescript,mode=grpcwebtext:./grpc-chat-app/src

protoGen: 
	make protoGen-server && make protoGen-client
setup:
	go mod tidy
	docker compose up -d

serverRun:
	cls 
	go run server/server.go

clientRun: 
	cd grpc-chat-app && yarn start

teardown: 
	docker-compose down

reset: 
	docker-compose down -v
init:
# go mod int <go folder directory ( in this case github.com/AkinAD/grpc_chat_fromNodeToGo)>

installProto:
	go get -u github.com/golang/protobuf/protoc-gen-go
	PB_REL="https://github.com/protocolbuffers/protobuf/releases" && curl -LO $PB_REL/download/v3.18.0/protoc-3.18.0-linux-x86_64.zip
	unzip protoc-3.18.0-linux-x86_64.zip -d $HOME/.local && export PATH="$PATH:$HOME/.local/bin"
	npm install protoc-gen-grpc-web
	export PATH=$PATH:node_modules/.bin

