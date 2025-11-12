#!/bin/bash

echo "Opening firewall for RCC Web UI..."
echo "=================================="

# Check if ufw is active
if command -v ufw >/dev/null 2>&1; then
    echo "UFW firewall detected"
    echo "Run these commands to open port 3000:"
    echo ""
    echo "sudo ufw allow 3000"
    echo "sudo ufw reload"
    echo ""
    echo "Or to check current status:"
    echo "sudo ufw status"
    echo ""
fi

# Check if iptables is being used
if command -v iptables >/dev/null 2>&1; then
    echo "iptables detected"
    echo "Run these commands to open port 3000:"
    echo ""
    echo "sudo iptables -A INPUT -p tcp --dport 3000 -j ACCEPT"
    echo "sudo iptables -A OUTPUT -p tcp --sport 3000 -j ACCEPT"
    echo ""
fi

# Check if firewalld is being used
if command -v firewall-cmd >/dev/null 2>&1; then
    echo "firewalld detected"
    echo "Run these commands to open port 3000:"
    echo ""
    echo "sudo firewall-cmd --permanent --add-port=3000/tcp"
    echo "sudo firewall-cmd --reload"
    echo ""
fi

echo "Server is running on:"
echo "- http://192.168.1.120:3000"
echo "- http://192.168.1.141:3000" 
echo "- http://10.200.200.41:3000"
echo ""
echo "After opening the firewall, test with:"
echo "curl http://192.168.1.120:3000/"
