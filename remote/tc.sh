interface=$(ls /sys/class/net | grep -E '^e(n|th)')
sudo tc qdisc add dev $interface root tbf rate 1024Mbit burst 128mb latency 50ms

#sudo tc qdisc del dev $interface root
#sudo tc qdisc add dev $interface root handle 1: htb default 10
#sudo tc class add dev $interface parent 1: classid 1:1 htb rate 1Gbit burst 128m
#sudo tc filter add dev $interface protocol ip prio 1 u32 match ip dport 12000 0xffff flowid 1:1
#sudo tc class add dev $interface parent 1:1 classid 1:10 htb rate 1Gbit ceil 1Gbit burst 128m
#sudo tc qdisc add dev $interface parent 1:10 handle 10: netem delay 50ms

