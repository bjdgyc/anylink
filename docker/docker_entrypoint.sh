#! /bin/bash
version=(`wget -qO- -t1 -T2 "https://api.github.com/repos/bjdgyc/anylink/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g'`)
count=(`ls anylink | wc -w `)
wget https://github.com/bjdgyc/anylink/releases/download/${version}/anylink-deploy.tar.gz
tar xf anylink-deploy.tar.gz
rm -rf anylink-deploy.tar.gz
if [ ${count} -eq 0 ]; then
	echo "init anylink"
	mv anylink-deploy/* anylink/
else
	if [ ! -d "/anylink/log" ]; then
		mv anylink-deploy/log anylink/
	fi
	if [ ! -d "/anylink/conf" ]; then
                mv anylink-deploy/conf anylink/
        fi
	echo "update anylink"
	rm -rf anylink/ui anylink/anylink anylink/files
	mv anylink-deploy/ui anylink/
	mv anylink-deploy/anylink anylink/
	mv anylink-deploy/files anylink/
fi
rm -rf anylink-deploy
sysctl -w net.ipv4.ip_forward=1
if [[ ${mode} == pro ]];then
	iptables -t nat -A POSTROUTING -s ${iproute} -o eth0 -j MASQUERADE
	iptables -L -n -t nat
	/anylink/anylink -conf=/anylink/conf/server.toml
elif [[ ${mode} == password ]];then
	if [ -z ${password} ];then
		echo "invalid password"
	else
		/anylink/anylink -passwd ${password}
	fi
elif [[ ${mode} -eq jwt ]];then
	/anylink/anylink -secret
fi
