#!/usr/bin/env bash

for tc in *.sps; do
    echo "Testing $tc"
    larceny -I .. $tc
done

