GOOS=linux GOARCH=amd64 go build .
tar -zcf auctioneer-lite.tar.gz auctioneer-lite
source ~/.bashisms/s3_upload.bash
upload_to_s3 auctioneer-lite.tar.gz
rm auctioneer-lite auctioneer-lite.tar.gz