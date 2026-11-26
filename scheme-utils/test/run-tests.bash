#!/usr/bin/env bash

for tc in *.sps; do
    larceny -I .. $tc
done

