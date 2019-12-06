#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
wget --no-check-certificate -O "${DIR}/MDID.zip" https://www.sz.tsinghua.edu.cn/labs/vipl/download/MDID.zip
unzip "${DIR}/MDID.zip" -d "${DIR}"
