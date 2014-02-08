GO_VERSION=1.2

apt-get update
apt-get install -y make curl

/usr/bin/curl -o /usr/local/src/go${GO_VERSION}.linux-amd64.tar.gz https://go.googlecode.com/files/go${GO_VERSION}.linux-amd64.tar.gz
/bin/tar -C /usr/local -xzf /usr/local/src/go${GO_VERSION}.linux-amd64.tar.gz
/bin/echo 'export PATH=/vagrant/bin:/usr/local/go/bin:$PATH' >> /home/vagrant/.profile
/bin/echo 'export GOPATH=/vagrant' >> /home/vagrant/.profile
