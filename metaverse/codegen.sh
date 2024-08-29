GoDir=pbs/ #服务端PBS目录
#ClientDir=~/go/src/metaverse-client/pbs/ #客户端测试pbs目录
ProtoDir=protos/ #服务端proto目录
Protobuf=~/go/src/github.com/gogo/protobuf/gogoproto

protoc -I=$ProtoDir -I=$GOPATH/src -I=$Protobuf --gofast_out=plugins=grpc:$GoDir  --plugin=protoc-gen-grpc=grpc_csharp_plugin --grpc_opt=lite_client  $ProtoDir/*.proto
