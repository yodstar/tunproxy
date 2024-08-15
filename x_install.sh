#! /bin/bash
cd `dirname $0`

echo Build
make rebuild

echo Install
systemctl stop tunproxy
make install

echo Complete
systemctl daemon-reload
systemctl start tunproxy
