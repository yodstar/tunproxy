# tunproxy
A simple TCP/UDP proxy 

## Requirement
- Go 1.4 +

```
rpm --import https://mirror.go-repo.io/centos/RPM-GPG-KEY-GO-REPO
curl -s https://mirror.go-repo.io/centos/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
yum update golang
```

## Tutorial

### installation

- Clone repo and build (CentOS7):

```
git clone https://github.com/yodstar/tunproxy.git
cd tunproxy
make && make install
```

- Config file for server

```
{
	"Listen": ":18081",
	"Forward": {
		":443": "tcp://192.168.0.200:443"
	},
	"Logfile":"./logs/tunproxy_%s.log",
	"Filter":"WARN",
	"Level":"DEBUG"
}
```

- Config file for client

```
{
	"Server": "192.168.0.102:18081",
	"Logfile":"./logs/tunproxy_%s.log",
	"Filter":"WARN",
	"Level":"DEBUG"
}
```
