#!/bin/bash

go build zhihu.go

HOSTS=`cat hosts`

for i in ${HOSTS[@]}; do
  echo "transfer to $i"
  scp -P 10022 zhihu git@$i:/home/git
done
