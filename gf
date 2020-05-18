#!/bin/bash
eval `go build -work -a 2>&1` && find $WORK -type f -name "*.a" | xargs -I{} du -hxs "{}" | sed -e s:${WORK}/::g
