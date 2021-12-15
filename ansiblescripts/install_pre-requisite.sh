#!/bin/bash
sudo apt update -y
sudo apt upgrade -y

sudo apt install python3 python3-pip ansible sshpass -y
python3 -m pip install -U pip
python3 -m pip install -U setuptools