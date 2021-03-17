export PATH=$PATH:$(go env GOPATH)/bin 
export GOPATH=$(go env GOPATH)
go build         
appify -name Gezam -icon icons/1024.png gezam
chmod +x Gezam.app/Contents/MacOS/Gezam.app
rm gezam 