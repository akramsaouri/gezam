export PATH=$PATH:$(go env GOPATH)/bin 
export GOPATH=$(go env GOPATH)
go build         
appify -name gezam -icon icons/1024.png gezam 
rm gezam 