#! /bin/bash
DeployDir=/usr/local/tunproxy

echo Deploying Server2
ssh root@server2 "systemctl stop tunproxy"
scp ${DeployDir}/tunproxy root@server2:${DeployDir}
scp ${DeployDir}/tunproxy.conf root@server2:${DeployDir}
ssh root@server2 "systemctl start tunproxy"

echo Deploying Server3
ssh root@server3 "systemctl stop tunproxy"
scp ${DeployDir}/tunproxy root@server3:${DeployDir}
scp ${DeployDir}/tunproxy.conf root@server3:${DeployDir}
ssh root@server3 "systemctl start tunproxy"

echo Deploying Server4
ssh root@server4 "systemctl stop tunproxy"
scp ${DeployDir}/tunproxy root@server4:${DeployDir}
scp ${DeployDir}/tunproxy.conf root@server4:${DeployDir}
ssh root@server4 "systemctl start tunproxy"

echo Deploying Server5
ssh root@server5 "systemctl stop tunproxy"
scp ${DeployDir}/tunproxy root@server5:${DeployDir}
scp ${DeployDir}/tunproxy.conf root@server5:${DeployDir}
ssh root@server5 "systemctl start tunproxy"
