#!/usr/bin/env bash
set -ex

oLinux=~/date_hole_bin
buildLinux(){
  GOOS=linux GOARCH=amd64 go build -o $oLinux

}

runLinux(){
  $oLinux -file $FILE
}

$@