#!/bin/bash


make
killall daemon
rm -R run
sleep 1
mkdir run
cd run
../simulation/sim
