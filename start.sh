apt update

apt install wget 

wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin

go version