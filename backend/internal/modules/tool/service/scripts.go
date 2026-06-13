package service

type builtinScript struct {
	Name        string
	DisplayName string
	Category    string
	Description string
	CheckCmd    string
	VersionCmd  string
	InstallCmd  string
}

var builtinScripts = []builtinScript{
	{
		Name: "nginx", DisplayName: "Nginx", Category: "web",
		Description: "高性能 HTTP 和反向代理服务器",
		CheckCmd:    "which nginx || systemctl is-active nginx 2>/dev/null || nginx -v 2>&1 | head -1",
		VersionCmd:  "nginx -v 2>&1 | cut -d'/' -f2",
		InstallCmd: `#!/bin/bash
set -e
if command -v apt-get &>/dev/null; then
  apt-get update -qq && apt-get install -y -qq nginx
elif command -v yum &>/dev/null; then
  yum install -y nginx
else
  echo "unsupported package manager" && exit 1
fi
systemctl enable nginx && systemctl start nginx
echo "Nginx installed successfully"`,
	},
	{
		Name: "mysql", DisplayName: "MySQL", Category: "database",
		Description: "MySQL 关系型数据库（社区版）",
		CheckCmd:    "which mysqld || systemctl is-active mysqld 2>/dev/null || mysql --version 2>&1 | head -1",
		VersionCmd:  "mysql --version 2>&1 | awk '{print $3}'",
		InstallCmd: `#!/bin/bash
set -e
if command -v apt-get &>/dev/null; then
  DEBIAN_FRONTEND=noninteractive apt-get update -qq && apt-get install -y -qq mysql-server
elif command -v yum &>/dev/null; then
  yum install -y mysql-server
else
  echo "unsupported package manager" && exit 1
fi
systemctl enable mysqld && systemctl start mysqld
echo "MySQL installed successfully"`,
	},
	{
		Name: "redis", DisplayName: "Redis", Category: "database",
		Description: "高性能内存 Key-Value 缓存数据库",
		CheckCmd:    "which redis-server || systemctl is-active redis 2>/dev/null || redis-server --version 2>&1 | head -1",
		VersionCmd:  "redis-server --version 2>&1 | awk '{print $3}' | cut -d'=' -f2",
		InstallCmd: `#!/bin/bash
set -e
if command -v apt-get &>/dev/null; then
  apt-get update -qq && apt-get install -y -qq redis-server
elif command -v yum &>/dev/null; then
  yum install -y redis
else
  echo "unsupported package manager" && exit 1
fi
systemctl enable redis && systemctl start redis
echo "Redis installed successfully"`,
	},
	{
		Name: "node_exporter", DisplayName: "Node Exporter", Category: "monitoring",
		Description: "Prometheus Node Exporter 主机指标采集器",
		CheckCmd:    "which node_exporter || systemctl is-active node_exporter 2>/dev/null || node_exporter --version 2>&1 | head -1",
		VersionCmd:  "node_exporter --version 2>&1 | head -1 | awk '{print $3}'",
		InstallCmd: `#!/bin/bash
set -e
VERSION=${1:-1.7.0}
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
URL="https://github.com/prometheus/node_exporter/releases/download/v${VERSION}/node_exporter-${VERSION}.linux-${ARCH}.tar.gz"
cd /tmp
curl -sL "$URL" | tar xz
cp node_exporter-${VERSION}.linux-${ARCH}/node_exporter /usr/local/bin/
rm -rf node_exporter-${VERSION}.linux-${ARCH}
cat > /etc/systemd/system/node_exporter.service << 'UNIT'
[Unit]
Description=Node Exporter
After=network.target
[Service]
ExecStart=/usr/local/bin/node_exporter
Restart=always
[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload && systemctl enable node_exporter && systemctl start node_exporter
echo "Node Exporter installed successfully"`,
	},
	{
		Name: "consul_agent", DisplayName: "Consul Agent", Category: "service-mesh",
		Description: "HashiCorp Consul 服务发现与配置管理代理",
		CheckCmd:    "which consul || systemctl is-active consul 2>/dev/null || consul version 2>&1 | head -1",
		VersionCmd:  "consul version 2>&1 | head -1 | awk '{print $2}'",
		InstallCmd: `#!/bin/bash
set -e
VERSION=${1:-1.18.0}
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
URL="https://releases.hashicorp.com/consul/${VERSION}/consul_${VERSION}_linux_${ARCH}.zip"
cd /tmp
curl -sL "$URL" -o consul.zip
unzip -o consul.zip
cp consul /usr/local/bin/
rm -f consul consul.zip
mkdir -p /etc/consul.d /var/lib/consul
cat > /etc/systemd/system/consul.service << 'UNIT'
[Unit]
Description=Consul Agent
After=network.target
[Service]
ExecStart=/usr/local/bin/consul agent -config-dir=/etc/consul.d -data-dir=/var/lib/consul
Restart=always
[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload && systemctl enable consul && systemctl start consul
echo "Consul Agent installed successfully"`,
	},
}
