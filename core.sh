#!/bin/bash

HOSTS=(10.64.123.72 10.64.138.233 10.175.43.94 10.175.43.100 10.175.43.104 10.175.43.105)

for i in ${HOSTS[@]}; do
  core=`ssh git@$i -p 10022 nproc`
  echo "$i $core"
done
