#!/bin/bash

if [ "a$1" = "a" ];then
	echo "必要的参数:"
	echo "	sh start.sh -name game@127.0.0.1 -cookie 6d27544c07937e4a7fab8123291cc4df"
fi

ulimit -c  unlimited
export GOTRACEBACK=crash
sh -c "./bin/main -WriteLogStd false $* 2>main.core &"

