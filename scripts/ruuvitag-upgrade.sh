#!/usr/bin/env bash

pushd .
cd ..
sudo systemctl stop ruuvitag-gollector.service
sudo cp ruuvitag-gollector /usr/local/bin/
sudo systemctl start ruuvitag-gollector.service
popd
