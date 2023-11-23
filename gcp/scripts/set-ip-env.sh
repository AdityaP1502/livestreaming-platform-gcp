#!/bin/bash

# place this script in /etc/profile.d
export VM_IP="$(ip addr show scope global | awk '$1=="inet" {split($2,I,"/");print I[1]}')"
export TRANSCODER_LB_IP=127.0.0.1
