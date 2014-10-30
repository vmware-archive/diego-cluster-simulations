GOOS=linux GOARCH=amd64 go build .
tar -zcf rep-lite.tar.gz rep-lite
source ~/.bashisms/s3_upload.bash
upload_to_s3 rep-lite.tar.gz
rm rep-lite rep-lite.tar.gz