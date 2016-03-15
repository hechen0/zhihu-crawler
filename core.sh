#!/bin/bash

HOSTS=`cat hosts`

for i in ${HOSTS[@]}; do
  core=`ssh git@$i -p 10022 nproc`
  echo "$i $core"
done
