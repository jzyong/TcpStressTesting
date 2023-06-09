:gate message
tool\protoc --go_out="..\..\client\message" --go_opt=paths=source_relative --go-grpc_out="..\..\client\message" --go-grpc_opt=paths=source_relative *.proto
pause
