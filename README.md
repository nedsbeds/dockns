# dockns
A small proof of concept go app that monitors containers for traefik host labels and manages your hosts file 

## Features

Responds to docker start and stop events

Parses Traefik `Host:` labels

Adds or removes host entries from /etc/hosts

## Running 
This is a POC. Backup your hosts file first! 

Download the binary and then run `sudo ./dockns`
