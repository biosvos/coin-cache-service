#!/bin/bash

if [ $# -ne 1 ]; then
    echo "run with project name"
    exit
fi

replace_fmt="s/go-template/$1/g"

find . -type f -not -path "./.git/*" | xargs -i sed -i ${replace_fmt} {}

mv cmd/go-template cmd/$1
