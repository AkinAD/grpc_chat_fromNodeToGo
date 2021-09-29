protoGen-server: 
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/chat.proto
protoGen-client:
#	mkdir -p ./grpc-chat-app/src/proto
# fucking windows
#	mkdir -p .\grpc-chat-app\src\proto
# to resolve the pesky error with proto-gen-web https://huayuzhang.medium.com/i-encountered-the-error-572fba76bd1d
	protoc -I=. ./proto/*.proto \
		--js_out=import_style=commonjs:./grpc-chat-app/src \
		--grpc-web_out=import_style=typescript,mode=grpcwebtext:./grpc-chat-app/src
	

setup:
	go mod tidy
	docker compose up

serverRun: 
	go run server/server.go

clientRun: 
	cd grpc-chat-app && yarn start

init:
# go mod int <go folder directory ( in this case github.com/AkinAD/grpc_chat_fromNodeToGo)>