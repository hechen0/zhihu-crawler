#!/bin/bash

HOSTS=`cat hosts`

for i in ${HOSTS[@]}; do
  result=`ssh git@$i -p 10022 $@`
  echo "------------------ $i result ----------------------"
  echo "$result"
  echo 
  echo
done
