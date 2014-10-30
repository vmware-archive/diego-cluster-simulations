GOOS=linux GOARCH=amd64 go build .
GOOS=linux GOARCH=amd64 ginkgo build
tar -zcf auctionscenarios.tar.gz auctionscenarios auctionscenarios.test
source ~/.bashisms/s3_upload.bash
upload_to_s3 auctionscenarios.tar.gz
rm auctionscenarios.tar.gz auctionscenarios auctionscenarios.test