#!/usr/bin/env bash

p=`pwd`
for d in $(ls ./cmd); do
  echo "building cmd/$d"
  cd $p/cmd/$d
  go build
done
cd $p
