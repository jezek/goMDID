Recreates the metrics of [MDID dataset](https://www.sz.tsinghua.edu.cn/labs/vipl/mdid.html) using Go language. 

Why?
====

To be able evaluate Image Quality Assessment (IQA) in Go language (and an exercise for me).

Requirements:
============

- [Go language](https://golang.org/) installed.
- 1.8GB free space for download and extract dataset. After extracting the zip file can be deleted, unpacked dataset is ~1GB.

Installation & run:
===================

1) clone repository: ```git clone https://github.com/goMDID```
2) enter repository: ```cd goMDID```
3) load & extract dataset: ```./dataset/getMDID.sh```
4) run project: ```go run *.go```

Go third party dependencies:
===========================
- golang.org/x/image/bmp
-	github.com/dgryski/go-onlinestats *
-	github.com/mcgrew/gostats *
-	gonum.org/v1/gonum/stat *

*) can be omitted, currently only used for comparison/verifying of self-implemented correlation methods.

### Sources
1. [W. Sun, F. Zhou, Q. M. Liao. MDID: a multiply distorted image database for image quality assessment, Pattern Recognit. 61C (2017) pp. 153-168.](https://www.sz.tsinghua.edu.cn/labs/vipl/mdid.html)
