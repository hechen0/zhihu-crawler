#!/bin/bash

HOSTS=(10.64.123.72 10.64.138.233 10.64.119.140 10.175.43.94 10.175.43.100)

for i in ${HOSTS[@]}; do
  result=`ssh git@$i -p 10022 $@`
  echo "$i $result"
done
